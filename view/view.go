package view

import (
	"encoding/json"
	"fmt"
	"github.com/russross/blackfriday"
	"html"
	"html/template"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"
	"database/sql"
	_ "github.com/lib/pq"
	"log"
)

type Page struct {
	Title       string
	Body        string
	DisplayBody template.HTML
	Path        template.HTML
	Information string
}

type User struct {
	Username       string
	Password       string
	Approvalstatus string
	Adminstatus    string
}

type Version map[string]*Page

var startTemplates = template.Must(template.ParseFiles("./static/templates/main.html", "./static/templates/head.html", "./static/templates/menu.html", "./static/templates/content_start.html", "./static/templates/footer.html"))

var name string

func dbOpen () *sql.DB {
	db, err := sql.Open("postgres", "user=postgres dbname=wiki sslmode=disable")
	if err != nil {
		log.Fatal(err)
		}
	return db
}


func startPage(w http.ResponseWriter) {

	articles := make([]string, 0)

	db := dbOpen()

	rows, err := db.Query("SELECT title FROM pages")
		if err != nil {
			log.Fatal(err)
		}

		defer rows.Close()
		for rows.Next() {
			err := rows.Scan(&name)
			if err != nil {
					log.Fatal(err)
			}
			articles = append(articles, name)
	    }
        err = rows.Err()
        if err != nil {
			log.Fatal(err)
        }

	//We use an anonymous struct for rendering:
	data := struct {
		Title       string
		MenuEntries [][2]string
		Articles    []string
	}{
		"start",
		[][2]string{
			{"Home", "/view/start"},
			{"Users", "/users"},
			{"Files", "/files"},
			{"Logout", "/logout"},
		},
		articles,
	}
	startTemplates.ExecuteTemplate(w, "main", &data)
}

func loadPage(title string) (*Page, error) {
	//Reading an article from disk:
	filename := "./articles/" + title
	file, err := os.Open(filename)
	//Map of versions expected in the article:
	versions := make(map[string]*Page)
	defer file.Close()
	//Forwarding errors to caller:
	if err != nil {
		return nil, err
	}
	//Decoding file contents:
	decoder := json.NewDecoder(file)
	decoder.Decode(&versions)
	//Finding latest version:
	var latest string
	for k := range versions {
		if k > latest {
			latest = k
		}
	}
	return versions[latest], nil
}

var pageTemplates = template.Must(template.ParseFiles("./static/templates/main.html", "./static/templates/head.html", "./static/templates/menu_view.html", "./static/templates/content_view.html", "./static/templates/footer.html"))

func renderPage(w http.ResponseWriter, p *Page) {
	data := struct {
		Title       string
		DisplayBody template.HTML
		Information string
		MenuEntries [][2]string
	}{
		p.Title,
		p.DisplayBody,
		p.Information,
		[][2]string{
			{"Home", "/view/start"},
			{"Edit this Page!", "/edit/" + p.Title},
			{"Users", "/users"},
			{"Files", "/files"},
		},
	}
	pageTemplates.ExecuteTemplate(w, "main", &data)
}

func ViewHandler(w http.ResponseWriter, r *http.Request, title string) {
	if title == "start" {
		startPage(w)
	} else {
		p, err := loadPage(title)
		if err != nil {
			http.Redirect(w, r, "/edit/"+title, http.StatusFound)
			return
		}
		renderPage(w, p)
	}
}

var editTemplates = template.Must(template.ParseFiles("./static/templates/main_edit.html", "./static/templates/head.html", "./static/templates/menu_edit.html", "./static/templates/content_view.html", "./static/templates/footer.html"))

func EditHandler(w http.ResponseWriter, r *http.Request, title string) {
	p, err := loadPage(title)
	if err != nil {
		p = &Page{Title: title}
	}
	editTemplates.ExecuteTemplate(w, "main", p)
}

func SaveHandler(w http.ResponseWriter, r *http.Request, title string) {
	body := r.FormValue("body")

	dBody := template.HTML(blackfriday.MarkdownBasic([]byte(html.EscapeString(body))))
	const layout = "2006-01-02 15:04:05"
	t := time.Now()
	cookie, err := r.Cookie("User")
	if err != nil {
		// just do something
	}
	author := cookie.Value
	information := "Aktualisiert:" + t.Format(layout) + " by " + string(author)
	p := &Page{Title: title, Body: body, DisplayBody: dBody, Information: information}
	versions := make(map[string]*Page)
	filename := "./articles/" + p.Title
	file, err := os.Open(filename)
	if err == nil {
		defer file.Close()
		decoder := json.NewDecoder(file)
		decoder.Decode(&versions)

	}
	versions[t.Format(layout)] = p
	out, err := json.Marshal(versions)
	if err != nil {
		fmt.Println("nicht jsoned")
	}

	ioutil.WriteFile(filename, out, 0600)
	http.Redirect(w, r, "/view/"+title, http.StatusFound)
}

