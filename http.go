package main

import (
	"bytes"
	"crypto/tls"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/NYTimes/gziphandler"
	"github.com/bddjr/hlfhr"
	"golang.org/x/net/http2"
)

// inits http server
func HTTP_Init(port string) (*hlfhr.Server, error) {
	log.Println("HTTP_Init: Init HTTP server.")

	// configure server (hlfhr is used to redirect http to https)
	srv := hlfhr.New(&http.Server{
		Addr:    port,
		Handler: gziphandler.GzipHandler(http.DefaultServeMux),
		TLSConfig: &tls.Config{
			MinVersion: tls.VersionTLS13,
		},
		ReadTimeout:    5 * time.Second,
		WriteTimeout:   10 * time.Second,
		IdleTimeout:    120 * time.Second,
		MaxHeaderBytes: 16 << 10,
	})

	srv.HttpOnHttpsPortErrorHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hlfhr.RedirectToHttps(w, r, http.StatusPermanentRedirect)
	})

	// http/2 support
	if err := http2.ConfigureServer(srv.Server, nil); err != nil {
		return nil, err
	}

	// cache 404 page
	data, err := G_Frontend.ReadFile("frontend-dist/404.html")
	if err != nil {
		return nil, err
	}
	G_FrontendCache["404.html"] = data

	return srv, nil
}

func HTTP_NotFound(w http.ResponseWriter, r *http.Request, path string) {
	if path == "404.html" {
		http.Error(w, "Not Found", http.StatusNotFound)
	} else {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprint(w, string(G_FrontendCache["404.html"]))
	}
}

func HTTP_ServeStatic(w http.ResponseWriter, r *http.Request, path string) {
	var data []byte

	if G_FrontendOverride == "" {
		var ok bool
		G_FrontendMu.RLock()
		data, ok = G_FrontendCache[path]
		G_FrontendMu.RUnlock()

		if !ok {
			G_FrontendMu.RLock()
			defer G_FrontendMu.RUnlock()
			b, err := G_Frontend.ReadFile("frontend-dist/" + path)
			if err != nil {
				HTTP_NotFound(w, r, path)
				return
			}
			G_FrontendCache[path] = b
			data = b
		}
	} else {
		var err error
		data, err = os.ReadFile(G_FrontendOverride + "/" + path)
		if err != nil {
			log.Println("HTTP_ServeStatic:", err)
			HTTP_NotFound(w, r, path)
			return
		}
	}

	w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
	http.ServeContent(w, r, path, G_StartTime, bytes.NewReader(data))
}
