package main

import (
	"gowiki/auth"
	"gowiki/cert"
	"gowiki/upload"
	"gowiki/view"
	"log"
	"net/http"
	"os"
)

func main() {
	if _, certError := os.Stat("static/cert.pem"); os.IsNotExist(certError) {
		cert.Start()
	}
	if _, keyError := os.Stat("static/key.pem"); os.IsNotExist(keyError) {
		cert.Start()
	}

	http.HandleFunc("/view/", auth.Chkauth(view.MakeHandler(view.ViewHandler)))
	http.HandleFunc("/edit/", auth.Chkauth(view.MakeHandler(view.EditHandler)))
	http.HandleFunc("/save/", auth.Chkauth(view.MakeHandler(view.SaveHandler)))
	http.HandleFunc("/upload/", auth.Chkauth(upload.UploadHandler))
	http.HandleFunc("/vers/", auth.Chkauth(view.MakeVersionHandler(view.VersionHandler)))
	http.HandleFunc("/users/", auth.Chkauth(view.MakeUserHandler(view.UserHandler)))
	http.HandleFunc("/data/fileserver/", auth.Chkauth(view.MakeFileHandler(view.FileHandler)))
	//http.HandleFunc("/reg/", auth.RegisterHandler)
	http.HandleFunc("/register/", auth.RegisterHandler)
	http.HandleFunc("/auth/", auth.Auth)
	http.Handle("/static/css/", view.HandlerToHandleFunc(http.StripPrefix("/static/", http.FileServer(http.Dir("./static/")))))
	http.Handle("/static/", auth.Chkauth(view.HandlerToHandleFunc(http.StripPrefix("/static/", http.FileServer(http.Dir("./static/"))))))
	//http.Handle("/data/fileserver/", auth.Chkauth(view.HandlerToHandleFunc(http.StripPrefix("/data/",http.FileServer(http.Dir("./data/")))))
	http.Handle("/data/", auth.Chkauth(view.HandlerToHandleFunc(http.StripPrefix("/data/", http.FileServer(http.Dir("./data/"))))))
	http.HandleFunc("/", view.MakeRedirectHandler("/view/start"))
	if err := http.ListenAndServeTLS(":10443", "./static/cert.pem", "./static/key.pem", nil); err != nil {
		log.Fatalf("ListenAndServeTLS error: %v", err)
	}

}

/*func startHandler(w http.ResponseWriter, r *http.Request) {

}*/
