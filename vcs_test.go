package vanityurl_test

import (
	"testing"

	"go.wamod.dev/vanityurl"
)

func TestParseVCS(t *testing.T) {
	tt := []struct {
		name    string
		str     string
		want    vanityurl.VCS
		wantErr bool
	}{
		{
			name:    "empty",
			str:     "",
			wantErr: true,
		},
		{
			name:    "unknown",
			str:     "unknown",
			wantErr: true,
		},
		{
			name: "git",
			str:  "git",
			want: vanityurl.Git,
		},
		{
			name: "svn",
			str:  "svn",
			want: vanityurl.Svn,
		},
		{
			name: "mercurial",
			str:  "hg",
			want: vanityurl.Mercurial,
		},
		{
			name: "bazaar",
			str:  "bzr",
			want: vanityurl.Bazaar,
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			got, err := vanityurl.ParseVCS(tc.str)
			if tc.wantErr != (err != nil) {
				t.Errorf("ParseVCS() = %v; wantErr = %v", err, tc.wantErr)
			}

			if got != tc.want {
				t.Errorf("ParseVCS() = %s; want = %s", got, tc.want)
			}
		})
	}
}
