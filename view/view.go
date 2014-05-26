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
	"regexp"
	"sort"
	"time"
)

var templates = template.Must(template.ParseFiles("./static/startpage.html", "./static/adminpage.html", "./static/edit.html", "./static/view.html", "./static/upload.html", "./static/version.html", "./static/specificversion.html", "./static/users.html"))
var validPath = regexp.MustCompile("^/(edit|save|view|vers|users)/([a-zA-Z0-9]+)$")
var versPath = regexp.MustCompile("^/(vers)/([a-zA-Z0-9]+)/(.+)$")
var userPath = regexp.MustCompile("^/(users)(/)?$")

type Page struct {
	Title       string
	Body        string
	DisplayBody template.HTML
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

func loadPage(title string) (*Page, error) {

	var body string

	var dBody template.HTML
	var information string
	println("last updated")
	println(information)

	if title == "start" {
		files, _ := ioutil.ReadDir("./articles/")
		body = "Welcome! Have a look at the existing pages below or create a new one."
		for _, f := range files {
			HTMLAttr := "<tr><td><a href=\"/view/" + f.Name() + "\">" + f.Name() + "</a></td><td><a href=\"/vers/" + f.Name() + "\">" + f.Name() + "</a></td></tr>"
			dBody += template.HTML(HTMLAttr)
		}

	} else {
		var latestVersion string
		filename := "./articles/" + title
		file, err := os.Open(filename)
		versions := make(map[string]*Page)
		defer file.Close()
		if err != nil {
			return nil, err
		}
		decoder := json.NewDecoder(file)
		decoder.Decode(&versions)
		var keys string
		for k := range versions {
			if k > keys {
				keys = k
			}
		}
		//sort.Strings(keys)

		latestVersion = keys
		page := versions[latestVersion]
		return page, nil

	}

	return &Page{Title: title, Body: body, DisplayBody: dBody, Information: information}, nil
}

func ViewHandler(w http.ResponseWriter, r *http.Request, title string) {
	fmt.Println("entering viewhandler now")
	p, err := loadPage(title)
	if err != nil {
		http.Redirect(w, r, "/edit/"+title, http.StatusFound)
		return
	}
	if title != "start" {
		renderTemplate(w, "view", p)
	} else {
		renderTemplate(w, "startpage", p)
	}
}

func MakeRedirectHandler(target string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, target, http.StatusFound)
		return
	}
}

func EditHandler(w http.ResponseWriter, r *http.Request, title string) {
	p, err := loadPage(title)
	if err != nil {
		p = &Page{Title: title}
	}

	renderTemplate(w, "edit", p)
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
			HTMLAttr := "<tr><td>" + keys[k] + "</td><td>" + approvalstatus + "</td></tr>"

			println(keys[k])
			userlist = HTMLAttr + userlist
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
