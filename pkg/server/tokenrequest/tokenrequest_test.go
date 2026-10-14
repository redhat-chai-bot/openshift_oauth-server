package tokenrequest

import (
	"html"
	"net/http/httptest"
	"regexp"
	"strings"
	"testing"

	"github.com/openshift/oauth-server/pkg/server/assets"
)

func TestTokenDataEscaping(t *testing.T) {
	for _, value := range []string{"placeholder-not-a-token", `</script><script>alert(1)</script>`, `\" onmouseover=\"alert(1)`, "'&<>\n\u2028\u2029"} {
		t.Run(value, func(t *testing.T) {
			response := httptest.NewRecorder()
			renderToken(response, tokenData{AccessToken: value, PublicMasterURL: value})
			if response.Header().Get("Content-Security-Policy") != assets.ContentSecurityPolicy {
				t.Fatal("missing CSP")
			}

			body := response.Body.String()
			for _, attribute := range []string{"data-access-token", "data-server"} {
				match := regexp.MustCompile(attribute + `="([^"]*)"`).FindStringSubmatch(body)
				if len(match) != 2 || html.UnescapeString(match[1]) != value {
					t.Fatal("dynamic data did not round trip")
				}
			}
			scripts := regexp.MustCompile(`<script\b([^>]*)>([\s\S]*?)</script>`).FindAllStringSubmatch(body, -1)
			if len(scripts) != 1 || strings.TrimSpace(scripts[0][2]) != "" || !strings.Contains(scripts[0][1], `src="`+assets.Prefix+`token.js"`) {
				t.Fatal("inline or dynamic script")
			}
			if regexp.MustCompile(`(?i)<[^>]+\s+on[a-z]+\s*=`).MatchString(regexp.MustCompile(`"[^"]*"|'[^']*'`).ReplaceAllString(body, `""`)) {
				t.Fatal("injected event attribute")
			}

		})
	}
}

func TestTokenErrorsAndForm(t *testing.T) {
	for _, render := range []func(*httptest.ResponseRecorder){
		func(response *httptest.ResponseRecorder) {
			renderToken(response, tokenData{sharedData: sharedData{Error: "<script>placeholder</script>"}})
		},
		func(response *httptest.ResponseRecorder) {
			renderForm(response, formData{Code: `\"<script>placeholder</script>`, CSRF: "placeholder"})
		},
	} {
		response := httptest.NewRecorder()
		render(response)
		if response.Header().Get("Content-Security-Policy") != assets.ContentSecurityPolicy || strings.Contains(response.Body.String(), "<script>") {
			t.Fatal("unsafe form/error response")
		}
	}
}
