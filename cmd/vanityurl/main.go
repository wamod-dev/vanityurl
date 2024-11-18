package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"gopkg.in/yaml.v3"

	"go.wamod.dev/vanityurl"
	"go.wamod.dev/vanityurl/cmd/vanityurl/version"
)

const (
	flagConfigVar     = "config"
	envConfigVar      = "VANITYURL_CONFIG"
	defaultConfigFile = "vanityurl.yml"
	defaultHost       = "0.0.0.0"
	defaultPort       = 8080
	defaultCacheAge   = 24 * time.Hour
)

func main() {
	sigChan := make(chan os.Signal, 1)

	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	if err := run(os.Args, os.Stderr, sigChan); err != nil {
		os.Exit(1)
	}
}

func run(args []string, stderr io.Writer, sigChan <-chan os.Signal) error {
	logger := slog.New(slog.NewTextHandler(stderr, nil))
	logger.Info("Starting server", "version", version.Version())

	cfg, err := parseConfig(args)
	if err != nil {
		logger.Error("Failed to parse config", "err", err)

		return err
	}

	logger.Info("Loaded config", slog.Group("config",
		"host", cfg.Host,
		"port", cfg.Port,
		"cache_age", cfg.CacheAge,
		"packages_total", len(cfg.Packages),
	))

	packages := make([]vanityurl.Package, len(cfg.Packages))

	for i, pkg := range cfg.Packages {
		logger.Info("Configuring package", slog.Group("package",
			"id", i,
			"path", pkg.Path,
			"display", pkg.Display,
			"vcs", pkg.VCS.String(),
			"repository_url", pkg.RepositoryURL,
		))

		packages[i] = vanityurl.Package{
			Path:          pkg.Path,
			Display:       pkg.Display,
			VCS:           pkg.VCS.Value,
			RepositoryURL: pkg.RepositoryURL,
		}
	}

	logger.Info("Creating resolver")

	resolver, err := vanityurl.NewResolver(packages...)
	if err != nil {
		logger.Error("Failed to create resolver", "err", err)

		return err
	}

	srv := http.Server{
		Addr: fmt.Sprintf(":%d", cfg.Port),
		Handler: vanityurl.NewServer(resolver, &vanityurl.ServerOptions{
			Host:     cfg.Host,
			CacheAge: cfg.CacheAge,
		}),
		ErrorLog:          log.Default(),
		ReadTimeout:       5 * time.Second,
		ReadHeaderTimeout: 5 * time.Second,
		WriteTimeout:      5 * time.Second,
		IdleTimeout:       5 * time.Second,
	}

	errch := make(chan error)
	donech := make(chan struct{})

	go func() {
		<-sigChan

		logger.Info("Closing server")

		if err := srv.Close(); err != nil {
			logger.Error("Failed to close server", "err", err)
			errch <- err
		}
	}()

	go func() {
		err := srv.ListenAndServe()
		if errors.Is(err, http.ErrServerClosed) {
			donech <- struct{}{}
		} else {
			logger.Error("HTTP server failure", "err", err)
			errch <- err
		}
	}()

	logger.Info("Listening", "addr", srv.Addr)

	select {
	case err := <-errch:
		return err
	case <-donech:
		return nil
	}
}

func parseConfig(args []string) (yamlConfig, error) {
	cfgName := stringValue{
		value: defaultConfigFile,
	}

	if err := parseFlags(args, &cfgName); err != nil {
		return yamlConfig{}, err
	} else if !cfgName.set {
		parseEnv(&cfgName)
	}

	file, err := os.Open(cfgName.value)
	if err != nil {
		return yamlConfig{}, fmt.Errorf("failed to open config file: %w", err)
	}

	var cfg yamlConfig

	err = yaml.NewDecoder(file).Decode(&cfg)
	if err != nil {
		return yamlConfig{}, fmt.Errorf("failed to parse config file: %w", err)
	}

	if cfg.Host == "" {
		cfg.Host = defaultHost
	}

	if cfg.Port == 0 {
		cfg.Port = defaultPort
	}

	if cfg.CacheAge == 0 {
		cfg.CacheAge = defaultCacheAge
	}

	return cfg, nil
}

type yamlConfig struct {
	Host     string        `yaml:"host"`
	Port     uint          `yaml:"port"`
	CacheAge time.Duration `yaml:"cache_age"`
	Packages []yamlPackage `yaml:"packages"`
}

type yamlPackage struct {
	Path          string  `yaml:"path"`
	VCS           yamlVCS `yaml:"vcs"`
	Display       string  `yaml:"display"`
	RepositoryURL string  `yaml:"repository_url"`
}

type yamlVCS struct {
	Value vanityurl.VCS
}

func (vcs yamlVCS) MarshalYAML() (interface{}, error) { //nolint:unparam
	return vcs.String(), nil
}

func (vcs *yamlVCS) UnmarshalYAML(value *yaml.Node) error {
	var strPtr *string

	err := value.Decode(&strPtr)
	if err != nil {
		return err
	} else if strPtr == nil {
		vcs.Value = 0

		return nil
	}

	parsed, err := vanityurl.ParseVCS(*strPtr)
	if err != nil {
		return err
	}

	vcs.Value = parsed

	return nil
}

func (vcs yamlVCS) String() string {
	return vcs.Value.String()
}

type stringValue struct {
	value string
	set   bool
}

func (v *stringValue) Set(str string) error {
	v.value = str
	v.set = true

	return nil
}

func (v *stringValue) String() string {
	return v.value
}

func parseFlags(args []string, cfgName *stringValue) error {
	fset := flag.NewFlagSet("", flag.ContinueOnError)

	fset.Var(cfgName, flagConfigVar, "Config file location")

	return fset.Parse(args)
}

func parseEnv(cfgName *stringValue) {
	val, ok := os.LookupEnv(envConfigVar)
	if !ok {
		return
	}

	_ = cfgName.Set(val)
}
