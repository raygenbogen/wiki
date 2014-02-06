package auth

import (
	"html/template"
	"net/http"
	"encoding/json"
	"os"







)

type Users struct{
	Username string
	Password string
}
var templates = template.Must(template.ParseFiles("./static/auth.html", "./static/register.html"))

func Auth (w http.ResponseWriter, r *http.Request){
	renderAuth(w,"upload", nil)

}
func renderAuth (w http.ResponseWriter, tmpl string , data interface{}){
	templates.ExecuteTemplate(w, "auth.html", nil)
}

func RegisterHandler(w http.ResponseWriter,r *http.Request){
	switch r.Method {

	case "GET":
	renderRegister(w, "register.html", nil)
	case "POST":
	username := r.FormValue("username")
	password := r.FormValue("password")

	println(username)
	println(password)
	if len(password) < 8{
		println("more letters")
		renderRegister(w, "register.html", "More Letters.Do it.Do it Now!")
		return
	}
	users := &Users{username, password}
	jsonedusers, err :=json.Marshal(users)

	println(string(jsonedusers))
		if err != nil{
			println("did we jsoned the users yet?")

		}
	//some error handling
	f, err := os.OpenFile("./users/users", os.O_RDWR | os.O_CREATE, 0600)
	//some error handling
	if err != nil{
		println("where is this file everyone's talking about?")
		return

	}
	defer f.Close()
	_, err = f.Write(jsonedusers)
	if err != nil{
		println("Nope.Didn't write that shit!")
		return
	}
	f.Close()
	http.Redirect(w, r, "/auth/", http.StatusFound)
}

}
func Register( w http.ResponseWriter, r *http.Request){

}

func renderRegister (w http.ResponseWriter, tmpl string, data interface{}){
	templates.ExecuteTemplate(w, "register.html", data)
}