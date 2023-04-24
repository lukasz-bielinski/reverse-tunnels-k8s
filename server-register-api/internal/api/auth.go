package api

import (
	"net/http"
	"os"
)

func BasicAuth(handler http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		username, password, ok := r.BasicAuth()

		if !ok || username != os.Getenv("USERNAME") || password != os.Getenv("PASSWORD") {
			w.Header().Set("WWW-Authenticate", `Basic realm="Restricted"`)
			http.Error(w, "Unauthorized.", http.StatusUnauthorized)
			return
		}

		handler(w, r)
	}
}
