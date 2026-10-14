package assets_test

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/openshift/oauth-server/pkg/api"
	"github.com/openshift/oauth-server/pkg/server/assets"
	"github.com/openshift/oauth-server/pkg/server/errorpage"
	"github.com/openshift/oauth-server/pkg/server/locales"
	"github.com/openshift/oauth-server/pkg/server/login"
	"github.com/openshift/oauth-server/pkg/server/selectprovider"
)

func TestScripts(t *testing.T) {
	mux := http.NewServeMux()
	assets.Install(mux)
	for _, test := range []struct {
		name, method string
		status       int
	}{
		{"theme.js", "GET", 200}, {"login.js", "GET", 200}, {"token.js", "GET", 200},
		{"token.js", "HEAD", 200}, {"token.js", "POST", 405}, {"missing.js", "GET", 404},
		{"nested/token.js", "GET", 404},
	} {
		t.Run(test.method+test.name, func(t *testing.T) {
			response := httptest.NewRecorder()
			mux.ServeHTTP(response, httptest.NewRequest(test.method, assets.Prefix+test.name, nil))
			if response.Code != test.status {
				t.Fatalf("status %d", response.Code)
			}
			if test.status == 200 && response.Header().Get("Content-Type") != "text/javascript; charset=utf-8" {
				t.Fatal("incorrect MIME")
			}
			if response.Header().Get("X-Content-Type-Options") != "nosniff" || !strings.Contains(response.Header().Get("Cache-Control"), "no-store") {
				t.Fatal("missing security headers")
			}
			if test.method == "HEAD" && response.Body.Len() != 0 {
				t.Fatal("HEAD returned body")
			}
			if test.method == "GET" && test.status == 200 && response.Body.Len() == 0 {
				t.Fatal("empty script")
			}
		})
	}
}

func TestRenderersBuiltinAndCustom(t *testing.T) {
	hostile := `</script><script>alert("placeholder")</script><img src=x onerror="alert(1)">`
	renderers := []struct {
		name    string
		render  func(string, http.ResponseWriter, *http.Request) error
		scripts int
	}{
		{"login", func(file string, response http.ResponseWriter, request *http.Request) error {
			renderer, err := login.NewLoginFormRenderer(file)
			if err != nil {
				return err
			}
			renderer.Render(login.LoginForm{Error: hostile, Values: login.LoginFormFields{Username: hostile}, Locale: locales.GetLocale("en")}, response, request)
			return nil
		}, 2},
		{"error", func(file string, response http.ResponseWriter, request *http.Request) error {
			renderer, err := errorpage.NewErrorPageTemplateRenderer(file)
			if err != nil {
				return err
			}
			renderer.Render(errorpage.ErrorData{Error: hostile}, response, request)
			return nil
		}, 1},
		{"provider", func(file string, response http.ResponseWriter, request *http.Request) error {
			renderer, err := selectprovider.NewSelectProviderRenderer(file)
			if err != nil {
				return err
			}
			renderer.Render([]api.ProviderInfo{{Name: hostile, URL: "https://example.invalid/provider"}, {Name: "second", URL: "https://example.invalid/second"}}, response, request)
			return nil
		}, 1},
		{"single-provider", func(file string, response http.ResponseWriter, request *http.Request) error {
			renderer, err := selectprovider.NewSelectProviderRenderer(file)
			if err != nil {
				return err
			}
			renderer.Render([]api.ProviderInfo{{Name: hostile, URL: "https://example.invalid/provider"}}, response, request)
			return nil
		}, 1},
	}
	mux := http.NewServeMux()
	assets.Install(mux)
	for _, renderer := range renderers {
		t.Run(renderer.name, func(t *testing.T) {
			for _, custom := range []bool{false, true} {
				file := ""
				if custom {
					file = filepath.Join(t.TempDir(), "custom.html")
					if err := os.WriteFile(file, []byte(`<script>window.customCompatibility=true</script>{{.Locale.LogIn}}`), 0600); err != nil {
						t.Fatal(err)
					}
				}
				response := httptest.NewRecorder()
				request := httptest.NewRequest("GET", "/", nil)
				request.Header.Set("Accept-Language", "en")
				if err := renderer.render(file, response, request); err != nil {
					t.Fatal(err)
				}
				if custom {
					if response.Header().Get("Content-Security-Policy") != "" || !strings.Contains(response.Body.String(), "window.customCompatibility=true") {
						t.Fatal("custom template compatibility changed")
					}
					inherited := httptest.NewRecorder()
					inherited.Header().Set("Content-Security-Policy", "object-src 'none'")
					if err := renderer.render(file, inherited, request); err != nil {
						t.Fatal(err)
					}
					if inherited.Header().Get("Content-Security-Policy") != "object-src 'none'" {
						t.Fatal("custom renderer changed inherited CSP")
					}
					continue
				}
				if response.Header().Get("Content-Security-Policy") != assets.ContentSecurityPolicy {
					t.Fatal("missing builtin CSP")
				}
				if strings.Contains(response.Body.String(), hostile) {
					t.Fatal("unescaped hostile data")
				}

				body := response.Body.String()
				for _, class := range []string{"pf-v6-theme-dark", "pf-v6-theme-high-contrast"} {
					if !strings.Contains(body, "."+class) {
						t.Fatalf("missing generated CSS for %s", class)
					}
				}
				if renderer.name == "login" && !strings.Contains(body, ".pf-m-in-progress") {
					t.Fatal("missing login progress CSS")
				}
				scripts := regexp.MustCompile(`<script\b([^>]*)>([\s\S]*?)</script>`).FindAllStringSubmatch(body, -1)
				if len(scripts) != renderer.scripts {
					t.Fatalf("scripts %d, want %d", len(scripts), renderer.scripts)
				}
				if regexp.MustCompile(`(?i)<[^>]+\s+on[a-z]+\s*=`).MatchString(regexp.MustCompile(`"[^"]*"|'[^']*'`).ReplaceAllString(body, `""`)) {
					t.Fatal("inline event attribute")
				}
				for _, script := range scripts {
					source := regexp.MustCompile(`src="([^"]+)"`).FindStringSubmatch(script[1])
					if len(source) != 2 || !strings.HasPrefix(source[1], assets.Prefix) || strings.TrimSpace(script[2]) != "" {
						t.Fatal("inline or nonlocal script")
					}
					served := httptest.NewRecorder()
					mux.ServeHTTP(served, httptest.NewRequest("GET", source[1], nil))
					if served.Code != 200 {
						t.Fatal("script not served")
					}
				}

			}
			if err := renderer.render(filepath.Join(t.TempDir(), "absent.html"), httptest.NewRecorder(), httptest.NewRequest("GET", "/", nil)); err == nil {
				t.Fatal("missing custom template accepted")
			}
			malformed := filepath.Join(t.TempDir(), "malformed.html")
			if err := os.WriteFile(malformed, []byte("{{if}}"), 0600); err != nil {
				t.Fatal(err)
			}
			if err := renderer.render(malformed, httptest.NewRecorder(), httptest.NewRequest("GET", "/", nil)); err == nil {
				t.Fatal("malformed custom template accepted")
			}
		})
	}
}
