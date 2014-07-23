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
)

var templates = template.Must(template.ParseFiles("./static/files.html", "./static/upload.html"))
var videoPath = regexp.MustCompile("^/(files)/(.+)[.](mkv|avi|webm|mp4|mpg|mpeg|wmv|ogg|mp3|flac)")

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

var startTemplates = template.Must(template.ParseFiles("./static/templates/main.html", "./static/templates/head.html", "./static/templates/menu.html", "./static/templates/content_start.html"))

func startPage(w http.ResponseWriter) {
	files, _ := ioutil.ReadDir("./articles/")
	//Note that the articles slice starts with len=0, but cap=len(files)
	var articles []string = make([]string, 0, len(files))
	for _, f := range files {
		name := f.Name()
		//We add only names that don't start with a '.':
		if name[0] != '.' {
			l := len(articles)
			articles = articles[0 : l+1]
			articles[l] = name
		}
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

var pageTemplates = template.Must(template.ParseFiles("./static/templates/main.html", "./static/templates/head.html", "./static/templates/menu_view.html", "./static/templates/content_view.html"))

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
			{"Users", "/Users"},
			{"Files", "/Files"},
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

var editTemplates = template.Must(template.ParseFiles("./static/templates/main_edit.html", "./static/templates/head.html", "./static/templates/menu_edit.html", "./static/templates/content_view.html"))

func EditHandler(w http.ResponseWriter, r *http.Request, title string) {
	p, err := loadPage(title)
	if err != nil {
		p = &Page{Title: title}
	}
	editTemplates.ExecuteTemplate(w, "main", p)
}

func renderTemplate(w http.ResponseWriter, tmpl string, p *Page) {
	err := templates.ExecuteTemplate(w, tmpl+".html", p)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
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

var versionTemplates = template.Must(template.ParseFiles("./static/templates/main.html", "./static/templates/head.html", "./static/templates/menu.html", "./static/templates/content_version.html"))
var specificVersionTemplates = template.Must(template.ParseFiles("./static/templates/main_specificVersion.html", "./static/templates/head.html", "./static/templates/content_view.html"))

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

var userTemplates = template.Must(template.ParseFiles("./static/templates/main.html", "./static/templates/head.html", "./static/templates/menu.html", "./static/templates/content_users.html"))

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
	userfiles, _ := ioutil.ReadDir("./users")
	userlist := make([]User, 0, len(userfiles))
	for _, u := range userfiles {
		if u.Name()[0] != '.' {
			userfile, err := os.Open("./users/" + u.Name())
			if err != nil {
				fmt.Printf("Error opening file for user: %s\n", u.Name())
			}
			var user User
			decoder = json.NewDecoder(userfile)
			err = decoder.Decode(&user)
			if err != nil {
				fmt.Printf("Error decoding JSON in user file: %s\n", u.Name())
			}
			userlist = append(userlist, user)
		}
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
		},
		visitor.Adminstatus == "admin",
		userlist,
	}
	userTemplates.ExecuteTemplate(w, "main", &data)
}

func ChangeAdminstatus(w http.ResponseWriter, r *http.Request, user string) {
	var visitor User
	var specificUser User
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
	userfile, err := os.Open("./users/" + user)
	if err != nil {
		println("error opening the file")
	}
	defer userfile.Close()
	specificdecoder := json.NewDecoder(userfile)
	err = specificdecoder.Decode(&specificUser)
	if visitor.Adminstatus == "admin" {
		if specificUser.Adminstatus == "admin" {
			specificUser.Adminstatus = "user"
		} else {
			specificUser.Adminstatus = "admin"
		}
		jsonedUser, _ := json.Marshal(specificUser)
		ioutil.WriteFile("./users/"+user, jsonedUser, 0600)
	}
	http.Redirect(w, r, "/users/", http.StatusFound)
}

func ChangeApprovalstatus(w http.ResponseWriter, r *http.Request, user string) {
	var visitor User
	var specificUser User
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
	userfile, err := os.Open("./users/" + user)
	if err != nil {
		println("error opening the file")
	}
	defer userfile.Close()
	specificdecoder := json.NewDecoder(userfile)
	err = specificdecoder.Decode(&specificUser)

	if visitor.Adminstatus == "admin" {
		if specificUser.Approvalstatus == "approved" {
			specificUser.Approvalstatus = "not approved"
		} else {
			specificUser.Approvalstatus = "approved"
		}
		jsonedUser, _ := json.Marshal(specificUser)
		ioutil.WriteFile("./users/"+user, jsonedUser, 0600)
	}
	http.Redirect(w, r, "/users/", http.StatusFound)
}

func FileHandler(w http.ResponseWriter, r *http.Request, path string) {
	replacer := strings.NewReplacer("files", "data/fileserver")
	newpath := replacer.Replace(path)
	subpath := newpath
	var coolnewpath template.HTML
	_, currentdirectory := filepath.Split(path)
	if path != "/files/" {
		coolnewpath = template.HTML("<li class=\"active\"><a href=\"" + currentdirectory + "\">" + currentdirectory + "</a></li>")
	}
	for i := strings.Count(subpath, "/") - 3; i > 0; i-- {
		basepath := filepath.Dir(subpath)

		replacer := strings.NewReplacer("data/fileserver", "files")
		basepath = replacer.Replace(basepath)
		subpath = basepath
		_, lastpartofsubpath := filepath.Split(basepath)

		HTMLAttr := "<li class \"active\"><a href=\"" + basepath + "\">" + lastpartofsubpath + "</a></li>"
		coolnewpath = template.HTML(HTMLAttr) + coolnewpath

	}
	var dBody template.HTML
	m := videoPath.FindStringSubmatch(path)
	if m == nil {

		files, _ := ioutil.ReadDir("." + newpath)
		title := currentdirectory
		for _, f := range files {

			HTMLAttr := "<tr><td><a href=\"" + path + "/" + f.Name() + "\">" + f.Name() + "</a></td></tr>"
			dBody += template.HTML(HTMLAttr)
		}
		renderTemplate(w, "files", &Page{Title: title, DisplayBody: dBody, Path: coolnewpath})
	} else {
		title := currentdirectory
		replace := strings.NewReplacer("webm", "vtt")
		subpath := replace.Replace(newpath)

		HTMLAttr := "<tr><td align=\"center\" valign=\"middle\"><video width=\"1280\" height=\"720\" preload=\"auto\" controls><source src=\"" + newpath + "\" type=video/webm /><track src=" + subpath + " kind=\"subtitle\" src=\"de-DE\" label=\"german\"/>Your browser does not support the video tag.</video></td></tr>"
		dBody = template.HTML(HTMLAttr)
		renderTemplate(w, "files", &Page{Title: title, DisplayBody: dBody, Path: coolnewpath})
	}

}
