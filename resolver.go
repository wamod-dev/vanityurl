package vanityurl

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"sort"
	"strings"
)

// Resolver type for packages.
type Resolver interface {
	// ResolvePackage for a given path.
	// Returns [ErrPackageNotFound] if package not found.
	ResolvePackage(ctx context.Context, path string) (Package, error)
}

type resolver struct {
	pset []Package
}

// NewResolver creates a static resolver with a given [Package] set.
func NewResolver(pset ...Package) (Resolver, error) {
	pathMap := map[string]struct{}{}
	for i, pkg := range pset {
		if _, ok := pathMap[pkg.Path]; ok {
			return nil, fmt.Errorf("%w: duplicate paths: %s", ErrInvalidPackage, pkg.Path)
		}

		pkg, err := pkg.AdjustFields()
		if err != nil {
			return nil, err
		}

		pathMap[pkg.Path] = struct{}{}
		pset[i] = pkg
	}

	slices.SortFunc(pset, func(a, b Package) int {
		return strings.Compare(a.Path, b.Path)
	})

	return &resolver{pset}, nil
}

func (r *resolver) ResolvePackage(_ context.Context, path string) (Package, error) {
	i := sort.Search(len(r.pset), func(i int) bool {
		return r.pset[i].Path >= path
	})

	if i < len(r.pset) && r.pset[i].Path == path {
		return r.pset[i], nil
	}

	if i > 0 && strings.HasPrefix(path, r.pset[i-1].Path+"/") {
		return r.pset[i-1], nil
	}

	var match *Package

	matchSubpathLen := len(path)

	for j := 0; j < i; j++ {
		if len(r.pset[j].Path) >= len(path) {
			continue
		}

		subpath := strings.TrimPrefix(path, r.pset[j].Path+"/")

		if len(subpath) < matchSubpathLen {
			matchSubpathLen = len(subpath)
			match = &r.pset[j]
		}
	}

	if match != nil {
		return *match, nil
	}

	return Package{}, ErrPackageNotFound
}

type multiResolver struct {
	rset []Resolver
}

// NewMultiResolver create a new resolver from multiple other resolvers.
func NewMultiResolver(rset ...Resolver) Resolver {
	return &multiResolver{rset}
}

func (r *multiResolver) ResolvePackage(ctx context.Context, path string) (Package, error) {
	for _, rr := range r.rset {
		pkg, err := rr.ResolvePackage(ctx, path)
		if errors.Is(err, ErrPackageNotFound) {
			continue
		} else if err != nil {
			return Package{}, err
		}

		return pkg, nil
	}

	return Package{}, ErrPackageNotFound
}
