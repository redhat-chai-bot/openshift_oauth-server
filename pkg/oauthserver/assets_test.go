package oauthserver_test

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/openshift/oauth-server/pkg/server/assets"
	"github.com/openshift/oauth-server/pkg/server/headers"
	"k8s.io/apiserver/pkg/authentication/request/anonymous"
	"k8s.io/apiserver/pkg/authorization/authorizer"
	"k8s.io/apiserver/pkg/authorization/path"
)

func TestOAuthAssetsExistingAnonymousPathRule(t *testing.T) {
	pathAuthorizer, err := path.NewAuthorizer([]string{"/oauth/*"})
	if err != nil {
		t.Fatal(err)
	}
	for _, script := range []string{"theme.js", "login.js", "token.js"} {
		request := httptest.NewRequest(http.MethodGet, assets.Prefix+script, nil)
		decision, _, err := pathAuthorizer.Authorize(context.Background(), authorizer.AttributesRecord{Path: request.URL.Path})
		if err != nil || decision != authorizer.DecisionAllow {
			t.Fatal("existing OAuth path exception does not allow asset")
		}
		response, allowed, err := anonymous.NewAuthenticator(nil).AuthenticateRequest(request)
		if err != nil || !allowed || response.User.GetName() != "system:anonymous" {
			t.Fatal("anonymous authentication component rejected asset")
		}
	}
	decision, _, err := pathAuthorizer.Authorize(context.Background(), authorizer.AttributesRecord{Path: "/oauth-assets/v1/theme.js"})
	if err != nil || decision == authorizer.DecisionAllow {
		t.Fatal("OAuth exception broadened")
	}
}

func TestWithOAuthAssetsAnonymousMux(t *testing.T) {
	server := httptest.NewServer(headers.WithStandardHeaders(setup(t)))
	defer server.Close()
	for _, test := range []struct {
		name   string
		status int
	}{
		{"theme.js", http.StatusOK}, {"login.js", http.StatusOK}, {"token.js", http.StatusOK}, {"absent.js", http.StatusNotFound},
	} {
		t.Run(test.name, func(t *testing.T) {
			response, err := server.Client().Get(server.URL + assets.Prefix + test.name)
			if err != nil {
				t.Fatal(err)
			}
			defer response.Body.Close()
			body, err := io.ReadAll(response.Body)
			if err != nil {
				t.Fatal(err)
			}
			if response.StatusCode != test.status {
				t.Fatalf("status %d, want %d", response.StatusCode, test.status)
			}
			if response.Header.Get("X-XSS-Protection") != "1; mode=block" {
				t.Fatal("legacy header changed")
			}
			if !strings.Contains(response.Header.Get("Cache-Control"), "no-store") {
				t.Fatal("cache boundary changed")
			}
			if test.status == http.StatusOK && (len(body) == 0 || response.Header.Get("Content-Type") != "text/javascript; charset=utf-8") {
				t.Fatal("empty asset or incorrect MIME")
			}
		})
	}
}
