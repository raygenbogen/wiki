package main
import(
	"fmt"
	"io/ioutil"
	"net/http"
	"html/template"
	"regexp"
	)

var templates = template.Must(template.ParseFiles("edit.html", "view.html"))
var validPath = regexp.MustCompile("^/(edit|save|view)/([a-zA-Z0-9]+)$")
var invalidTitle = regexp.MustCompile("^$")

type Page struct{
	Title string
	Body [] byte
}

func (p *Page) save() error {
	filename := p.Title + ".txt"
	return ioutil.WriteFile(filename, p.Body, 0600)
}

func loadPage(title string) (*Page, error) {
	filename :=title + ".txt"
	body,err :=ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	return &Page{Title: title, Body: body}, nil
}

func main (){
	http.HandleFunc("/view/", makeHandler(viewHandler))
	http.HandleFunc("/edit/", makeHandler(editHandler))
	http.HandleFunc("/save/", makeHandler(saveHandler))
	//http.HandleFunc("/",makeHandler(redirectHandler))
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("./static/"))))
	http.ListenAndServe(":8080", nil)
}

func handler (w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Hi there, I love %s!", r.URL.Path[1:])
}

func viewHandler (w http.ResponseWriter, r *http.Request, title string){
	//figuring out the best way to forbid empty titles
	//m := invalidTitle.FindStringSubmatch(title)
        /*if len(title) == 0 {
        	title := "start"
        	http.Redirect(w,r,"/view/"+title,http.StatusFound)
            http.NotFound(w, r)
                        
		}*/
	
	p, err := loadPage(title)
	if err !=nil {
		http.Redirect(w, r, "/edit/"+title, http.StatusFound)
		return
	}
		
	
	renderTemplate(w, "view", p)
}
/*
func redirectHandler(w http.ResponseWriter, r *http.Request, title string){
	title = "start"
    http.Redirect(w,r,"/view/"+title,http.StatusFound)
    return
}*/

func editHandler( w http.ResponseWriter, r *http.Request, title string){
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

func saveHandler(w http.ResponseWriter, r *http.Request, title string){
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
        	//title := "start"
        	//http.Redirect(w,r,"/view/"+title,http.StatusFound)
            http.NotFound(w, r)
                        
		}
		
		fn(w, r, m[2])
	}
}