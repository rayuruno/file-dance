package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
)

type ctxKey string

const (
	ck ctxKey = "proxyUrl"
)

func serveHTTP() error {
	return http.ListenAndServe(":8080", proxyHandler())
}

func proxyHandler() http.Handler {
	rp := httputil.ReverseProxy{
		Rewrite: func(r *httputil.ProxyRequest) {
			proxyUrl, ok := r.In.Context().Value(ck).(*url.URL)
			if !ok {
				log.Println("remote proxy not found", ck)
				return
			}
			r.SetURL(proxyUrl)
			r.Out.Host = r.In.Host
		},
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !(r.Method == http.MethodGet || r.Method == http.MethodHead) {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		user := parseSubdomain(r.Host)
		proxyAddr := getProxyAddress(user)
		if proxyAddr == "" {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		proxyUrl, err := url.Parse(fmt.Sprintf("http://%s", proxyAddr))
		if err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		ctx := context.WithValue(r.Context(), ck, proxyUrl)
		rp.ServeHTTP(w, r.Clone(ctx))
	})
}

func parseSubdomain(host string) string {
	return strings.TrimSuffix(host, ".file.dance")
}
