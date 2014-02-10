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
	http.HandleFunc("/view/", view.MakeHandler(view.ViewHandler))
	http.HandleFunc("/edit/", view.MakeHandler(view.EditHandler))
	http.HandleFunc("/save/", view.MakeHandler(view.SaveHandler))
	http.HandleFunc("/upload/", upload.UploadHandler)
	//http.HandleFunc("/reg/", auth.RegisterHandler)
	http.HandleFunc("/register/", auth.RegisterHandler)
	http.HandleFunc("/auth/", auth.Auth)
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("./static/"))))
	http.Handle("/data/", http.StripPrefix("/data/", http.FileServer(http.Dir("./data/"))))
	http.HandleFunc("/", view.MakeRedirectHandler("/view/start"))
	if err := http.ListenAndServeTLS(":10443", "./static/cert.pem", "./static/key.pem", nil); err != nil {
		log.Fatalf("ListenAndServeTLS error: %v", err)
	}

}

/*func startHandler(w http.ResponseWriter, r *http.Request) {

}*/
