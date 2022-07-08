package plugin

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
)

func TestPlugin_uploadPackage(t *testing.T) {
	t.Run("Creation of package was successful", func(t *testing.T) {
		zipContent := "not really a zip"
		owner := "gitea-owner"
		version := "2.0.0"
		giteaUser := "gitea-user"
		giteaToken := "1234"

		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			username, password, _ := r.BasicAuth()
			if username != giteaUser || password != giteaToken {
				t.Error("Should include username and password provided by the config as basic auth header")
			}

			body, _ := io.ReadAll(r.Body)
			if string(body) != zipContent {
				t.Error("Should include zip archive in http body")
			}

			expectedUrl := fmt.Sprintf("/api/packages/%s/composer", owner)
			if r.URL.Path != expectedUrl {
				t.Errorf("requested url should be \"%s\", but got %s", expectedUrl, r.URL.Path)
			}

			if r.URL.Query().Get("version") != version {
				t.Errorf("request parameters should include version = %s, got %s", version, r.URL.Query().Get("version"))
			}

			w.WriteHeader(http.StatusCreated)
		}))
		defer ts.Close()

		p := Plugin{config: Config{
			Owner:      owner,
			Version:    version,
			GiteaURL:   ts.URL,
			GiteaUser:  giteaUser,
			GiteaToken: giteaToken,
		}}

		err := p.uploadPackage(strings.NewReader(zipContent))

		if err != nil {
			t.Errorf("should not return error, got %v", err)
		}
	})

	t.Run("Creation of package failed", func(t *testing.T) {
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusBadRequest)
		}))
		defer ts.Close()

		p := Plugin{config: Config{
			GiteaURL: ts.URL,
		}}

		err := p.uploadPackage(strings.NewReader(""))

		if err != ErrPackageExists {
			t.Errorf("Should return an error, got %v", err)
		}
	})

	t.Run("Sets correct authorization header if an oauth token is used", func(t *testing.T) {
		oauthToken := "a-valid-o-auth-token-returned-by-netrc"
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Header.Get("Authorization") != "Bearer "+oauthToken {
				t.Errorf("Should use bearer authentication if user is signed in via Oauth")
			}
			w.WriteHeader(http.StatusBadRequest)
		}))
		defer ts.Close()

		p := Plugin{config: Config{
			GiteaURL:   ts.URL,
			GiteaUser:  oauthToken,
			GiteaToken: "x-oauth-basic",
		}}

		err := p.uploadPackage(bytes.NewReader([]byte("")))

		if err != ErrPackageExists {
			t.Errorf("Should return an error, got %v", err)
		}
	})
}

func TestPlugin_validFiles(t *testing.T) {
	badDirs := []string{
		".env/foo",
		"vendor/composer",
		"node_modules",
		"node_modules/bin",
		"src/Resources/frontend/node_modules/bin",
		"src/vendor/bin",
	}

	t.Cleanup(func() {
		for _, p := range badDirs {
			_ = os.Remove(p)
		}
	})

	for _, p := range badDirs {
		_ = os.MkdirAll(p, 0700)
	}

	p := Plugin{Config{}}
	files, err := p.validFiles()

	if err != nil {
		t.Error("should not return error")
	}

	for _, file := range files {
		for _, badPath := range badDirs {
			if strings.Contains(file, badPath) {
				t.Errorf("excluded file %s should not be included in valid files", file)
			}
		}
	}
}