func HandlerToHandleFunc(handler http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		handler.ServeHTTP(w, r)
	}
}

var versionTemplates = template.Must(template.ParseFiles("./static/templates/main.html", "./static/templates/head.html", "./static/templates/menu.html", "./static/templates/content_version.html", "./static/templates/footer.html"))
var specificVersionTemplates = template.Must(template.ParseFiles("./static/templates/main_specificVersion.html", "./static/templates/head.html", "./static/templates/content_view.html", "./static/templates/footer.html"))

func VersionHandler(w http.ResponseWriter, r *http.Request, title string, version *string) {
	filename := "./articles/" + title
	file, err := os.Open(filename)
	if err != nil {
		return
	}
	versions := make(map[string]*Page)
	defer file.Close()

	decoder := json.NewDecoder(file)
	decoder.Decode(&versions)

	if version == nil {
		keys := make([]string, 0, len(versions))
		for k := range versions {
			keys = append(keys, k)

		}
		sort.Strings(keys)
		reverseKeys := make([]string, len(keys))
		for i, k := range keys {
			reverseKeys[len(keys)-i-1] = k
		}
		data := struct {
			Title       string
			Versions    []string
			MenuEntries [][2]string
		}{
			title,
			reverseKeys,
			[][2]string{
				{"Home", "/view/start"},
				{title, "/view/" + title},
				{"Files", "/files"},
			},
		}
		versionTemplates.ExecuteTemplate(w, "main", &data)
	} else {
		specificVersionTemplates.ExecuteTemplate(w, "main", versions[*version])
	}

}

var userTemplates = template.Must(template.ParseFiles("./static/templates/main.html", "./static/templates/head.html", "./static/templates/menu.html", "./static/templates/content_users.html", "./static/templates/footer.html"))

func UserHandler(w http.ResponseWriter, r *http.Request, user string) {
	cookie, err := r.Cookie("User")
	if err != nil {
		return
	}

	username := cookie.Value
	file, err := os.Open("./users/" + username)
	if err != nil {
		fmt.Printf("Could not open file for user: %s\n", username)
	}
	defer file.Close()
	decoder := json.NewDecoder(file)
	var visitor User
	err = decoder.Decode(&visitor)
	if err != nil {
		fmt.Printf("Could not decode json for user: %s\n", username)
	}
	//Reading in the userlist:
	//userfiles, _ := ioutil.ReadDir("./users")

	userlist := make([]User, 0)

	db := dbOpen()

	rows, err := db.Query("SELECT name, password, approved, admin FROM users ORDER BY name ASC")
	if err != nil {
		log.Fatal(err)
	}

	defer rows.Close()
	var password, approved, admin string
	for rows.Next() {
		err := rows.Scan(&name, &password, &approved, &admin)
		if err != nil {
				log.Fatal(err)
		}
		user := User{name, password, approved, admin}
		userlist = append(userlist, user)
	}
    err = rows.Err()
    if err != nil {
		log.Fatal(err)
    }

	//Composing data to render, and rendering:
	data := struct {
		Title       string
		MenuEntries [][2]string
		ShowAdmin   bool
		Users       []User
	}{
		"Overview of users",
		[][2]string{
			{"Home", "/view/start"},
			{"Files", "/files"},
			{"Delete your own Account", "/remove/"+username},
		},
		visitor.Adminstatus == "admin",
		userlist,
	}
	userTemplates.ExecuteTemplate(w, "main", &data)
}

func ChangeAdminstatus(w http.ResponseWriter, r *http.Request, user string) {
	cookie, err := r.Cookie("User")
	if err != nil {
		return
	}
	username := cookie.Value

	db := dbOpen()

	var visitorStatus string
	var specificUserStatus string

	rows, err := db.Query("SELECT admin FROM users WHERE name = $1", username)
	defer rows.Close()
	for rows.Next() {
		err = rows.Scan(&visitorStatus)
		if err != nil {
			log.Fatal(err)
		}
	}

	rows, err = db.Query("SELECT admin FROM users WHERE name = $1", user)
	defer rows.Close()
	for rows.Next() {
		err = rows.Scan(&specificUserStatus)
		if err != nil {
			log.Fatal(err)
		}
	}

	if visitorStatus == "admin" {
		if specificUserStatus == "admin" {
			rows, _ := db.Query("UPDATE users SET admin = $1 WHERE name = $2", "user", user)
			defer rows.Close()
		} else {
			rows, _ := db.Query("UPDATE users SET admin = $1 WHERE name = $2", "admin", user)
			defer rows.Close()
		}
	}
	http.Redirect(w, r, "/users/", http.StatusFound)
}

