package auth

import (
	"html/template"
	"net/http"
	"encoding/json"
	"os"
	"io/ioutil"
	"code.google.com/p/go.crypto/bcrypt"







)

type Users struct{
	Username string
	Password string
}
var templates = template.Must(template.ParseFiles("./static/auth.html", "./static/register.html"))

func Auth (w http.ResponseWriter, r *http.Request){
	
	switch r.Method{
		case "GET":
		renderAuth(w,"auth.html", nil)
	case "POST":
		username := r.FormValue("username")
		enteredPassword := r.FormValue("password")
		f, err :=os.OpenFile("./users/users",os.O_RDWR | os.O_CREATE, 0600 )
		if err != nil{
			println("error opening the file")
		}
		defer f.Close()
		decoder := json.NewDecoder(f)
		users := make(map[string]string)
		err = decoder.Decode(&users)

	if err != nil{
		println("error decoding")
	}else {
		if _,ok :=users[username]; ok{
			println("user is there")
			storedPassword := users[username]
			err := bcrypt.CompareHashAndPassword( []byte(storedPassword), []byte(enteredPassword))
			if err != nil{
				println("password is wrong")
				renderAuth(w, "auth.hmtl", "Wrong Password!")

			}else {
				//do coockie stuff
				http.Redirect(w, r, "/view/start", http.StatusFound)
			}


		}else{
			println("guessing not ok?")
			http.Redirect(w,r,"/reg/", http.StatusFound)
		}
	}
		
	}

}
func renderAuth (w http.ResponseWriter, tmpl string , data interface{}){
	templates.ExecuteTemplate(w, "auth.html", data)
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
	

	
	
	

		
	//some error handling
	f, err := os.OpenFile("./users/users", os.O_RDWR | os.O_CREATE, 0600)
	//some error handling
	if err != nil{
		println("where is this file everyone's talking about?")
		return

	}
	defer f.Close()
	decoder := json.NewDecoder(f)
	users := make(map[string]string)
	err = decoder.Decode(&users)
	if err != nil{
		println("error decoding")
	}else {
		if _,ok :=users[username]; ok{
			renderRegister(w, "register.html", "User already exists!")
			println("user already exists")

		}else{

			cryptPassword, err := Crypt([]byte(password))
			if err != nil{
				//log.Fatal(err)
			}
			users[username] = string(cryptPassword)
		}
	}
	jsonedusers, err :=json.Marshal(users)
		println(string(jsonedusers))
		if err != nil{
			println("did we jsoned the users yet?")

		}

	ioutil.WriteFile("./users/users", jsonedusers, 0600)
	/*if err != nil{
		println("Nope.Didn't write that shit!")
		return
	}*/
	f.Close()
	http.Redirect(w, r, "/auth/", http.StatusFound)
}

}
func Register( w http.ResponseWriter, r *http.Request){

}

func renderRegister (w http.ResponseWriter, tmpl string, data interface{}){
	templates.ExecuteTemplate(w, "register.html", data)
}

func clear (b []byte){
	for i := 0; i< len(b); i++{
		b[i] =0;
	}
}
func Crypt(password []byte) ([]byte, error){
	defer clear(password)
	return bcrypt.GenerateFromPassword(password, bcrypt.DefaultCost)
}