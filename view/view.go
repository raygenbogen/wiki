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

var templates = template.Must(template.ParseFiles("./static/files.html", "./static/adminpage.html", "./static/upload.html", "./static/version.html", "./static/specificversion.html", "./static/users.html"))
var validPath = regexp.MustCompile("^/(edit|save|view|vers|users)/([a-zA-Z0-9]+)$")
var versPath = regexp.MustCompile("^/(vers)/([a-zA-Z0-9]+)/(.+)$")
var userPath = regexp.MustCompile("^/(users)(/)?$")
var filePath = regexp.MustCompile("^/(files)/(?)")
var videoPath = regexp.MustCompile("^/(files)/(.+)[.](mkv|avi|webm|mp4|mpg|mpeg|wmv|ogg|mp3|flac)")
var approvalPath = regexp.MustCompile("^/(approve|disapprove)/(.+)")
var adminPath = regexp.MustCompile("^/(makeAdmin)/(.+)")

type Page struct {
	Title       string
	Body        string
	DisplayBody template.HTML
	Path        template.HTML
	Information string
}
type Version map[string]*Page

func MakeHandler(fn func(http.ResponseWriter, *http.Request, string)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		m := validPath.FindStringSubmatch(r.URL.Path)
		if m == nil {
			title := "start"
			http.Redirect(w, r, "/view/"+title, http.StatusFound)

		} else {
			fn(w, r, m[2])
		}
	}
}

func MakeVersionHandler(fn func(http.ResponseWriter, *http.Request, string, *string)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		println(r.URL.Path)
		m := versPath.FindStringSubmatch(r.URL.Path)
		println(m)
		if m == nil {
			println(m)
			n := validPath.FindStringSubmatch(r.URL.Path)
			println(n)
			if n == nil {
				println(n)
				http.Redirect(w, r, "/view/start", http.StatusFound)
				return
			}
			// list all versions of an articles
			fn(w, r, n[2], nil)
		} else {
			// show specific version
			fn(w, r, m[2], &m[3])
		}
	}
}

func MakeUserHandler(fn func(http.ResponseWriter, *http.Request, string)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		m := validPath.FindStringSubmatch(r.URL.Path)
		println(r.URL.Path)
		if m == nil {
			println("not a valid path")
			n := userPath.FindStringSubmatch(r.URL.Path)
			if n == nil {
				println("not even a users path")
				println(n)
				http.Redirect(w, r, "/users/", http.StatusFound)
				return
			}
			fn(w, r, "")
		} else {
			fn(w, r, m[2])

		}

	}
}

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
	fmt.Println("entering viewhandler now")
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

func MakeRedirectHandler(target string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, target, http.StatusFound)
		return
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

	//v := Version(versions)

	ioutil.WriteFile(filename, out, 0600)
	http.Redirect(w, r, "/view/"+title, http.StatusFound)
}

func HandlerToHandleFunc(handler http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		handler.ServeHTTP(w, r)
	}
}

func VersionHandler(w http.ResponseWriter, r *http.Request, title string, version *string) {
	var dBody template.HTML
	filename := "./articles/" + title
	file, err := os.Open(filename)
	if err != nil {
		return
	}
	versions := make(map[string]*Page)
	defer file.Close()

	decoder := json.NewDecoder(file)
	decoder.Decode(&versions)
	//var keys string

	if version == nil {
		var olderVersions string
		keys := make([]string, 0, len(versions))
		for k := range versions {
			keys = append(keys, k)

		}
		sort.Strings(keys)
		for k := range keys {
			HTMLAttr := "<li><a href=\"/vers/" + title + "/" + keys[k] + "\" target=\"versions\">" + keys[k] + "</a></li>"
			println(keys[k])
			olderVersions = HTMLAttr + olderVersions
		}
		dBody += template.HTML(olderVersions)
		renderTemplate(w, "version", &Page{Title: title, DisplayBody: dBody})
	} else {
		println(version)
		page := versions[*version]
		dBody := page.DisplayBody
		renderTemplate(w, "specificversion", &Page{Title: title, DisplayBody: dBody})
	}

}

