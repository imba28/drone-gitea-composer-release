package plugin

import (
	"errors"
	"fmt"
	"regexp"
)

var (
	defaultIgnoredPatterns = []string{
		"vendor",
		"node_modules",
	}
)

type Config struct {
	Owner          string
	Version        string
	GiteaURL       string
	GiteaUser      string
	GiteaToken     string
	IgnorePatterns []string
}

func (c Config) Validate() error {
	if c.Owner == "" {
		return fmt.Errorf("you must provide an owner of the package")
	}
	if c.Version == "" {
		return fmt.Errorf("if no tag is set you must manually specify a version")
	}
	if c.GiteaURL == "" {
		return fmt.Errorf("you must provide the url of your Gitea instance")
	}
	if c.GiteaUser == "" || c.GiteaToken == "" {
		return fmt.Errorf("you must provide valid credentials (username + password/token) for your Gitea instance")
	}
	return nil
}

func ignoredPatterns(p []string) ([]*regexp.Regexp, error) {
	p = append(p, defaultIgnoredPatterns...)
	patterns := make([]*regexp.Regexp, len(p))
	for i := range p {
		reg, err := regexp.Compile(p[i])
		if err != nil {
			return nil, errors.New(fmt.Sprintf("%s is not a valid regular expression", p[i]))
		}
		patterns[i] = reg
	}
	return patterns, nil
}
