package handler

import (
	"fmt"
	"net/http"
)

func FileServer(httpDir string) http.Handler {
	return maxAgeHandler(60, http.FileServer(http.Dir(httpDir)))
}

func maxAgeHandler(seconds int, h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Cache-Control", fmt.Sprintf("max-age=%d, public, must-revalidate, proxy-revalidate", seconds))
		h.ServeHTTP(w, r)
	})
}
