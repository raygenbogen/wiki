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
	"time"
	"sort"
)

var templates = template.Must(template.ParseFiles("./static/edit.html", "./static/view.html", "./static/upload.html", "./static/version.html", "./static/specificversion.html"))
var validPath = regexp.MustCompile("^/(edit|save|view|vers)/([a-zA-Z0-9]+)$")
var versPath = regexp.MustCompile("^/(vers)/([a-zA-Z0-9]+)/(.+)$")

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
		if m == nil {
			n := validPath.FindStringSubmatch(r.URL.Path)
			if n == nil {
				http.Redirect(w, r, "/view/start", http.StatusFound)
				println("weder valif noch irgendwas")
				return
			}
			println("versionsübersicht")
			fn(w, r, n[2], nil)
		} else {
			println("spezifische version")
			fn(w, r, m[2], &m[3])
		}
	}
}

func loadPage(title string) (*Page, error) {

	var body string

	var dBody template.HTML
	var information string

	if title == "start" {
		files, _ := ioutil.ReadDir("./articles/")

		dBody = "Welcome! Have a look at the existing pages below or create a new one.<br><br>"
		for _, f := range files {
			HTMLAttr := "<li><a href=\"/view/" + f.Name() + "\">" + f.Name() + "</a></li>"
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

	renderTemplate(w, "view", p)
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
		keys := make ([]string, 0, len(versions))
		for k := range versions {
			keys = append(keys,k)
			
			
		}
		sort.Strings(keys)
		for k := range keys {
			HTMLAttr := "<li><a href=\"/vers/"+ title +"/"+ keys[k] + "\" target=\"versions\">" + keys[k] + "</a></li>"
			println(keys[k])
			olderVersions = HTMLAttr + olderVersions
		}
		dBody += template.HTML(olderVersions)
		renderTemplate(w, "version", &Page{Title: title,DisplayBody: dBody})
	} else {
		println(version)
		page := versions[*version]
		dBody := page.DisplayBody
		renderTemplate(w, "specificversion", &Page{Title: title ,DisplayBody: dBody})
	}

}
