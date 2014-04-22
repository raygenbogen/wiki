package auth

import (
	"code.google.com/p/go.crypto/bcrypt"
	"encoding/json"
	"html/template"
	"io/ioutil"
	"net/http"
	"os"
	"time"
	"fmt"
	"encoding/hex"
	"crypto/md5"
)

/*type Users struct {
	Username string
	Password string
	Approved string
	Admin string
}*/

var templates = template.Must(template.ParseFiles("./static/auth.html", "./static/register.html"))

func Auth(w http.ResponseWriter, r *http.Request) {

	switch r.Method {
	case "GET":
		renderAuth(w, "auth.html", nil)
	case "POST":
		username := r.FormValue("username")
		enteredPassword := r.FormValue("password")
		f, err := os.OpenFile("./users/users", os.O_RDWR|os.O_CREATE, 0600)
		fi, erro := os.OpenFile("./users/approvedusers", os.O_RDWR|os.O_CREATE, 0600)
		if err != nil {
			println("error opening the file")
		}
		if erro != nil {

		}
		defer f.Close()
		defer fi.Close()
		decoder := json.NewDecoder(f)
		approvaldecoder := json.NewDecoder(fi)
		users := make(map[string]string)
		approvedusers := make(map[string]string)
		err = decoder.Decode(&users)
		erro = approvaldecoder.Decode(&approvedusers)
		if erro != nil{

		}
		if err != nil {
			println("error decoding")
			http.Redirect(w,r, "/register", http.StatusFound)
		} else {
			if _, ok := users[username]; ok {

				println("user is there")
				if _, okay := approvedusers[username]; okay{
					println("user is also in approvedusers")
					stateofapproval := approvedusers[username]
					if stateofapproval != "approved" {
						println("not approved")
						renderAuth(w, "auth.html", "You are not yet approved! Please contact one of the Administrators!")
					}
				}
				storedPassword := users[username]
				err := bcrypt.CompareHashAndPassword([]byte(storedPassword), []byte(enteredPassword))
				if err != nil {
					println("password is wrong")
					renderAuth(w, "auth.hmtl", "Wrong Password!")

				} else {
					//do cookie stuff
					t := time.Now()
					expiration := time.Now().AddDate(1,0,0)
					fmt.Println(expiration)
					//expirationexpiration.Year += 1
					h := md5.New()
					const layout = "2006-01-02 15:04:05"
					hashedTime := t.Format(layout)
					hashedStoredPassword := hex.EncodeToString(h.Sum([]byte(storedPassword)))
					cookie := http.Cookie{Name: "User", Value: username, Path: "/", Expires: expiration}
					cookie2 := http.Cookie{Name: hashedStoredPassword, Value: hashedTime, Path: "/" ,Expires:expiration}
					http.SetCookie(w, &cookie)  
					http.SetCookie(w, &cookie2)
					fmt.Println(cookie.Value)
					fmt.Println(cookie2.Name)
					http.Redirect(w, r, "/view/start", http.StatusFound)

				}


			} else {
				println("guessing not ok?")
				http.Redirect(w, r, "/register/", http.StatusFound)
			}
		}

	}

}
func renderAuth(w http.ResponseWriter, tmpl string, data interface{}) {
	templates.ExecuteTemplate(w, "auth.html", data)
}

func RegisterHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {

	case "GET":
		renderRegister(w, "register.html", nil)
	case "POST":
		username := r.FormValue("username")
		password := r.FormValue("password")
 		println(username)
		println(password)
		if len(password) < 8 {
			println("more letters")
			renderRegister(w, "register.html", "More Letters.Do it.Do it Now!")
			return
		}

		//some error handling
		f, err := os.OpenFile("./users/users", os.O_RDWR|os.O_CREATE, 0600)
		fi, erro := os.OpenFile("./users/admins", os.O_RDWR|os.O_CREATE, 0600)
		fil, er := os.OpenFile("./users/approvedusers", os.O_RDWR|os.O_CREATE, 0600)
		if erro != nil{
			println("no adminsFile")
		}
		if er != nil{

		}

		defer fil.Close()
		defer fi.Close()

		//some error handling
		if err != nil {
			println("where is this file everyone's talking about?")
			return

		}
		defer f.Close()
		decoder := json.NewDecoder(f)
		admindecoder := json.NewDecoder(fi)
		approvaldecoder := json.NewDecoder(fil)
		users := make(map[string]string)
		admins := make (map[string]string)
		approvedusers := make (map[string]string)
		err = decoder.Decode(&users)
		erro = admindecoder.Decode(&admins)
		er = approvaldecoder.Decode(&approvedusers)
		if err != nil {
			println("error decoding")
		} else {
			if _, ok := users[username]; ok {
				renderRegister(w, "register.html", "User already exists!")
				println("user already exists")

			} else {

				cryptPassword, err := Crypt([]byte(password))
				if err != nil {
					//log.Fatal(err)
				}
				users[username] = string(cryptPassword)
				admins[username] = "IsNotAdmin"
				approvedusers[username] = "NotYetApproved"
			}
		}
		jsonedadmins, er := json.Marshal(admins)
		jsonedapprovals, erro := json.Marshal(approvedusers)
		jsonedusers, err := json.Marshal(users)
		println(string(jsonedusers))
		if err != nil {
			println("did we jsoned the users yet?")

		}
		if er != nil{

		}
		if erro != nil {

		}
		ioutil.WriteFile("./users/users", jsonedusers, 0600)
		f.Close()
		ioutil.WriteFile("./users/approvedusers", jsonedapprovals, 0600)
		ioutil.WriteFile("./users/admins", jsonedadmins, 0600)
		http.Redirect(w, r, "/auth/", http.StatusFound)
	}

}
func Register(w http.ResponseWriter, r *http.Request) {

}

func renderRegister(w http.ResponseWriter, tmpl string, data interface{}) {
	templates.ExecuteTemplate(w, "register.html", data)
}

func clear(b []byte) {
	for i := 0; i < len(b); i++ {
		b[i] = 0
	}
}
func Crypt(password []byte) ([]byte, error) {
	defer clear(password)
	return bcrypt.GenerateFromPassword(password, bcrypt.DefaultCost)
}

func Chkauth(f http.HandlerFunc) http.HandlerFunc{
	return func(w http.ResponseWriter, r *http.Request){
		fmt.Println("do i even enter this function?")
		if r == nil{
			fmt.Println("there is no request.ficken")
		}
		cookie, err := r.Cookie("User")
		
		if err!= nil{
			fmt.Println("i'm not reading anything cookielike")
			fmt.Println(err)
			http.Redirect(w,r, "/auth/",http.StatusFound)
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

		if err != nil {
			println("error decoding")
			http.Redirect(w,r,"/auth", http.StatusFound)
		} else {
			if _, ok := users[username]; ok {
				println("user is there")
				storedPassword := users[username]
				h := md5.New()
				hashedStoredPassword := hex.EncodeToString(h.Sum([]byte(storedPassword)))
				_, err := r.Cookie(hashedStoredPassword)
				if err != nil{
					println("no cookie found")
					http.Redirect(w,r, "/auth/", http.StatusFound)
					return
				}
			}else {
			println("no user found")
			http.Redirect(w,r, "/auth", http.StatusFound)
		}
		}


		f(w,r)
	}

}