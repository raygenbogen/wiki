package view

import (
	"encoding/json"
	"fmt"
	"github.com/russross/blackfriday"
	"io/ioutil"
	"net/http"
	"html"
	"html/template"
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

func (p *Page) save() error {
	filename := "./articles/" + p.Title
	out, err := json.Marshal(p)
	if err != nil {
		return err
	}
	return ioutil.WriteFile(filename, out, 0600)
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
		filename := "./articles/" + title
		bodie, err := ioutil.ReadFile(filename)
		if err != nil {
			return nil, err
		}

		var in map[string]interface{}
		json.Unmarshal(bodie, &in)
		fmt.Println(in)
		body = in["Body"].(string)
		dBody = template.HTML(in["DisplayBody"].(string))
		information = in["Information"].(string)

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
	//body = html.EscapeString(body) 
	dBody := template.HTML(blackfriday.MarkdownBasic([]byte(html.EscapeString(body))))
	const layout = "2006-01-02 15:04:05"
	t := time.Now()
	information := "Aktualisiert:" + t.Format(layout)
	p := &Page{Title: title, Body: body, DisplayBody: dBody, Information: information}
	err := p.save()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/view/"+title, http.StatusFound)
}
