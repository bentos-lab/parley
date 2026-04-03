package rest

import (
	"embed"
	"io/fs"
	"net/http"

	"github.com/go-chi/chi/v5"
)

//go:embed dist
var embeddedDist embed.FS

var (
	fileServer http.Handler
	spaFiles   fs.FS
	indexHTML  []byte
)

func init() {
	var err error
	spaFiles, err = fs.Sub(embeddedDist, "dist")
	if err != nil {
		panic("rest: unable to initialize embedded SPA assets: " + err.Error())
	}

	indexHTML, err = fs.ReadFile(spaFiles, "index.html")
	if err != nil {
		panic("rest: unable to read SPA index.html: " + err.Error())
	}

	fileServer = http.FileServer(http.FS(spaFiles))
}

// spaHandler ensures the root path returns the embedded index file and delegates other requests to the file server.
// Parameters: none. Returns: an http.Handler that serves embedded SPA files.
func spaHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/" {
			serveIndex(w)
			return
		}
		fileServer.ServeHTTP(w, r)
	})
}

// serveIndex writes the embedded SPA index file to the response writer.
// Parameters: w is the HTTP response writer. Returns: nothing.
func serveIndex(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, _ = w.Write(indexHTML)
}

// RegisterStaticAssets registers the SPA file server on the provided router.
// Parameters: router is the chi router used to handle requests. Returns: nothing.
func RegisterStaticAssets(router chi.Router) {
	router.Handle("/*", spaHandler())
}
