package view
/**
  This package defines different proxyfunctions that produce a http.HandlerFunc.
*/
import (
	"net/http"
)

func MakeHandler(fn func(http.ResponseWriter, *http.Request, string)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		m := validPath.FindStringSubmatch(r.URL.Path)
		if m == nil {
			title := "start"
			http.Redirect(w, r, "/view/"+title, http.StatusFound)
		} else {
			fn(w, r, m[2])
		}
	}
}

func MakeVersionHandler(fn func(http.ResponseWriter, *http.Request, string, *string)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		println(r.URL.Path)
		m := versPath.FindStringSubmatch(r.URL.Path)
		if m == nil {
			n := validPath.FindStringSubmatch(r.URL.Path)
			if n == nil {
				http.Redirect(w, r, "/view/start", http.StatusFound)
				return
			}
			// list all versions of an articles
			fn(w, r, n[2], nil)
		} else {
			// show specific version
			fn(w, r, m[2], &m[3])
		}
	}
}

func MakeUserHandler(fn func(http.ResponseWriter, *http.Request, string)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		m := validPath.FindStringSubmatch(r.URL.Path)
		println(r.URL.Path)
		if m == nil {
			println("not a valid path")
			n := userPath.FindStringSubmatch(r.URL.Path)
			if n == nil {
				println("not even a users path")
				println(n)
				http.Redirect(w, r, "/users/", http.StatusFound)
				return
			}
			fn(w, r, "")
		} else {
			fn(w, r, m[2])
		}
	}
}

func MakeRedirectHandler(target string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, target, http.StatusFound)
		return
	}
}

func MakeApprovalHandler(fn func(http.ResponseWriter, *http.Request, string)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		m := approvalPath.FindStringSubmatch(r.URL.Path)
		if m == nil {
			return
		}
		fn(w, r, m[2])
	}
}

func MakeAdminHandler(fn func(http.ResponseWriter, *http.Request, string)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		m := adminPath.FindStringSubmatch(r.URL.Path)
		if m == nil {
			return
		}
		fn(w, r, m[2])
	}
}

func MakeFileHandler(fn func(http.ResponseWriter, *http.Request, string)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		m := filePath.FindStringSubmatch(r.URL.Path)
		if m == nil {
			println("not a valid path")
			return
		}
		fn(w, r, r.URL.Path)
	}
}
