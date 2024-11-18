package vanityurl

import (
	"cmp"
)

//nolint:gochecknoglobals
var (
	validVCS = map[VCS]struct{}{
		Git:       {},
		Svn:       {},
		Mercurial: {},
		Bazaar:    {},
	}
	vcsMapStrValues = map[VCS]string{
		Git:       "git",
		Svn:       "svn",
		Mercurial: "hg",
		Bazaar:    "bzr",
	}
	vcsMapStrKeys = map[string]VCS{
		"git": Git,
		"svn": Svn,
		"hg":  Mercurial,
		"bzr": Bazaar,
	}
)

// VCS Version Control System type. Allowed are [Git], [Svn], [Mercurial] and [Bazaar]
type VCS uint8

const (
	Git VCS = iota + 1
	Svn
	Mercurial
	Bazaar
)

// Valid check for VCS type.
func (vcs VCS) Valid() bool {
	_, ok := validVCS[vcs]

	return ok
}

// String representation for VCS types. Unknown types return "unspecified".
func (vcs VCS) String() string {
	return cmp.Or(vcsMapStrValues[vcs], "unspecified")
}

// ParseVCS type from string. Returns [ErrInvalidVCS] when failed.
func ParseVCS(str string) (VCS, error) {
	vcs, ok := vcsMapStrKeys[str]
	if !ok {
		return 0, ErrInvalidVCS
	}

	return vcs, nil
}
