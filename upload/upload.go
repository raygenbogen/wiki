package upload

import (
	"crypto/md5"
	"encoding/hex"
	"html/template"
	"io"
	"net/http"
	"os"
	"regexp"
	"time"
)

var hashedTime string
var validType = regexp.MustCompile("^.*.(gif|jpeg|jpg)$")
var templates = template.Must(template.ParseFiles("./static/upload.html"))

func renderUpload(w http.ResponseWriter, tmpl string, data interface{}) {
	templates.ExecuteTemplate(w, "upload.html", data)
}

func UploadHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {

	case "GET":
		renderUpload(w, "upload", nil)

	case "POST":

		reader, err := r.MultipartReader()
		if err != nil {

			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		h := md5.New()
		const layout = "2006-01-02 15:04:05"
		t := time.Now()

		hashedTime = t.Format(layout)

		hashedTime = hex.EncodeToString(h.Sum([]byte(hashedTime)))

		os.Mkdir("data/upload/"+hashedTime+"/", 0700)

		for {
			part, err := reader.NextPart()
			if err == io.EOF {
				break
			}

			if part.FileName() == "" {
				continue
			}
			m := validType.FindStringSubmatch(part.FileName())
			if m == nil {

				continue

			}

			dst, err := os.Create("data/upload/" + hashedTime + "/" + part.FileName())
			defer dst.Close()

			if err != nil {

				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			if _, err := io.Copy(dst, part); err != nil {

				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		}

		http.Redirect(w, r, "/data/upload/"+hashedTime, http.StatusFound)

	}
}
