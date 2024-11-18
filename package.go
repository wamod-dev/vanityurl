package vanityurl

import (
	"cmp"
	_ "embed"
	"fmt"
	"io"
	"net/url"
	"slices"
	"strings"
	"text/template"
)

const (
	pkgHeadName = "head"
	pkgDocName  = "document"
)

//nolint:gochecknoglobals
var (
	//go:embed package.gohtml
	pkgTmplRaw string

	pkgTmpl = template.Must(template.New("package").Parse(pkgTmplRaw))
)

// Package type for rendering Go vanity url HTML elements.
type Package struct {
	Path          string
	VCS           VCS
	Display       string
	RepositoryURL string
}

// RenderHead 'go-import' and 'go-source' HTML meta elements of the package
func (pkg Package) RenderHead(wr io.Writer, host string) error {
	return pkgTmpl.ExecuteTemplate(wr, pkgHeadName, struct {
		Package Package
		Host    string
	}{
		Package: pkg,
		Host:    host,
	})
}

// RenderDocument full HTML document of the package.
func (pkg Package) RenderDocument(wr io.Writer, host, subpath string) error {
	return pkgTmpl.ExecuteTemplate(wr, pkgDocName, struct {
		Package Package
		Host    string
		Subpath string
	}{
		Package: pkg,
		Host:    host,
		Subpath: subpath,
	})
}

// AdjustFields to cleanup existing fields and detect missing vcs and display.
// Returns [ErrInvalidPackage] if package is configured incorrectly
func (pkg Package) AdjustFields() (Package, error) {
	// Cleanup path field
	pkg.Path = strings.TrimSpace(pkg.Path)
	pkg.Path = strings.TrimSuffix(pkg.Path, "/")

	if !strings.HasPrefix(pkg.Path, "/") {
		pkg.Path = "/" + pkg.Path
	}

	// Cleanup repository url field
	parsedURL, err := url.Parse(strings.TrimSpace(pkg.RepositoryURL))
	if err != nil {
		return Package{}, fmt.Errorf("%w: invalid repository url: %w", ErrInvalidPackage, err)
	} else if len(parsedURL.Query()) > 0 {
		return Package{}, fmt.Errorf("%w: repository url contains query parameters: %s", ErrInvalidPackage, parsedURL.RawQuery)
	} else if !slices.Contains([]string{"https", "http"}, parsedURL.Scheme) {
		return Package{}, fmt.Errorf("%w: invalid repository url schema: %s", ErrInvalidPackage, parsedURL.Scheme)
	}

	pkg.RepositoryURL = fmt.Sprintf("%s://%s%s", parsedURL.Scheme, parsedURL.Host, parsedURL.Path)

	// Detect VCS if missing
	pkg.VCS = cmp.Or(
		pkg.VCS,
		DetectVCS(pkg.RepositoryURL),
	)

	// Fail if VCS still missing
	if !pkg.VCS.Valid() {
		return Package{}, fmt.Errorf("%w: could not detect VCS", ErrInvalidPackage)
	}

	// Detect display field if missing
	pkg.Display = cmp.Or(
		pkg.Display,
		DetectDisplay(pkg.RepositoryURL),
	)

	return pkg, nil
}
