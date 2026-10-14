package assets

import (
	"embed"
	"net/http"
	"strings"

	oauthserver "github.com/openshift/oauth-server/pkg"
)

const Prefix = "/oauth/assets/v1/"
const ContentSecurityPolicy = "script-src 'self'; object-src 'none'; base-uri 'none'"

//go:embed *.js
var scripts embed.FS

func SetBuiltinHTMLHeaders(w http.ResponseWriter) {
	w.Header().Set("Content-Security-Policy", ContentSecurityPolicy)
	w.Header().Set("Cache-Control", "no-cache, no-store, max-age=0, must-revalidate")
}

func Install(mux oauthserver.Mux) {
	mux.HandleFunc(Prefix, serveScript)
}

func serveScript(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.Header().Set("Cache-Control", "no-cache, no-store, max-age=0, must-revalidate")
	if req.Method != http.MethodGet && req.Method != http.MethodHead {
		w.Header().Set("Allow", "GET, HEAD")
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	name := strings.TrimPrefix(req.URL.Path, Prefix)
	switch name {
	case "theme.js", "login.js", "token.js":
	default:
		http.NotFound(w, req)
		return
	}
	content, err := scripts.ReadFile(name)
	if err != nil {
		http.Error(w, "script unavailable", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/javascript; charset=utf-8")
	if req.Method == http.MethodGet {
		_, _ = w.Write(content)
	}
}
