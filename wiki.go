package main

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/russross/blackfriday"
	"html/template"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"regexp"
	"time"
)

var templates = template.Must(template.ParseFiles("./static/edit.html", "./static/view.html", "./static/upload.html", "./static/fileserver.html"))
var validPath = regexp.MustCompile("^/(edit|save|view)/([a-zA-Z0-9]+)$")
var validType = regexp.MustCompile("^.*.(gif|jpeg|jpg)$")
var hashedTime string

type Page struct {
	Title       string
	Body        string
	DisplayBody template.HTML
	Information string
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

func main() {

	http.HandleFunc("/view/", makeHandler(viewHandler))
	http.HandleFunc("/edit/", makeHandler(editHandler))
	http.HandleFunc("/save/", makeHandler(saveHandler))
	//http.HandleFunc("/data/upload/", dataHandler)
	http.HandleFunc("/upload/", uploadHandler)
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("./static/"))))
	http.Handle("/data/", http.StripPrefix("/data/", http.FileServer(http.Dir("./data/"))))
	http.HandleFunc("/", makeRedirectHandler("/view/start"))
	if err := http.ListenAndServeTLS(":10443", "./static/certificate.pem", "./static/key.pem", nil); err != nil {
		log.Fatalf("ListenAndServeTLS error: %v", err)
	}

}

func startHandler(w http.ResponseWriter, r *http.Request) {

}
func handler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Hi there, I love %s!", r.URL.Path[1:])
}

func viewHandler(w http.ResponseWriter, r *http.Request, title string) {
	p, err := loadPage(title)
	if err != nil {
		http.Redirect(w, r, "/edit/"+title, http.StatusFound)
		return
	}

	renderTemplate(w, "view", p)
}

func makeRedirectHandler(target string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, target, http.StatusFound)
		return
	}
}

func editHandler(w http.ResponseWriter, r *http.Request, title string) {
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

func saveHandler(w http.ResponseWriter, r *http.Request, title string) {
	body := r.FormValue("body")
	dBody := template.HTML(blackfriday.MarkdownBasic([]byte(body)))
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

func renderUpload(w http.ResponseWriter, tmpl string, data interface{}) {
	templates.ExecuteTemplate(w, "upload.html", data)
}

func uploadHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {

	case "GET":
		renderUpload(w, "upload", nil)
		fmt.Println("getting stuff")

	case "POST":

		fmt.Println("saving files to disk")
		reader, err := r.MultipartReader()
		if err != nil {

			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		h := md5.New()
		const layout = "2006-01-02 15:04:05"
		t := time.Now()

		hashedTime = t.Format(layout)
		fmt.Println(hashedTime)
		hashedTime = hex.EncodeToString(h.Sum([]byte(hashedTime)))
		fmt.Println(hashedTime)
		os.Mkdir("data/upload/"+hashedTime+"/", 0700)
		
		for {
			fmt.Println("uploading parts of file")
			part, err := reader.NextPart()
			if err == io.EOF {
				break
			}

			if part.FileName() == "" {
				continue
			}
			m:=validType.FindStringSubmatch(part.FileName())
			if m == nil {
				fmt.Println("error much?")
				continue

			}
			fmt.Println("actually writing stuff")
			dst, err := os.Create("data/upload/" + hashedTime + "/" + part.FileName())
			defer dst.Close()

			if err != nil {
				fmt.Println("error at writing")
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			if _, err := io.Copy(dst, part); err != nil {
				fmt.Println("error while copying")
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		}
		
			
			http.Redirect(w, r, "/data/upload/"+hashedTime, http.StatusFound)

		
	}
}

func makeHandler(fn func(http.ResponseWriter, *http.Request, string)) http.HandlerFunc {
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
