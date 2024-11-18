package vanityurl_test

import (
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strings"
	"testing"
	"time"

	"go.wamod.dev/vanityurl"
)

func TestNewServer(t *testing.T) {
	tt := []struct {
		name     string
		resolver vanityurl.Resolver
		opts     *vanityurl.ServerOptions

		wantHost     string
		wantCacheAge time.Duration
		wantResolver vanityurl.Resolver
	}{
		{
			name:         "nil_options",
			resolver:     failingResolver{},
			opts:         nil,
			wantHost:     "",
			wantCacheAge: 24 * time.Hour,
		},
		{
			name:     "with_host",
			resolver: failingResolver{},
			opts: &vanityurl.ServerOptions{
				Host: "go.foo.dev",
			},
			wantHost:     "go.foo.dev",
			wantCacheAge: 24 * time.Hour,
		},
		{
			name:     "with_cache_age",
			resolver: failingResolver{},
			opts: &vanityurl.ServerOptions{
				CacheAge: 123 * time.Second,
			},
			wantCacheAge: 123 * time.Second,
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			srv := vanityurl.NewServer(tc.resolver, tc.opts)

			if host := srv.Host(); tc.wantHost != host {
				t.Errorf("Server.Host() = %s; wantHost = %s", host, tc.wantHost)
			}

			if cacheAge := srv.CacheAge(); tc.wantCacheAge != cacheAge {
				t.Errorf("Server.CacheAge() = %v; wantCacheAge = %v", cacheAge, tc.wantCacheAge)
			}

			if resolver := srv.Resolver(); tc.resolver != resolver {
				t.Errorf("Server.Resolver() = %v; want same resolver = %v", resolver, tc.resolver)
			}
		})
	}
}

func TestServerHandler(t *testing.T) {
	tt := []struct {
		name        string
		srv         *vanityurl.Server
		path        string
		wantStatus  int
		wantHeaders map[string]string
		wantInBody  []string
	}{
		{
			name: "not_found",
			srv: vanityurl.NewServer(
				failingResolver{vanityurl.ErrPackageNotFound},
				nil,
			),
			path:       "/foo",
			wantStatus: http.StatusNotFound,
			wantHeaders: map[string]string{
				"Content-Type": "text/plain; charset=utf-8",
			},
			wantInBody: []string{"Package not found"},
		},
		{
			name: "resolver_error",
			srv: vanityurl.NewServer(
				failingResolver{os.ErrClosed},
				nil,
			),
			path:       "/foo",
			wantStatus: http.StatusInternalServerError,
			wantHeaders: map[string]string{
				"Content-Type": "text/plain; charset=utf-8",
			},
			wantInBody: []string{"Internal Server Error"},
		},
		{
			name: "found",
			srv: vanityurl.NewServer(
				mustResolver(t, vanityurl.Package{
					Path:          "/foo",
					VCS:           vanityurl.Git,
					Display:       "foo_display",
					RepositoryURL: "https://git.example.com/foo",
				}),
				&vanityurl.ServerOptions{
					Host:     "go.example.com",
					CacheAge: 60 * time.Second,
				},
			),
			path:       "/foo",
			wantStatus: http.StatusOK,
			wantHeaders: map[string]string{
				"Content-Type":  "text/html; charset=utf-8",
				"Cache-Control": "public, max-age=60",
			},
			wantInBody: []string{
				`<!DOCTYPE html>`,
				`<html>`,
				`<head>`,
				`<meta name="go-import" content="go.example.com/foo git https://git.example.com/foo">`,
				`<meta name="go-source" content="go.example.com/foo foo_display">`,
				`<meta http-equiv="refresh" content="0; url=https://pkg.go.dev/go.example.com/foo/">`,
				`</head>`,
				`<body>`,
				`Nothing to see here; <a href="https://pkg.go.dev/go.example.com/foo/">see the package on pkg.go.dev</a>.`,
				`</body>`,
				`</html>`,
			},
		},
		{
			name: "with_subpath",
			srv: vanityurl.NewServer(
				mustResolver(t, vanityurl.Package{
					Path:          "/foo",
					VCS:           vanityurl.Git,
					Display:       "foo_display",
					RepositoryURL: "https://git.example.com/foo",
				}),
				&vanityurl.ServerOptions{
					Host:     "go.example.com",
					CacheAge: 60 * time.Second,
				},
			),
			path:       "/foo/bar",
			wantStatus: http.StatusOK,
			wantHeaders: map[string]string{
				"Content-Type":  "text/html; charset=utf-8",
				"Cache-Control": "public, max-age=60",
			},
			wantInBody: []string{
				`<!DOCTYPE html>`,
				`<html>`,
				`<head>`,
				`<meta name="go-import" content="go.example.com/foo git https://git.example.com/foo">`,
				`<meta name="go-source" content="go.example.com/foo foo_display">`,
				`<meta http-equiv="refresh" content="0; url=https://pkg.go.dev/go.example.com/foo/bar">`,
				`</head>`,
				`<body>`,
				`Nothing to see here; <a href="https://pkg.go.dev/go.example.com/foo/bar">see the package on pkg.go.dev</a>.`,
				`</body>`,
				`</html>`,
			},
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			httpSrv := httptest.NewServer(tc.srv)

			path, err := url.JoinPath(httpSrv.URL, tc.path)
			if err != nil {
				t.Fatalf("got error while making request url: %v", err)
			}

			res, err := http.Get(path) //nolint:gosec
			if err != nil {
				t.Fatalf("got error while making request: %v", err)
			}

			if res.StatusCode != tc.wantStatus {
				t.Errorf("Server.ServeHTTP() status = %d; wantStatus = %d", res.StatusCode, tc.wantStatus)
			}

			for name, value := range tc.wantHeaders {
				if got := res.Header.Get(name); got != value {
					t.Errorf("Server.ServeHTTP() Header[%s] = %s; wantHeader = %s", name, got, value)
				}
			}

			defer res.Body.Close()

			body, err := io.ReadAll(res.Body)
			if err != nil {
				t.Fatalf("got error while reading response body: %v", err)
			}

			for _, want := range tc.wantInBody {
				if !strings.Contains(string(body), want) {
					t.Errorf("Server.ServeHTTP() body = %s; wantInBody = %s", string(body), want)
				}
			}
		})
	}
}
