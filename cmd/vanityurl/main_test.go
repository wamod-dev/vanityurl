package main

import (
	"bytes"
	"fmt"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"

	"go.wamod.dev/vanityurl"
)

func getFreePort() (port int, err error) {
	var a *net.TCPAddr

	if a, err = net.ResolveTCPAddr("tcp", "localhost:0"); err == nil {
		var l *net.TCPListener

		if l, err = net.ListenTCP("tcp", a); err == nil {
			defer l.Close()

			return l.Addr().(*net.TCPAddr).Port, nil
		}
	}

	return
}

func Test_run(t *testing.T) {
	tmpDir := t.TempDir()

	port, err := getFreePort()
	if err != nil {
		t.Fatalf("failed to get free port for listener")
	}

	cfgName := filepath.Join(tmpDir, "vanityurl.yml")
	cfgContents := strings.Join([]string{
		`host: go.foo.dev`,
		fmt.Sprintf("port: %d", port),
		`cache_age: 123s`,
		`packages:`,
		`  - path: /bar`,
		`    repository_url: https://github.com/foo/bar`,
	}, "\n")

	err = os.WriteFile(cfgName, []byte(cfgContents), 0o600)
	if err != nil {
		t.Fatalf("failed to create temp config: %v", err)
	}

	sigch := make(chan os.Signal)
	errch := make(chan error)
	donech := make(chan struct{})

	output := bytes.NewBuffer(nil)

	go func() {
		err := run([]string{"-config", cfgName}, output, sigch)
		if err != nil {
			errch <- err

			return
		}

		donech <- struct{}{}
	}()

	time.Sleep(100 * time.Millisecond)

	res, err := http.Get(fmt.Sprintf("http://localhost:%d/bar", port))
	if err != nil {
		t.Fatalf("got error from HTTP handler: %v", err)
	}

	defer res.Body.Close()

	if want := 200; res.StatusCode != want {
		t.Fatalf("response.StatusCode = %d; want %d", res.StatusCode, want)
	}

	sigch <- os.Kill

	select {
	case err := <-errch:
		t.Fatalf("run() = got error = %v", err)
	case <-donech:
	}

	logs := output.String()

	wantLogs := []string{
		`Starting server`,
		`Loaded config`,
		`Configuring package`,
		`Creating resolver`,
		`Listening`,
		`Closing server`,
	}

	for _, msg := range wantLogs {
		if !strings.Contains(logs, msg) {
			t.Errorf("log message not found; expected msg = %s", msg)
		}
	}
}

