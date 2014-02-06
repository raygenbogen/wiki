package auth

import (
	"html/template"
	"net/http"






)

var templates = template.Must(template.ParseFiles("./static/auth.html", "./static/register.html"))

func Auth (w http.ResponseWriter, r *http.Request){
	renderAuth(w,"upload", nil)

}
func renderAuth (w http.ResponseWriter, tmpl string , data interface{}){
	templates.ExecuteTemplate(w, "auth.html", nil)
}

func RegisterHandler(w http.ResponseWriter,r *http.Request){
	renderRegister(w, "register.html", nil)
}

func renderRegister (w http.ResponseWriter, tmpl string, data interface{}){
	templates.ExecuteTemplate(w, "register.html", nil)
}