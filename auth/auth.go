package auth

import (
	"code.google.com/p/go.crypto/bcrypt"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"html/template"
	"net/http"
	"os"
	"time"
	"database/sql"
	_ "github.com/lib/pq"
	"log"
)

var templates = template.Must(template.ParseFiles("./static/templates/main.html", "./static/templates/head.html", "./static/templates/menu.html", "./static/templates/content_auth.html", "./static/templates/footer.html"))

type RenderData struct {
	Title       string
	Headline    string
	Message     string
	Action      string
	Submit      string
	MenuEntries [][2]string
}

type User struct {
	Username       string
	Password       string
	Approvalstatus string
	Adminstatus    string
}

func dbOpen () *sql.DB {
	db, err := sql.Open("postgres", "user=postgres dbname=wiki sslmode=disable")
	if err != nil {
		log.Fatal(err)
		}
	return db
}

func Auth(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		renderAuth(w, "")
	case "POST":

		username := r.FormValue("username")
		enteredPassword := r.FormValue("password")

		db := dbOpen()

		rows, err := db.Query("SELECT password, approved, admin FROM users WHERE name = $1", username)
		if err != nil {
			log.Fatal(err)
		}

		var password, approved, admin string
		var user User

		count_rows := 0
		for rows.Next() {
			count_rows += 1
			err := rows.Scan(&password, &approved, &admin)
			if err != nil {
					log.Fatal(err)
			}
			user = User{username, password, approved, admin}
		}
		if count_rows == 0 {
			renderAuth(w, "This User doesn't seem to exist!")
			return
		}
		defer rows.Close()
		err = rows.Err()
		if err != nil {
			log.Fatal(err)
		}

		if user.Approvalstatus == "approved" {
			err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(enteredPassword))
			if err != nil {
				renderAuth(w, "Wrong Password!")

			} else {
				t := time.Now()
				expiration := time.Now().AddDate(1, 0, 0)
				h := md5.New()
				const layout = "2006-01-02 15:04:05"
				hashedTime := t.Format(layout)
				hashedStoredPassword := hex.EncodeToString(h.Sum([]byte(user.Password)))
				cookie := http.Cookie{Name: "User", Value: username, Path: "/", Expires: expiration}
				cookie2 := http.Cookie{Name: hashedStoredPassword, Value: hashedTime, Path: "/", Expires: expiration}
				http.SetCookie(w, &cookie)
				http.SetCookie(w, &cookie2)
				http.Redirect(w, r, "/view/start", http.StatusFound)

			}

		} else {
			renderAuth(w, "You are not approved. Please contact the Administrator!")

		}
	}
}

func renderAuth(w http.ResponseWriter, message string) {
	templates.ExecuteTemplate(w, "main", &RenderData{
		"Login",
		"Login",
		message,
		"/auth/",
		"Login",
		[][2]string{
			{"Home", "/view/start"},
			{"Create Account", "/register"},
		},
	})
}

func Register(w http.ResponseWriter, r *http.Request) {
	switch r.Method {

	case "GET":
		renderRegister(w, "")
	case "POST":
		username := r.FormValue("username")
		password := r.FormValue("password")
		if len(password) < 8 {
			renderRegister(w, "Please choose a password with at least 8 characters!")
			return
		}
		cryptPassword, err := Crypt([]byte(password))
		if err != nil {
			println("Error encrypting the password")
		}

		db := dbOpen()

		rows, err := db.Query("SELECT name FROM users")
		if err != nil {
			log.Fatal(err)
		}
		var name string
		isnew := true
		count_rows := 0
		for rows.Next() {
			count_rows += 1
			err := rows.Scan(&name)
			if err != nil {
					log.Fatal(err)
			}
			if name == username {
				isnew = false
			}
		}
		if isnew == false {
			renderRegister(w, "This User already exists!")
			return
		}
		if isnew {
			if count_rows == 0 {
				rows, err = db.Query("INSERT INTO users (name, password, approved, admin) VALUES ($1, $2, $3, $4)", username, cryptPassword, "approved", "admin")
				defer rows.Close()
				if err != nil {
					log.Fatal(err)
					return
				}
			} else {
				rows, err = db.Query("INSERT INTO users (name, password, approved, admin) VALUES ($1, $2, $3, $4)", username, cryptPassword, "not approved", "user")
				defer rows.Close()
				if err != nil {
					log.Fatal(err)
					return
				}
			}
			http.Redirect(w, r, "/auth/", http.StatusFound)
		}
	}
}

func renderRegister(w http.ResponseWriter, message string) {
	templates.ExecuteTemplate(w, "main", &RenderData{
		"Create Account",
		"Register",
		message,
		"/register/",
		"Create",
		[][2]string{{"Home", "/view/start"}},
	})
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

func Chkauth(f http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("User")
		if err != nil {
			http.Redirect(w, r, "/auth/", http.StatusFound)
			return
		}
		username := cookie.Value

		db := dbOpen()
		rows, err := db.Query("SELECT password, approved, admin FROM users WHERE name = $1", username)
		if err != nil {
			log.Fatal(err)
		}

		var password, approved, admin string
		var user User

		count_rows := 0
		for rows.Next() {
		    count_rows += 1
		    err := rows.Scan(&password, &approved, &admin)
		    if err != nil {
		        log.Fatal(err)
		    }
		    user = User{username, password, approved, admin}
		}
		if count_rows == 0 {
			//no user found
			http.Redirect(w, r, "/auth/", http.StatusFound)
		    return
		} else {
			h := md5.New()
			hashedStoredPassword := hex.EncodeToString(h.Sum([]byte(user.Password)))
			_, err = r.Cookie(hashedStoredPassword)
			if err != nil {
				println("no cookie found")
				http.Redirect(w, r, "/auth/", http.StatusFound)
				return
			}
			if user.Approvalstatus != "approved" {
				http.Redirect(w, r, "/auth/", http.StatusFound)
			}
		}

		f(w, r)
	}

}

func Logout(w http.ResponseWriter, r *http.Request){
	expiration := time.Now()
	invalidcookie := http.Cookie{Name: "User", Value: "expired", Path: "/", Expires: expiration}
	cookie, err := r.Cookie("User")
	if err != nil {
		http.Redirect(w, r, "/auth/", http.StatusFound)
		return
	}
	username := cookie.Value
	filename := "./users/" + username
	file, erro := os.Open(filename)
	if erro != nil {
		println("Error opening the file")
	}
	defer file.Close()
	decoder := json.NewDecoder(file)
	var user User
	err = decoder.Decode(&user)
	if err != nil {
		println("Error decoding json")
	}
	h := md5.New()
	hashedStoredPassword := hex.EncodeToString(h.Sum([]byte(user.Password)))
	_, err = r.Cookie(hashedStoredPassword)
	if err != nil {
		println("no cookie found")
		http.Redirect(w, r, "/auth/", http.StatusFound)
		return
	}
	invalidcookie2 := http.Cookie{Name: hashedStoredPassword, Value: "expired", Path: "/", Expires: expiration}
	http.SetCookie(w, &invalidcookie)
	http.SetCookie(w, &invalidcookie2)
	http.Redirect(w, r, "/view/start", http.StatusFound)
}