func ChangeApprovalstatus(w http.ResponseWriter, r *http.Request, user string) {
	cookie, err := r.Cookie("User")
	if err != nil {
		return
	}
	username := cookie.Value

	db := dbOpen()

	var visitorStatus string
	var specificUserStatus string

	rows, err := db.Query("SELECT admin FROM users WHERE name = $1", username)
	defer rows.Close()
	for rows.Next() {
		err = rows.Scan(&visitorStatus)
		if err != nil {
			log.Fatal(err)
		}
	}

	rows, err = db.Query("SELECT approved FROM users WHERE name = $1", user)
	defer rows.Close()
	for rows.Next() {
		err = rows.Scan(&specificUserStatus)
		if err != nil {
			log.Fatal(err)
		}
	}

	if visitorStatus == "admin" {
		if specificUserStatus == "approved" {
			rows, _ := db.Query("UPDATE users SET approved = $1 WHERE name = $2", "not approved", user)
			defer rows.Close()
		} else {
			rows, _ := db.Query("UPDATE users SET approved = $1 WHERE name = $2", "approved", user)
			defer rows.Close()
		}
	}
	http.Redirect(w, r, "/users/", http.StatusFound)
}

func DeleteUser(w http.ResponseWriter, r *http.Request, user string) {
	var visitor User
	cookie, err := r.Cookie("User")
	if err != nil {
		return
	}
	username := cookie.Value
	visitorfile, err := os.Open("./users/" + username)
	if err != nil {
		println("error opening the file")
	}
	defer visitorfile.Close()
	visitordecoder := json.NewDecoder(visitorfile)
	err = visitordecoder.Decode(&visitor)
	if user == username || visitor.Adminstatus == "admin"{
		err := os.Remove("./users/" + user)
		if err != nil {
			println("error opening the file")
		}
	
	}
	http.Redirect(w, r, "/users/", http.StatusFound)
}

var videoPath = regexp.MustCompile("^/(files)/(.+)[.](mkv|avi|webm|mp4|mpg|mpeg|wmv|ogg|mp3|flac)")
var subPath = regexp.MustCompile("^/files/(.*)$")
var videoFile = regexp.MustCompile("^([^\\.][^/]*)\\.(mkv|avi|webm|mp4|mpg|mpeg|wmv|ogg|mp3|flac)$")

var fileTemplates = template.Must(template.ParseFiles("./static/templates/main.html", "./static/templates/head.html", "./static/templates/menu_files.html", "./static/templates/content_files.html", "./static/templates/footer.html"))

func FileHandler(w http.ResponseWriter, r *http.Request, path string) {
	matches := subPath.FindStringSubmatch(path)
	if matches == nil || len(matches) < 2 {
		fmt.Printf("Unexpected path in view.go:FileHandler: %s\n", path)
	} else { // We've got a valid path:
		pathParts := strings.Split(matches[1], "/")
		//Building path for the menu:
		type pathEntry struct {
			Href, Text string
			Active     bool
		}
		menuPath := make([]pathEntry, 0, len(pathParts))
		helper := "/files"
		for _, p := range pathParts {
			helper += "/" + p
			if p == "" {
				continue
			}
			menuPath = append(menuPath, pathEntry{helper, p, false})
		}
		if len(menuPath) > 0 {
			menuPath[len(menuPath)-1].Active = true
		}
		//Reading directory data:
		replacer := strings.NewReplacer("/files", "./data/fileserver")
		localPath := replacer.Replace(path)
		fInfos, err := ioutil.ReadDir(localPath)
		type dirData struct{ Path, Name string }
		var dirs []dirData
		if err == nil {
			dirs = make([]dirData, 0, len(fInfos))
			for _, f := range fInfos {
				//Filtering hidden files:
				if len(f.Name()) > 0 && f.Name()[0] == '.' {
					continue
				}
				//Adding to listed files:
				dirs = append(dirs, dirData{path, f.Name()})
			}
		}
		//Case we don't have a dir:
		type videoData struct{ Source, Subpath string }
		var video *videoData
		if dirs == nil {
			//Have we got a video file?
			vFile, err := os.Stat(localPath)
			if err == nil && videoFile.MatchString(vFile.Name()) {
				video = &videoData{localPath[1:], ""}
				vttPath := strings.Replace(localPath, filepath.Ext(localPath), ".vtt", 1)
				_, err := os.Stat(vttPath)
				if err == nil {
					video.Subpath = vttPath[1:]
				}
			}
		}
		//Composing data to render, and rendering:
		data := struct {
			Title       string
			MenuEntries [][2]string
			Path        []pathEntry
			Directory   []dirData
			Video       *videoData
		}{
			"Files",
			[][2]string{
				{"Home", "/view/start"},
				{"Users", "/users"},
				{"Files", "/files"},
			},
			menuPath,
			dirs,
			video,
		}
		fileTemplates.ExecuteTemplate(w, "main", &data)
		return
	}
}
