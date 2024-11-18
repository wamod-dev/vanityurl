package vanityurl_test

import (
	"bytes"
	"io"
	"os"
	"strings"
	"testing"

	"go.wamod.dev/vanityurl"
)

type failWriter struct {
	err error
}

func (w failWriter) Write(_ []byte) (int, error) {
	return 0, w.err
}

func TestPackageRenderHead(t *testing.T) {
	tt := []struct {
		name     string
		writer   io.Writer
		pkg      vanityurl.Package
		host     string
		wantMeta []string
		wantErr  bool
	}{
		{
			name:   "simple",
			writer: bytes.NewBuffer(nil),
			host:   "go.example.com",
			pkg: vanityurl.Package{
				Path:          "/foo",
				VCS:           vanityurl.Git,
				Display:       "display",
				RepositoryURL: "https://git.repo.com",
			},
			wantMeta: []string{
				`<meta name="go-import" content="go.example.com/foo git https://git.repo.com">`,
				`<meta name="go-source" content="go.example.com/foo display">`,
			},
			wantErr: false,
		},
		{
			name:   "bad_writer",
			writer: failWriter{os.ErrClosed},
			host:   "go.example.com",
			pkg: vanityurl.Package{
				Path:          "/foo",
				VCS:           vanityurl.Git,
				Display:       "display",
				RepositoryURL: "https://git.repo.com",
			},
			wantErr: true,
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			buf := bytes.NewBuffer(nil)

			err := tc.pkg.RenderHead(io.MultiWriter(buf, tc.writer), tc.host)
			if tc.wantErr != (err != nil) {
				t.Errorf("Package.RenderHead() = %v; wantErr = %v", err, tc.wantErr)
			}

			output := buf.String()
			for _, meta := range tc.wantMeta {
				if !strings.Contains(output, meta) {
					t.Errorf("Package.RenderHead() expected meta in output = %s", meta)
				}
			}
		})
	}
}

func TestPackageRenderDocument(t *testing.T) {
	tt := []struct {
		name        string
		writer      io.Writer
		pkg         vanityurl.Package
		host        string
		subpath     string
		wantElement []string
		wantErr     bool
	}{
		{
			name:    "simple",
			writer:  bytes.NewBuffer(nil),
			host:    "go.example.com",
			subpath: "/bar",
			pkg: vanityurl.Package{
				Path:          "/foo",
				VCS:           vanityurl.Git,
				Display:       "display",
				RepositoryURL: "https://git.repo.com",
			},
			wantElement: []string{
				`<!DOCTYPE html>`,
				`<html>`,
				`<head>`,
				`<meta name="go-import" content="go.example.com/foo git https://git.repo.com">`,
				`<meta name="go-source" content="go.example.com/foo display">`,
				`<meta http-equiv="refresh" content="0; url=https://pkg.go.dev/go.example.com/foo//bar">`,
				`</head>`,
				`<body>`,
				`Nothing to see here; <a href="https://pkg.go.dev/go.example.com/foo//bar">see the package on pkg.go.dev</a>.`,
				`</body>`,
				`</html>`,
				`<!DOCTYPE html>`,
				`<html>`,
				`<head>`,
			},
			wantErr: false,
		},
		{
			name:    "bad_writer",
			writer:  failWriter{os.ErrClosed},
			host:    "go.example.com",
			subpath: "/bar",
			pkg: vanityurl.Package{
				Path:          "/foo",
				VCS:           vanityurl.Git,
				Display:       "display",
				RepositoryURL: "https://git.repo.com",
			},
			wantErr: true,
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			buf := bytes.NewBuffer(nil)

			err := tc.pkg.RenderDocument(io.MultiWriter(buf, tc.writer), tc.host, tc.subpath)
			if tc.wantErr != (err != nil) {
				t.Errorf("Package.RenderDocument() = %v; wantErr = %v", err, tc.wantErr)
			}

			output := buf.String()
			for _, meta := range tc.wantElement {
				if !strings.Contains(output, meta) {
					t.Errorf("Package.RenderDocument() expected element in output = %s", meta)
				}
			}
		})
	}
}

