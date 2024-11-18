package vanityurl_test

import (
	"strings"
	"testing"

	"go.wamod.dev/vanityurl"
)

func TestDetectVCS(t *testing.T) {
	tt := []struct {
		name    string
		repoURL string
		want    vanityurl.VCS
	}{
		{
			name:    "invalid_url",
			repoURL: "invalid\nurl",
			want:    0,
		},
		{
			name:    "empty",
			repoURL: "",
			want:    0,
		},
		{
			name:    "github",
			repoURL: "https://github.com/example/foo",
			want:    vanityurl.Git,
		},
		{
			name:    "gitlab",
			repoURL: "https://gitlab.com/example/foo",
			want:    vanityurl.Git,
		},
		{
			name:    "bitbucket",
			repoURL: "https://bitbucket.org/example/foo",
			want:    vanityurl.Git,
		},
	}
	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			if got := vanityurl.DetectVCS(tc.repoURL); got != tc.want {
				t.Errorf("DetectVCS() = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestDetectDisplay(t *testing.T) {
	tt := []struct {
		name    string
		repoURL string
		want    string
	}{
		{
			name:    "invalid_url",
			repoURL: "invalid\nurl",
			want:    "",
		},
		{
			name:    "empty",
			repoURL: "",
			want:    "",
		},
		{
			name:    "github",
			repoURL: "https://github.com/example/foo",
			want: strings.Join([]string{
				"https://github.com/example/foo",
				"https://github.com/example/foo/tree/master{/dir}",
				"https://github.com/example/foo/blob/master{/dir}/{file}#L{line}",
			}, " "),
		},
		{
			name:    "gitlab",
			repoURL: "https://gitlab.com/example/foo",
			want: strings.Join([]string{
				"https://gitlab.com/example/foo",
				"https://gitlab.com/example/foo/tree/master{/dir}",
				"https://gitlab.com/example/foo/blob/master{/dir}/{file}#L{line}",
			}, " "),
		},
		{
			name:    "bitbucket",
			repoURL: "https://bitbucket.org/example/foo",
			want: strings.Join([]string{
				"https://bitbucket.org/example/foo",
				"https://bitbucket.org/example/foo/src/default{/dir}",
				"https://bitbucket.org/example/foo/src/default{/dir}/{file}#{file}-{line}",
			}, " "),
		},
	}
	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			if got := vanityurl.DetectDisplay(tc.repoURL); got != tc.want {
				t.Errorf("DetectDisplay() = %v, want %v", got, tc.want)
			}
		})
	}
}