func UserHandler(w http.ResponseWriter, r *http.Request, user string) {
	var dBody template.HTML

	cookie, err := r.Cookie("User")
	if err != nil {
		return
	}
	fmt.Println(cookie.Name)
	username := cookie.Value
	file, err := os.OpenFile("./users/users", os.O_RDWR|os.O_CREATE, 0600)
	if err != nil {
		println("error opening the file")
	}
	defer file.Close()
	decoder := json.NewDecoder(file)
	users := make(map[string]string)
	err = decoder.Decode(&users)
	adminfile, err := os.OpenFile("./users/admins", os.O_RDWR|os.O_CREATE, 0600)
	if err != nil {
		println("error opening the file")
	}
	defer adminfile.Close()
	admindecoder := json.NewDecoder(adminfile)
	admins := make(map[string]string)
	err = admindecoder.Decode(&admins)
	adminstatus := admins[username]
	println(adminstatus)
	var userlist string
	if adminstatus == "IsAdmin" {
		approvalfile, err := os.OpenFile("./users/approvedusers", os.O_RDWR|os.O_CREATE, 0600)
		if err != nil {
			println("error opening the file")
		}
		defer approvalfile.Close()
		approvaldecoder := json.NewDecoder(approvalfile)
		approvedusers := make(map[string]string)
		err = approvaldecoder.Decode(&approvedusers)
		keys := make([]string, 0, len(users))
		for k := range users {
			keys = append(keys, k)

		}
		sort.Strings(keys)
		for k := range keys {
			approvalstatus := approvedusers[keys[k]]
			if approvalstatus != "approved" {
				adminstatus = admins[keys[k]]
				println(adminstatus)
				HTMLAttr := "<tr><td>" + keys[k] + "</td><td><form action=\"/approve/" + keys[k] + "\" method=\"post\"><button type=\"submit\" class=\"btn btn-primary\">" + approvalstatus + "</button></form></td><td><form action=\"/makeAdmin/" + keys[k] + "\" method=\"post\"><button type=\"submit\" class=\"btn btn-primary\">" + adminstatus + "</button></form></td></tr>"
				userlist = userlist + HTMLAttr
			} else {
				adminstatus = admins[keys[k]]
				println(adminstatus)
				HTMLAttr := "<tr><td>" + keys[k] + "</td><td><form action=\"/disapprove/" + keys[k] + "\" method=\"post\"><button type=\"submit\" class=\"btn btn-primary\">" + approvalstatus + "</button></form></td><td><form action=\"/makeAdmin/" + keys[k] + "\" method=\"post\"><button type=\"submit\" class=\"btn btn-primary\">" + adminstatus + "</button></form></td></tr>"
				userlist = userlist + HTMLAttr
			}

			println(keys[k])

		}
		dBody += template.HTML(userlist)
		renderTemplate(w, "adminpage", &Page{DisplayBody: dBody})
	} else {
		keys := make([]string, 0, len(users))
		for k := range users {
			keys = append(keys, k)

		}
		sort.Strings(keys)
		for k := range keys {

			HTMLAttr := "<tr><td>" + keys[k] + "</td></tr>"

			println(keys[k])
			userlist = userlist + HTMLAttr
		}
		dBody += template.HTML(userlist)
		renderTemplate(w, "users", &Page{DisplayBody: dBody})
	}

	//renderTemplate(w, "users", &Page{DisplayBody: dBody})
}

func MakeApprovalHandler(fn func(http.ResponseWriter, *http.Request, string)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		m := approvalPath.FindStringSubmatch(r.URL.Path)
		if m == nil {
			return
		}
		fn(w, r, m[2])
	}
}

func MakeAdminHandler(fn func(http.ResponseWriter, *http.Request, string)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		m := adminPath.FindStringSubmatch(r.URL.Path)
		if m == nil {
			return
		}
		fn(w, r, m[2])
	}
}

func MakeAdmin(w http.ResponseWriter, r *http.Request, user string) {
	cookie, err := r.Cookie("User")
	if err != nil {
		return
	}
	println("ready to make some admins")
	username := cookie.Value
	adminfile, err := os.OpenFile("./users/admins", os.O_RDWR|os.O_CREATE, 0600)
	if err != nil {
		println("error opening the file")
	}
	defer adminfile.Close()
	admindecoder := json.NewDecoder(adminfile)
	admins := make(map[string]string)
	err = admindecoder.Decode(&admins)
	adminstatus := admins[username]
	if adminstatus == "IsAdmin" {
		if admins[user] == "IsAdmin" {
			admins[user] = "IsNoTAdmin"
		} else {
			println(admins[user])
			admins[user] = "IsAdmin"
			println(admins[user])
		}
		jsonedadmins, _ := json.Marshal(admins)
		ioutil.WriteFile("./users/admins", jsonedadmins, 0600)
	}
	//fmt.Fprintf(w,"approved")
	http.Redirect(w, r, "/users/", http.StatusFound)
}

