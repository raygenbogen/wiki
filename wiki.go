package main
import(
	"fmt"
	"io/ioutil"
	"net/http"
	"html/template"
	"regexp"
	"github.com/russross/blackfriday"
	"log"
	)

var templates = template.Must(template.ParseFiles("./static/edit.html", "./static/view.html"))
var validPath = regexp.MustCompile("^/(edit|save|view)/([a-zA-Z0-9]+)$")
var invalidTitle = regexp.MustCompile("^$")

type Page struct{
	Title string
	Body [] byte
	DisplayBody template.HTML
}

func (p *Page) save() error {
	filename := "./articles/" + p.Title
	return ioutil.WriteFile(filename, p.Body, 0600)
}

func loadPage (title string) (*Page, error) {
	var body []byte 
	var err error
	var dBody template.HTML
	if title == "start"{
		files, erro := ioutil.ReadDir("./articles/") 
		err = erro
		dBody = "Welcome! Have a look at the existing pages below or create a new one.<br><br>"
		for _, f := range files {
			HTMLAttr := "<li><a href= /view/"+f.Name() + ">" + f.Name() + "</a></li>"
			dBody += template.HTML(HTMLAttr)
		}
	}else{
		filename :="./articles/" + title
		body,err = ioutil.ReadFile(filename)
		dBody = template.HTML(blackfriday.MarkdownBasic(body))
	}
	if err != nil {
		return nil, err
	}
	return &Page{Title: title, Body: body , DisplayBody: dBody}, nil
}

func main (){
	//go http.ListenAndServeTLS(":10443", "./static/cert.pem","./static/key.pem", nil)
	//http.ListenAndServe(":8080", http.HandlerFunc(betterProto))
	/*if err := http.ListenAndServe(":8080", http.HandlerFunc(betterProto));
		err != nil {
		log.Fatalf("ListenAndServe error: %v", err)
	}*/
	
	http.HandleFunc("/view/", makeHandler(viewHandler))
	http.HandleFunc("/edit/", makeHandler(editHandler))
	http.HandleFunc("/save/", makeHandler(saveHandler))
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("./static/"))))
	http.HandleFunc("/", makeRedirectHandler("/view/start"))
	if err := http.ListenAndServeTLS(":10443", "./static/certificate.pem", "./static/key.pem", nil); err != nil {
		log.Fatalf("ListenAndServeTLS error: %v", err)
	}
	
	

	
}


/*
func betterProto(w http.ResponseWriter, req *http.Request) {
	
	println(req.URL.Path+"\n") 
	http.Redirect(w, req, "https://localhost:10443/"+req.URL.Path, http.StatusMovedPermanently)

}*/


func startHandler (w http.ResponseWriter, r *http.Request){

}
func handler (w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Hi there, I love %s!", r.URL.Path[1:])
}

func viewHandler (w http.ResponseWriter, r *http.Request, title string){
	p, err := loadPage(title)
	if err !=nil {
		http.Redirect(w, r, "/edit/"+title, http.StatusFound)
		return
	}
		
	
	renderTemplate(w, "view", p)
}

func makeRedirectHandler (target string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request){
		http.Redirect(w,r,target, http.StatusFound)
		return
	}
}

func editHandler ( w http.ResponseWriter, r *http.Request, title string){
	p, err := loadPage(title)
	if err != nil {
		p = &Page{Title:title}
	}

	renderTemplate(w, "edit", p)
}

func renderTemplate (w http.ResponseWriter, tmpl string, p *Page){
	err := templates.ExecuteTemplate(w, tmpl+".html", p)
	if err != nil{
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func saveHandler (w http.ResponseWriter, r *http.Request, title string){
	body := r.FormValue("body")
	p := &Page{Title: title, Body: []byte(body)}
	err := p.save()
	if err != nil{
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	http.Redirect( w, r, "/view/"+title, http.StatusFound)
}


	
func makeHandler(fn func(http.ResponseWriter, *http.Request, string)) http.HandlerFunc {
	return func (w http.ResponseWriter, r *http.Request){
		m := validPath.FindStringSubmatch(r.URL.Path)
        if m == nil {
        	title := "start"
        	http.Redirect(w,r,"/view/"+title,http.StatusFound)

		}else{
			fn(w, r, m[2])
		}
	}
}