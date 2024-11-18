package vanityurl

import (
	"fmt"
)

var (
	ErrPackageNotFound = fmt.Errorf("vanityurl: package not found")
	ErrInvalidPackage  = fmt.Errorf("vanityurl: invalid package")
	ErrInvalidVCS      = fmt.Errorf("vanityurl: invalid vcs")
)
