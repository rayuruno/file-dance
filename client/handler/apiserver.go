package handler

import (
	"encoding/json"
	"errors"
	"io/fs"
	"mime"
	"net/http"
	"net/url"
	"path"
	"strings"
)

func ApiServer(httpDir string) http.Handler {
	return maxAgeHandler(60, apiServer(httpDir))
}
func apiServer(httpDir string) http.Handler {
	dir := http.Dir(httpDir)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		upath := r.URL.Path
		if !strings.HasPrefix(upath, "/") {
			upath = "/" + upath
			r.URL.Path = upath
		}
		name := path.Clean(upath)
		f, err := dir.Open(name)

		if err != nil {
			msg, code := toHTTPError(err)
			writeError(w, msg, code)
			return
		}
		defer f.Close()

		d, err := f.Stat()
		if err != nil {
			msg, code := toHTTPError(err)
			writeError(w, msg, code)
			return
		}
		if d.IsDir() {
			// if checkIfModifiedSince(r, d.ModTime()) == condFalse {
			// 	writeNotModified(w)
			// 	return
			// }
			// setLastModified(w, d.ModTime())
			dirList(w, r, f)
			return
		}
		http.ServeFile(w, r, path.Join(httpDir, name))
	})
}

func toHTTPError(err error) (msg string, httpStatus int) {
	if errors.Is(err, fs.ErrNotExist) {
		return "404 page not found", http.StatusNotFound
	}
	if errors.Is(err, fs.ErrPermission) {
		return "403 Forbidden", http.StatusForbidden
	}
	// Default:
	return "500 Internal Server Error", http.StatusInternalServerError
}

func writeError(w http.ResponseWriter, error string, code int) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(code)
	w.Write([]byte(`{"error":"` + error + `"}`))
}

func dirList(w http.ResponseWriter, r *http.Request, f http.File) {
	dirs, err := f.Readdir(-1)

	if err != nil {
		writeError(w, "Error reading directory", http.StatusInternalServerError)
		return
	}
	// sort.Slice(dirs, func(i, j int) bool { return dirs.name(i) < dirs.name(j) })

	// w.Header().Set("Content-Type", "text/html; charset=utf-8")
	// fmt.Fprintf(w, "<pre>\n")
	out := make([]map[string]any, len(dirs))
	for i, n := 0, len(dirs); i < n; i++ {
		name := dirs[i].Name()
		if dirs[i].IsDir() {
			name += "/"
		}
		// name may contain '?' or '#', which must be escaped to remain
		// part of the URL path, and not indicate the start of a query
		// string or fragment.
		url := url.URL{Path: name}

		out[i] = map[string]any{
			"url":  url.String(),
			"dir":  dirs[i].IsDir(),
			"type": mime.TypeByExtension(path.Ext(dirs[i].Name())),
		}
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	// fmt.Fprintf(w, "</pre>\n")
	json.NewEncoder(w).Encode(out)
}
