package dashboard

import (
	"embed"
	"io/fs"
	"net/http"
)

//go:embed static
var embedded embed.FS

func New() http.Handler {
	sub, _ := fs.Sub(embedded, "static")
	return http.FileServer(http.FS(sub))
}
