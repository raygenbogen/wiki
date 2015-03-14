package main

import (
	"./auth"
	"./cert"
	"./upload"
	"./view"
	"log"
	"net/http"
	"os"
)

func main() {
	if _, certError := os.Stat("cert.pem"); os.IsNotExist(certError) {
		cert.Start()
	}
	if _, keyError := os.Stat("key.pem"); os.IsNotExist(keyError) {
		cert.Start()
	}
	http.HandleFunc("/favicon.ico", handlerICon)
	http.HandleFunc("/view/", auth.Chkauth(view.MakeHandler(view.ViewHandler)))
	http.HandleFunc("/edit/", auth.Chkauth(view.MakeHandler(view.EditHandler)))
	http.HandleFunc("/save/", auth.Chkauth(view.MakeHandler(view.SaveHandler)))
	http.HandleFunc("/upload/", auth.Chkauth(upload.UploadHandler))
	http.HandleFunc("/vers/", auth.Chkauth(view.MakeVersionHandler(view.VersionHandler)))
	http.HandleFunc("/users/", auth.Chkauth(view.MakeUserHandler(view.UserHandler)))
	http.HandleFunc("/files/", auth.Chkauth(view.MakeFileHandler(view.FileHandler)))
	http.HandleFunc("/logout/", auth.Logout)
	http.HandleFunc("/changeApprovalstatus/", auth.Chkauth(view.MakeApprovalHandler(view.ChangeApprovalstatus)))
	http.HandleFunc("/changeAdminstatus/", auth.Chkauth(view.MakeAdminHandler(view.ChangeAdminstatus)))
	http.HandleFunc("/remove/", auth.Chkauth(view.DeleteUserHandler(view.DeleteUser)))
	http.HandleFunc("/register/", auth.Register)
	http.HandleFunc("/auth/", auth.Auth)
	http.Handle("/static/css/", view.HandlerToHandleFunc(http.StripPrefix("/static/", http.FileServer(http.Dir("./static/")))))
	http.Handle("/static/", auth.Chkauth(view.HandlerToHandleFunc(http.StripPrefix("/static/", http.FileServer(http.Dir("./static/"))))))
	http.Handle("/data/", auth.Chkauth(func(w http.ResponseWriter, r *http.Request){http.ServeFile(w,r,r.URL.Path[1:])}))
	http.HandleFunc("/", view.MakeRedirectHandler("/view/start"))
	//http.ListenAndServe(":8080", nil)
	if err := http.ListenAndServeTLS(":10443", "cert.pem", "key.pem", nil); err != nil {
		log.Fatalf("ListenAndServeTLS error: %v", err)
	}

}
func handlerICon(w http.ResponseWriter, r *http.Request) { http.ServeFile(w, r, "./static/favicon.ico") }
