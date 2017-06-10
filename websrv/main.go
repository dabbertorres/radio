package main

import (
	"fmt"
	"log"
	"net/http"
	"path/filepath"
	"time"

	"MediaServer/internal/interrupt"
	"MediaServer/websrv/file"
	"MediaServer/urlgen"
)

var (
	registry *file.Registry

	// annoying to type, short for "sanitize"
	san = filepath.Clean
)

func main() {
	log.SetFlags(log.Llongfile | log.LstdFlags)
	
	var err error

	registry, err = file.NewRegistry("app")
	if err != nil {
		panic(err)
	}
	defer registry.Close()

	err = registry.Walk(nil)
	if err != nil {
		panic(err)
	}
	
	err = urlgen.Load()
	if err != nil {
		panic(err)
	}

	serverMux := http.NewServeMux()

	// basic server files!
	serverMux.HandleFunc("/", customHandler(san("app/html/index.html"), "text/html"))
	serverMux.HandleFunc("/css/", handler("text/css"))
	serverMux.HandleFunc("/html/", handler("text/html"))
	serverMux.HandleFunc("/js/", handler("text/javascript"))

	// images!
	serverMux.HandleFunc("/img/png", handler("image/png"))
	serverMux.HandleFunc("/img/svg", handler("image/svg+xml"))

	// favicon config crap (bless you, realfavicongenerator.net)
	serverMux.HandleFunc("/browserconfig.xml", customHandler(san("app/browserconfig.xml"), "application/xml"))
	serverMux.HandleFunc("/manifest.json", customHandler(san("app/manifest.json"), "application/json"))

	// actual favicons
	serverMux.HandleFunc("/android-chrome-192x192.png", customHandler(san("app/img/favicon/android-chrome-192x192.png"), "image/png"))
	serverMux.HandleFunc("/android-chrome-512x512.png", customHandler(san("app/img/favicon/android-chrome-512x512.png"), "image/png"))
	serverMux.HandleFunc("/apple-touch-icon.png", customHandler(san("app/img/favicon/apple-touch-icon.png"), "image/png"))
	serverMux.HandleFunc("/favicon.ico", customHandler(san("app/img/favicon/favicon.ico"), "image/x-icon"))
	serverMux.HandleFunc("/favicon.png", customHandler(san("app/img/favicon/favicon.png"), "image/png"))
	serverMux.HandleFunc("/favicon-16x16.png", customHandler(san("app/img/favicon/favicon-16x16.png"), "image/png"))
	serverMux.HandleFunc("/favicon-32x32.png", customHandler(san("app/img/favicon/favicon-32x32.png"), "image/png"))
	serverMux.HandleFunc("/mstile-150x150.png", customHandler(san("app/img/favicon/mstile-150x150.png"), "image/png"))
	serverMux.HandleFunc("/safari-pinned-tab.svg", customHandler(san("app/img/favicon/safari-pinned-tab.svg"), "image/svg"))

	// actually interesting stuff eventually
	serverMux.HandleFunc("/song/", songHandler)
	serverMux.HandleFunc("/station/", stationHandler)
	serverMux.HandleFunc("/search/", searchHandler)

	server := http.Server{
		Addr:           ":8080",
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 32,
		Handler:        serverMux,
	}

	// listen for termination signal!
	interrupt.OnExit(func() { server.Close() })

	// hey now we can do what we want
	fmt.Println("Serving...")
	err = server.ListenAndServe()
	if err == http.ErrServerClosed {
		fmt.Println("Done")
	} else {
		log.Println("Server shutdown unexpectedly:", err)
	}
}

func customHandler(path string, mimeType string) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()

		data := registry.Get(path)
		if data != nil {
			if mimeType != "" {
				w.Header().Add("Content-Type", mimeType)
			}

			w.Write(data)
		} else {
			w.WriteHeader(http.StatusNotFound)
		}
	}
}

func handler(mimeType string) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		path := filepath.Join(registry.BasePath, r.URL.EscapedPath())
		path = filepath.Clean(path)

		data := registry.Get(path)
		if data != nil {
			w.Header().Add("Content-Type", mimeType)
			w.Write(data)
		} else {
			w.WriteHeader(http.StatusNotFound)
		}
	}
}

func songHandler(w http.ResponseWriter, r *http.Request) {
	// TODO request file from the database
	
	w.Header().Add("Accept-Ranges", "bytes")
	w.WriteHeader(http.StatusPartialContent)
}

func stationHandler(w http.ResponseWriter, r *http.Request) {
	// TODO return specified station
}

func searchHandler(w http.ResponseWriter, r *http.Request) {
	// TODO return results for specified search parameters
}
