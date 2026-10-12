package admin

import (
	"io/fs"
	"mime"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/labstack/echo/v4"
)

func init() {
	mime.AddExtensionType(".js", "application/javascript")
	mime.AddExtensionType(".css", "text/css")
	mime.AddExtensionType(".svg", "image/svg+xml")
	mime.AddExtensionType(".woff", "font/woff")
	mime.AddExtensionType(".woff2", "font/woff2")
}

func RegisterHandlers(e *echo.Echo) {
	e.GET("/*", handleStaticFS)
}

func handleStaticFS(c echo.Context) error {
	reqPath := filepath.Clean(c.Request().URL.Path)
	reqPath = strings.TrimPrefix(reqPath, "/")
	if reqPath == "" || reqPath == "." {
		reqPath = "index.html"
	}
	content, err := fs.ReadFile(DistFS, "dist/"+reqPath)
	if err != nil {
		indexContent, err := fs.ReadFile(DistFS, "dist/index.html")
		if err != nil {
			return c.String(http.StatusNotFound, "Admin not built. Run 'npm run build:admin' first.")
		}
		c.Response().Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
		return c.HTMLBlob(http.StatusOK, indexContent)
	}
	contentType := mime.TypeByExtension(filepath.Ext(reqPath))
	if contentType == "" {
		contentType = http.DetectContentType(content)
	}
	if strings.HasPrefix(reqPath, "assets/") {
		c.Response().Header().Set("Cache-Control", "public, max-age=31536000, immutable")
	} else {
		c.Response().Header().Set("Cache-Control", "public, max-age=3600")
	}
	return c.Blob(http.StatusOK, contentType, content)
}
