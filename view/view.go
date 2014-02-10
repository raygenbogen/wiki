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
)

var templates = template.Must(template.ParseFiles("./static/edit.html", "./static/view.html", "./static/upload.html"))
var validPath = regexp.MustCompile("^/(edit|save|view)/([a-zA-Z0-9]+)$")

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

func loadPage(title string) (*Page, error) {
	var body string

	var dBody template.HTML
	var information string

	if title == "start" {
		files, _ := ioutil.ReadDir("./articles/")

		dBody = "Welcome! Have a look at the existing pages below or create a new one.<br><br>"
		for _, f := range files {
			HTMLAttr := "<li><a href= /view/" + f.Name() + ">" + f.Name() + "</a></li>"
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
	information := "Aktualisiert:" + t.Format(layout)
	versions := make(map[string]*Page)

	p := &Page{Title: title, Body: body, DisplayBody: dBody, Information: information}
	versions[t.Format(layout)] = p
	//v := Version(versions)

	filename := "./articles/" + p.Title
	out, err := json.Marshal(versions)
	if err != nil {
		fmt.Println("nicht jsoned")
	}
	ioutil.WriteFile(filename, out, 0600)
	http.Redirect(w, r, "/view/"+title, http.StatusFound)
}
