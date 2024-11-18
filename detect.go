package vanityurl

import (
	"fmt"
	"net/url"
)

const (
	hostGithub    = "github.com"
	hostGitlab    = "gitlab.com"
	hostBitbucket = "bitbucket.org"
)

//nolint:gochecknoglobals
var (
	hostVCSMap = map[string]VCS{
		hostGithub:    Git,
		hostGitlab:    Git,
		hostBitbucket: Git,
	}

	hostDisplayMap = map[string]func(string) string{
		hostGithub:    displayGithubOrGitlab,
		hostGitlab:    displayGithubOrGitlab,
		hostBitbucket: displayBitbucket,
	}
)

// DetectVCS type for a given repository URL.
// Supports detection for Github, Gitlab and Bitbucket.
func DetectVCS(repoURL string) VCS {
	url, err := url.Parse(repoURL)
	if err != nil {
		return 0
	}

	return hostVCSMap[url.Hostname()]
}

// DetectDisplay field for a given repository URL.
// Supports detection for Github, Gitlab and Bitbucket.
func DetectDisplay(repoURL string) string {
	url, err := url.Parse(repoURL)
	if err != nil {
		return ""
	}

	fn, ok := hostDisplayMap[url.Hostname()]
	if !ok {
		return ""
	}

	return fn(repoURL)
}

func displayGithubOrGitlab(repositoryURL string) string {
	return fmt.Sprintf("%v %v/tree/master{/dir} %v/blob/master{/dir}/{file}#L{line}",
		repositoryURL,
		repositoryURL,
		repositoryURL,
	)
}

func displayBitbucket(repositoryURL string) string {
	return fmt.Sprintf("%v %v/src/default{/dir} %v/src/default{/dir}/{file}#{file}-{line}",
		repositoryURL,
		repositoryURL,
		repositoryURL,
	)
}
