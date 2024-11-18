# vanityurl

[![License][license.icon]][license.page]
[![CI][ci.icon]][ci.page]
[![Coverage][coverage.icon]][coverage.page]
[![Report][report.icon]][report.page]
[![Documentation][docs.icon]][docs.page]
[![Release][release.icon]][release.page]

`vanityurl` is a simple server that allows to set custom import paths for your Go packages.
It can be embedded to your Go projects as a library or you can deploy it as separate server.

## Server

### Installation

#### Download binary

Download binary from [releases][release.page].

#### Docker

See `wamod/vanityurl` image on [Docker Hub][docker.page].

#### Go Install

```sh
go install go.wamod.dev/vanityurl/cmd/vanityurl@latest
```

### Configuration

```yml
# ./vanityurl.yml

host: go.example.dev # (optional)
port: 8080           # (optional)
cache_age: 24h       # (optional)

packages:
  - path: /foo
    repository_url: https://github.com/example/foo
    # vcs: git    (auto-detected for Github, Gitlab, Bitbucket)
    # display: "" (auto-detected for Github, Gitlab, Bitbucket)
```

### Running

#### Command

```sh
vanityurl -config ./vanityurl.yml
```

#### Docker

```sh
docker run -it --rm \
    -v ${PWD}/vanityurl.yml:/etc/vanityurl/config.yml \
    -p 8080:8080 \
    wamod/vanityurl
```

## Library

`vanityurl` can also be used as library. For more details see [pkg.go.dev][docs.page].

### Installation 

To use this library in your Go project, you can import it using:

```go
import "go.wamod.dev/vanityurl"
```

### Example

```go
// Create a basic package resolver
resolver, err := vanityurl.NewResolver(
	vanityurl.Package{
		Path:          "/foo",
		RepositoryURL: "https://github.com/example/foo",
	},
	vanityurl.Package{
		Path:          "/bar",
		RepositoryURL: "https://github.com/example/bar",
	},
)
if err != nil {
	panic(err)
}

// Create a new vanity url server
server := vanityurl.NewServer(resolver, &vanityurl.ServerOptions{
	Host:     "go.example.dev",
	CacheAge: 24 * time.Hour,
})

http.ListenAndServe(":8080", server)
```


## Contributing

Thank you for your interest in contributing to the `vanityurl`! We welcome and appreciate any contributions, whether they be bug reports, feature requests, or code changes.

If you've found a bug, please [create an issue][issue.page] describing the problem, including any relevant error messages and a minimal reproduction of the issue.

## License

`vanityurl` is licensed under the [MIT License][license.page].

[issue.page]:    https://github.com/wamod-dev/vanityurl/issues/new/choose
[license.icon]:  https://img.shields.io/badge/license-MIT-green.svg
[license.page]:  https://github.com/wamod-dev/vanityurl/blob/main/LICENSE
[ci.icon]:       https://github.com/wamod-dev/vanityurl/actions/workflows/go.yml/badge.svg
[ci.page]:       https://github.com/wamod-dev/vanityurl/actions/workflows/go.yml
[coverage.icon]: https://codecov.io/gh/wamod-dev/vanityurl/graph/badge.svg?token=MHCY50YZA3
[coverage.page]: https://codecov.io/gh/wamod-dev/vanityurl
[report.icon]:   https://goreportcard.com/badge/go.wamod.dev/vanityurl
[report.page]:   https://goreportcard.com/report/go.wamod.dev/vanityurl
[docs.icon]:     https://godoc.org/go.wamod.dev/vanityurl?status.svg
[docs.page]:     http://pkg.go.dev/go.wamod.dev/vanityurl
[release.icon]:  https://img.shields.io/github/release/wamod-dev/vanityurl.svg
[release.page]:  https://github.com/wamod-dev/vanityurl/releases/latest
[docker.page]:   https://hub.docker.com/r/wamod/vanityurl