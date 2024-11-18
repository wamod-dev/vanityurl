package vanityurl

import (
	"cmp"
	"errors"
	"fmt"
	"net/http"
	"time"
)

var defaultServerOptions = &ServerOptions{ //nolint:gochecknoglobals
	CacheAge: time.Hour * 24,
}

// ServerOptions for additional configuration.
type ServerOptions struct {
	// Host for Go packages. If empty then uses Host value from [http.Request].
	Host string
	// CacheAge for response Cache-Control header. Default is 24h.
	CacheAge time.Duration
}

// Server for Go package vanity urls that implements [http.Handler]
type Server struct {
	host     string
	cacheAge time.Duration

	resolver Resolver
}

// NewServer creates a new [Server] to serve Go vanity url endpoints.
func NewServer(resolver Resolver, opts *ServerOptions) *Server {
	if opts == nil {
		opts = defaultServerOptions
	}

	if opts.CacheAge <= 0 {
		opts.CacheAge = defaultServerOptions.CacheAge
	}

	return &Server{
		host:     opts.Host,
		cacheAge: opts.CacheAge,
		resolver: resolver,
	}
}

// Host returns host used by server.
func (srv *Server) Host() string {
	return srv.host
}

// CacheAge returns cache-age used for HTTP response Cache-Control header.
func (srv *Server) CacheAge() time.Duration {
	return srv.cacheAge
}

// Resolver returns [Resolver] server is using.
func (srv *Server) Resolver() Resolver {
	return srv.resolver
}

// ServeHTTP implementation of [http.Handler].
func (srv *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	pkg, err := srv.resolver.ResolvePackage(r.Context(), r.URL.Path)
	if errors.Is(err, ErrPackageNotFound) {
		http.Error(w, "Package not found", http.StatusNotFound)

		return
	} else if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)

		return
	}

	w.Header().Add("Content-Type", "text/html; charset=utf-8")
	w.Header().Add("Cache-Control", fmt.Sprintf("public, max-age=%d", srv.cacheAge/time.Second))

	var subpath string

	if len(pkg.Path) < len(r.URL.Path) {
		subpath = r.URL.Path[len(pkg.Path)+1:]
	}

	_ = pkg.RenderDocument(w, cmp.Or(srv.host, r.Host), subpath)
}