func TestPackageAdjustFields(t *testing.T) {
	tt := []struct {
		name    string
		pkg     vanityurl.Package
		want    vanityurl.Package
		wantErr bool
	}{
		{
			name: "no_change",
			pkg: vanityurl.Package{
				Path:          "/foo",
				VCS:           vanityurl.Git,
				Display:       "foo_display",
				RepositoryURL: "https://git.repo.com/foo",
			},
			want: vanityurl.Package{
				Path:          "/foo",
				VCS:           vanityurl.Git,
				Display:       "foo_display",
				RepositoryURL: "https://git.repo.com/foo",
			},
			wantErr: false,
		},
		{
			name: "detect_vcs",
			pkg: vanityurl.Package{
				Path:          "/foo",
				Display:       "foo_display",
				RepositoryURL: "https://github.com/foo",
			},
			want: vanityurl.Package{
				Path:          "/foo",
				VCS:           vanityurl.Git,
				Display:       "foo_display",
				RepositoryURL: "https://github.com/foo",
			},
			wantErr: false,
		},
		{
			name: "detect_display",
			pkg: vanityurl.Package{
				Path:          "/foo",
				VCS:           vanityurl.Git,
				RepositoryURL: "https://github.com/foo",
			},
			want: vanityurl.Package{
				Path:          "/foo",
				VCS:           vanityurl.Git,
				RepositoryURL: "https://github.com/foo",
				Display: strings.Join([]string{
					"https://github.com/foo",
					"https://github.com/foo/tree/master{/dir}",
					"https://github.com/foo/blob/master{/dir}/{file}#L{line}",
				}, " "),
			},
			wantErr: false,
		},
		{
			name: "add_path_prefix",
			pkg: vanityurl.Package{
				Path:          "foo",
				VCS:           vanityurl.Git,
				Display:       "foo_display",
				RepositoryURL: "https://git.repo.com/foo",
			},
			want: vanityurl.Package{
				Path:          "/foo",
				VCS:           vanityurl.Git,
				Display:       "foo_display",
				RepositoryURL: "https://git.repo.com/foo",
			},
			wantErr: false,
		},
		{
			name: "remove_path_suffix",
			pkg: vanityurl.Package{
				Path:          "/foo/",
				VCS:           vanityurl.Git,
				Display:       "foo_display",
				RepositoryURL: "https://git.repo.com/foo",
			},
			want: vanityurl.Package{
				Path:          "/foo",
				VCS:           vanityurl.Git,
				Display:       "foo_display",
				RepositoryURL: "https://git.repo.com/foo",
			},
			wantErr: false,
		},
		{
			name: "fail_repo_url_parse",
			pkg: vanityurl.Package{
				Path:          "/foo",
				VCS:           vanityurl.Git,
				Display:       "foo_display",
				RepositoryURL: "invalid\nurl",
			},
			wantErr: true,
		},
		{
			name: "invalid_repo_url_schema",
			pkg: vanityurl.Package{
				Path:          "/foo",
				VCS:           vanityurl.Git,
				Display:       "foo_display",
				RepositoryURL: "other://git.repo.com/foo",
			},
			wantErr: true,
		},
		{
			name: "repo_url_with_query",
			pkg: vanityurl.Package{
				Path:          "/foo",
				VCS:           vanityurl.Git,
				Display:       "foo_display",
				RepositoryURL: "https://git.repo.com/foo?bar",
			},
			wantErr: true,
		},
		{
			name: "invalid_vcs",
			pkg: vanityurl.Package{
				Path:          "/foo",
				VCS:           0,
				Display:       "foo_display",
				RepositoryURL: "https://git.repo.com/foo",
			},
			wantErr: true,
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			got, err := tc.pkg.AdjustFields()
			if tc.wantErr != (err != nil) {
				t.Errorf("Package.AdjustFields() error = %v; wantErr = %v", err, tc.wantErr)
			}

			if got != tc.want {
				t.Errorf("Package.AdjustFields() = %v; want = %v", got, tc.want)
			}
		})
	}
}
