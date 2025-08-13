package main

import (
	"crypto/tls"
	"log"
	"net/http"
	"time"

	"github.com/NYTimes/gziphandler"
	"github.com/bddjr/hlfhr"
	"golang.org/x/net/http2"
)

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

	return srv, nil
}