func Test_parseConfig(t *testing.T) {
	tmpDir := t.TempDir()

	tt := []struct {
		name       string
		args       []string
		files      map[string]string
		before     func(t testing.TB)
		wantErr    bool
		wantConfig yamlConfig
	}{
		{
			name: "empty",
			args: []string{"-config", filepath.Join(tmpDir, "empty.yml")},
			files: map[string]string{
				"empty.yml": "{}",
			},
			wantErr: false,
			wantConfig: yamlConfig{
				Host:     defaultHost,
				Port:     defaultPort,
				CacheAge: defaultCacheAge,
			},
		},
		{
			name: "set_by_env",
			args: []string{},
			before: func(t testing.TB) {
				t.Setenv("VANITYURL_CONFIG", filepath.Join(tmpDir, "set_by_env.yml"))
			},
			files: map[string]string{
				"set_by_env.yml": strings.Join([]string{
					`host: go.env.dev`,
					`port: 1234`,
					`cache_age: 123s`,
				}, "\n"),
			},
			wantConfig: yamlConfig{
				Host:     "go.env.dev",
				Port:     1234,
				CacheAge: 123 * time.Second,
			},
		},
		{
			name: "full",
			args: []string{"-config", filepath.Join(tmpDir, "full.yml")},
			files: map[string]string{
				"full.yml": strings.Join([]string{
					`host: go.full.dev`,
					`port: 1234`,
					`cache_age: 123s`,
					`packages:`,
					`  - path: /foo`,
					`    repository_url: https://git.example.dev/example-dev/foo`,
					`    vcs: git`,
					`    display: foo_display`,
				}, "\n"),
			},
			wantConfig: yamlConfig{
				Host:     "go.full.dev",
				Port:     1234,
				CacheAge: 123 * time.Second,
				Packages: []yamlPackage{
					{
						Path:          "/foo",
						RepositoryURL: "https://git.example.dev/example-dev/foo",
						Display:       "foo_display",
						VCS: yamlVCS{
							Value: vanityurl.Git,
						},
					},
				},
			},
		},
		{
			name: "malformed",
			args: []string{"-config", filepath.Join(tmpDir, "malformed.yml")},
			files: map[string]string{
				"malformed.yml": "!{}",
			},
			wantErr: true,
		},
		{
			name: "wrong_flag",
			args: []string{"-config-file", filepath.Join(tmpDir, "wrong_flag.yml")},
			files: map[string]string{
				"wrong_flag.yml": "",
			},
			wantErr: true,
		},
		{
			name:    "missing_file",
			args:    []string{"-config", filepath.Join(tmpDir, "missing.yml")},
			files:   map[string]string{},
			wantErr: true,
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			if tc.before != nil {
				tc.before(t)
			}

			for fn, contents := range tc.files {
				err := os.WriteFile(filepath.Join(tmpDir, fn), []byte(contents), 0o600)
				if err != nil {
					t.Fatalf("got error while preparing files: %v", err)
				}
			}

			got, err := parseConfig(tc.args)

			if tc.wantErr != (err != nil) {
				t.Errorf("parseConfig() = %v; wantErr = %v", err, tc.wantErr)
			}

			if !reflect.DeepEqual(tc.wantConfig, got) {
				t.Errorf("parseConfig() = %v; wantConfig = %v", got, tc.wantConfig)
			}
		})
	}
}

func Test_parseFlag(t *testing.T) {
	tt := []struct {
		name      string
		args      []string
		cfgName   stringValue
		wantValue stringValue
		wantErr   bool
	}{
		{
			name: "default",
			args: []string{},
			cfgName: stringValue{
				value: "foo.yml",
			},
			wantErr: false,
			wantValue: stringValue{
				value: "foo.yml",
				set:   false,
			},
		},
		{
			name: "override",
			args: []string{"-config", "bar.yml"},
			cfgName: stringValue{
				value: "foo.yml",
			},
			wantErr: false,
			wantValue: stringValue{
				value: "bar.yml",
				set:   true,
			},
		},
		{
			name: "unknown_flag",
			args: []string{"-unknown"},
			cfgName: stringValue{
				value: "foo.yml",
			},
			wantErr: true,
			wantValue: stringValue{
				value: "foo.yml",
				set:   false,
			},
		},
	}
	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			err := parseFlags(tc.args, &tc.cfgName)
			if tc.wantErr != (err != nil) {
				t.Errorf("parseFlag() = %v; wantErr = %v", err, tc.wantErr)
			}

			if tc.wantValue != tc.cfgName {
				t.Errorf("cfgName = %v; want = %v", tc.cfgName, tc.wantValue)
			}
		})
	}
}

func Test_parseEnv(t *testing.T) {
	cfgName := stringValue{
		value: "foo.yml",
		set:   false,
	}

	want := stringValue{
		value: "foo.yml",
		set:   false,
	}

	testFunc := func(t *testing.T) {
		parseEnv(&cfgName)

		if cfgName != want {
			t.Errorf("cfgName = %v; want = %v", cfgName, want)
		}
	}

	t.Run("use_default", testFunc)

	t.Setenv("VANITYURL_CONFIG", "bar.yml")

	want = stringValue{
		value: "bar.yml",
		set:   true,
	}

	t.Run("use_env", testFunc)
}
