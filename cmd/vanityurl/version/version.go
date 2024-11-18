package version

import "runtime/debug"

var version = ""

func init() { //nolint:gochecknoinits
	info, ok := debug.ReadBuildInfo()
	if ok && version == "" {
		version = info.Main.Version
	}
}

// Version of vanityurl package
func Version() string {
	return version
}