func ApproveUser(w http.ResponseWriter, r *http.Request, user string) {
	println("der user ist:" + user)
	cookie, err := r.Cookie("User")
	if err != nil {
		return
	}
	fmt.Println(cookie.Name)
	username := cookie.Value
	adminfile, err := os.OpenFile("./users/admins", os.O_RDWR|os.O_CREATE, 0600)
	if err != nil {
		println("error opening the file")
	}
	defer adminfile.Close()
	admindecoder := json.NewDecoder(adminfile)
	admins := make(map[string]string)
	err = admindecoder.Decode(&admins)
	adminstatus := admins[username]
	println(adminstatus)
	if adminstatus == "IsAdmin" {
		approvalfile, err := os.OpenFile("./users/approvedusers", os.O_RDWR|os.O_CREATE, 0600)
		if err != nil {
			println("error opening the file")
		}
		defer approvalfile.Close()
		approvaldecoder := json.NewDecoder(approvalfile)
		approvedusers := make(map[string]string)
		err = approvaldecoder.Decode(&approvedusers)
		println(approvedusers)
		approvedusers[user] = "approved"
		approvalstatus := approvedusers[user]
		println("der neue status von" + user + " ist:" + approvalstatus)
		jsonedapprovals, err := json.Marshal(approvedusers)
		ioutil.WriteFile("./users/approvedusers", jsonedapprovals, 0600)
	}
	//fmt.Fprintf(w,"approved")
	http.Redirect(w, r, "/users/", http.StatusFound)
}

func MakeDisApprovalHandler(fn func(http.ResponseWriter, *http.Request, string)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		m := approvalPath.FindStringSubmatch(r.URL.Path)
		if m == nil {
			return
		}
		fn(w, r, m[2])
	}
}

func DisApproveUser(w http.ResponseWriter, r *http.Request, user string) {
	println("der user ist:" + user)
	cookie, err := r.Cookie("User")
	if err != nil {
		return
	}
	fmt.Println(cookie.Name)
	username := cookie.Value
	adminfile, err := os.OpenFile("./users/admins", os.O_RDWR|os.O_CREATE, 0600)
	if err != nil {
		println("error opening the file")
	}
	defer adminfile.Close()
	admindecoder := json.NewDecoder(adminfile)
	admins := make(map[string]string)
	err = admindecoder.Decode(&admins)
	adminstatus := admins[username]
	println(adminstatus)
	if adminstatus == "IsAdmin" {
		approvalfile, err := os.OpenFile("./users/approvedusers", os.O_RDWR|os.O_CREATE, 0600)
		if err != nil {
			println("error opening the file")
		}
		defer approvalfile.Close()
		approvaldecoder := json.NewDecoder(approvalfile)
		approvedusers := make(map[string]string)
		err = approvaldecoder.Decode(&approvedusers)
		println(approvedusers)
		approvedusers[user] = "NotYetApproved"
		approvalstatus := approvedusers[user]
		println("der neue status von" + user + " ist:" + approvalstatus)
		jsonedapprovals, err := json.Marshal(approvedusers)
		ioutil.WriteFile("./users/approvedusers", jsonedapprovals, 0600)
	}
	//fmt.Fprintf(w,"approved")
	http.Redirect(w, r, "/users/", http.StatusFound)
}

func MakeFileHandler(fn func(http.ResponseWriter, *http.Request, string)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		m := filePath.FindStringSubmatch(r.URL.Path)
		if m == nil {
			println("not a valid path")
			return
		}

		fn(w, r, r.URL.Path)

	}
}

func FileHandler(w http.ResponseWriter, r *http.Request, path string) {
	replacer := strings.NewReplacer("files", "data/fileserver")
	println(path)
	newpath := replacer.Replace(path)
	subpath := newpath
	var coolnewpath template.HTML
	_, currentdirectory := filepath.Split(path)
	if path != "/files/" {
		coolnewpath = template.HTML("<li class=\"active\"><a href=\"" + currentdirectory + "\">" + currentdirectory + "</a></li>")
	}
	for i := strings.Count(subpath, "/") - 3; i > 0; i-- {
		basepath := filepath.Dir(subpath)
		println(basepath)
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

		HTMLAttr := "<tr><td><video width=\"100%\" height=\"80%\" preload=\"auto\" controls><source src=\"" + newpath + "\" type=video/webm /><track src=" + subpath + " kind=\"subtitle\" src=\"de-DE\" label=\"german\"/>Your browser does not support the video tag.</video></td></tr>"
		dBody = template.HTML(HTMLAttr)
		renderTemplate(w, "files", &Page{Title: title, DisplayBody: dBody, Path: coolnewpath})
	}

}
