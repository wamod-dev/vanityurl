package vanityurl_test

import (
	"context"
	"os"
	"testing"

	"go.wamod.dev/vanityurl"
)

func TestNewResolver(t *testing.T) {
	tt := []struct {
		name           string
		pset           []vanityurl.Package
		resolvePath    string
		wantErr        bool
		wantResolvePkg vanityurl.Package
		wantResolveErr bool
	}{
		{
			name:           "empty",
			pset:           []vanityurl.Package{},
			resolvePath:    "/baz",
			wantResolveErr: true,
		},
		{
			name: "not_found",
			pset: []vanityurl.Package{
				{
					Path:          "/foo",
					VCS:           vanityurl.Git,
					Display:       "foo_display",
					RepositoryURL: "https://git.example.com/foo",
				},
				{
					Path:          "/bar",
					VCS:           vanityurl.Git,
					Display:       "bar_display",
					RepositoryURL: "https://git.example.com/bar",
				},
			},
			resolvePath:    "/baz",
			wantResolveErr: true,
		},
		{
			name: "root_package",
			pset: []vanityurl.Package{
				{
					Path:          "/foo",
					VCS:           vanityurl.Git,
					Display:       "foo_display",
					RepositoryURL: "https://git.example.com/foo",
				},
				{
					Path:          "/bar",
					VCS:           vanityurl.Git,
					Display:       "bar_display",
					RepositoryURL: "https://git.example.com/bar",
				},
			},
			resolvePath: "/foo",
			wantResolvePkg: vanityurl.Package{
				Path:          "/foo",
				VCS:           vanityurl.Git,
				Display:       "foo_display",
				RepositoryURL: "https://git.example.com/foo",
			},
		},
		{
			name: "sub_package",
			pset: []vanityurl.Package{
				{
					Path:          "/foo",
					VCS:           vanityurl.Git,
					Display:       "foo_display",
					RepositoryURL: "https://git.example.com/foo",
				},
				{
					Path:          "/bar",
					VCS:           vanityurl.Git,
					Display:       "bar_display",
					RepositoryURL: "https://git.example.com/bar",
				},
			},
			resolvePath: "/bar/baz",
			wantResolvePkg: vanityurl.Package{
				Path:          "/bar",
				VCS:           vanityurl.Git,
				Display:       "bar_display",
				RepositoryURL: "https://git.example.com/bar",
			},
		},
		{
			name: "sub_of_root",
			pset: []vanityurl.Package{
				{
					Path:          "/foo",
					VCS:           vanityurl.Git,
					Display:       "foo_display",
					RepositoryURL: "https://git.example.com/foo",
				},
				{
					Path:          "/foo/bar",
					VCS:           vanityurl.Git,
					Display:       "foo_display",
					RepositoryURL: "https://git.example.com/foo",
				},
			},
			resolvePath: "/foo/baz",
			wantResolvePkg: vanityurl.Package{
				Path:          "/foo",
				VCS:           vanityurl.Git,
				Display:       "foo_display",
				RepositoryURL: "https://git.example.com/foo",
			},
		},
		{
			name: "duplicate",
			pset: []vanityurl.Package{
				{
					Path:          "/foo",
					VCS:           vanityurl.Git,
					Display:       "foo_display",
					RepositoryURL: "https://git.example.com/foo",
				},
				{
					Path:          "/foo",
					VCS:           vanityurl.Git,
					Display:       "foo_display",
					RepositoryURL: "https://git.example.com/foo",
				},
			},
			wantErr: true,
		},
		{
			name: "invalid_package",
			pset: []vanityurl.Package{
				{
					Path:          "/foo",
					RepositoryURL: "none://git.example.com/foo",
				},
			},
			wantErr: true,
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			resolver, err := vanityurl.NewResolver(tc.pset...)
			if tc.wantErr != (err != nil) {
				t.Fatalf("NewResolver() = %v; wantErr = %v", err, tc.wantErr)
			}

			if tc.wantErr {
				return
			}

			pkg, err := resolver.ResolvePackage(context.Background(), tc.resolvePath)
			if tc.wantResolveErr != (err != nil) {
				t.Errorf("Resolver.ResolvePackage() = %v; wantResolveErr = %v", err, tc.wantResolveErr)
			}

			if pkg != tc.wantResolvePkg {
				t.Errorf("Resolver.ResolvePackage() = %v; wantResolvePkg = %v", pkg, tc.wantResolvePkg)
			}
		})
	}
}

func mustResolver(t testing.TB, pset ...vanityurl.Package) vanityurl.Resolver {
	resolver, err := vanityurl.NewResolver(pset...)
	if err != nil {
		t.Fatalf("got error while creating resolver: %v", err)
	}

	return resolver
}

type failingResolver struct {
	err error
}

func (resolver failingResolver) ResolvePackage(_ context.Context, _ string) (vanityurl.Package, error) {
	return vanityurl.Package{}, resolver.err
}

func TestNewMultiResolver(t *testing.T) {
	tt := []struct {
		name    string
		rset    []vanityurl.Resolver
		path    string
		want    vanityurl.Package
		wantErr bool
	}{
		{
			name:    "empty",
			rset:    []vanityurl.Resolver{},
			path:    "/foo",
			wantErr: true,
		},
		{
			name: "first",
			rset: []vanityurl.Resolver{
				mustResolver(t,
					vanityurl.Package{
						Path:          "/foo",
						VCS:           vanityurl.Git,
						Display:       "foo_display",
						RepositoryURL: "https://git.example.com/foo",
					},
				),
				failingResolver{os.ErrInvalid},
			},
			path: "/foo",
			want: vanityurl.Package{
				Path:          "/foo",
				VCS:           vanityurl.Git,
				Display:       "foo_display",
				RepositoryURL: "https://git.example.com/foo",
			},
		},
		{
			name: "second",
			rset: []vanityurl.Resolver{
				failingResolver{vanityurl.ErrPackageNotFound},
				mustResolver(t,
					vanityurl.Package{
						Path:          "/foo",
						VCS:           vanityurl.Git,
						Display:       "foo_display",
						RepositoryURL: "https://git.example.com/foo",
					},
				),
			},
			path: "/foo",
			want: vanityurl.Package{
				Path:          "/foo",
				VCS:           vanityurl.Git,
				Display:       "foo_display",
				RepositoryURL: "https://git.example.com/foo",
			},
		},
		{
			name: "fail_before_second",
			rset: []vanityurl.Resolver{
				failingResolver{os.ErrInvalid},
				mustResolver(t,
					vanityurl.Package{
						Path:          "/foo",
						VCS:           vanityurl.Git,
						Display:       "foo_display",
						RepositoryURL: "https://git.example.com/foo",
					},
				),
			},
			path:    "/foo",
			wantErr: true,
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			resolver := vanityurl.NewMultiResolver(tc.rset...)

			pkg, err := resolver.ResolvePackage(context.Background(), tc.path)
			if tc.wantErr != (err != nil) {
				t.Errorf("MultiResolver.ResolvePackage() = %v; wantErr = %v", err, tc.wantErr)
			}

			if tc.want != pkg {
				t.Errorf("MultiResolver.ResolvePackage() = %v; want = %v", pkg, tc.want)
			}
		})
	}
}
