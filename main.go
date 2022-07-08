package main

import (
	"fmt"
	"github.com/cnmade/gonetrc"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli"
	"imba28/drone-gitea-composer-release/plugin"
	"net/url"
	"os"
)

func main() {
	app := cli.NewApp()
	app.Name = "gitea-composer release"
	app.Action = runPlugin
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:   "owner",
			Usage:  "Owner of composer package",
			EnvVar: "PLUGIN_OWNER,DRONE_REPO_OWNER,DRONE_REPO_NAMESPACE",
		},
		cli.StringFlag{
			Name:   "version",
			Usage:  "Version of Composer package to create, e.g 1.0.0",
			EnvVar: "PLUGIN_VERSION,DRONE_TAG",
		},
		cli.StringFlag{
			Name:   "gitea-url",
			Usage:  "URL of the gitea instance",
			EnvVar: "PLUGIN_GITEA_URL,DRONE_REPO_LINK,DRONE_REMOTE_URL,CI_REPO_LINK",
		},
		cli.StringSliceFlag{
			Name:   "ignore-patterns",
			Usage:  "Files to exclude when creating the package zip archive",
			EnvVar: "PLUGIN_IGNORE_PATTERNS",
		},
		cli.StringFlag{
			Name:   "gitea-user",
			Usage:  "Username of Gitea user used for authentication. Required when the repository is not private.",
			EnvVar: "PLUGIN_GITEA_USER",
		},
		cli.StringFlag{
			Name:   "gitea-token",
			Usage:  "a valid password or access token. Required when the repository is not private.",
			EnvVar: "PLUGIN_GITEA_TOKEN",
		},
	}

	if err := app.Run(os.Args); err != nil {
		logrus.Fatal(err)
	}
}

func runPlugin(c *cli.Context) error {
	config := plugin.Config{
		Owner:          c.String("owner"),
		IgnorePatterns: c.StringSlice("ignore-patterns"),
	}

	// parse URL of Gitea instance
	u, err := url.Parse(c.String("gitea-url"))
	if err != nil {
		return err
	}
	config.GiteaURL = fmt.Sprintf("%s://%s", u.Scheme, u.Host)

	username, token := c.String("gitea-user"), c.String("gitea-token")
	if username == "" || token == "" {
		// if credentials are not explicitly set try to read them from the .netrc file drone creates
		username, token = gonetrc.GetCredentials(u.Hostname())
	}
	config.GiteaUser, config.GiteaToken = username, token

	// parse version
	v := c.String("version")
	if v == "" {
		if f, err := os.Open("composer.json"); err == nil {
			defer f.Close()
			v = plugin.VersionFromComposerJson(f)
		}
	}
	config.Version = v

	if err := config.Validate(); err != nil {
		return err
	}

	return plugin.New(config).Execute()
}

// just to ensure the run function is compatible
var _ cli.ActionFunc = runPlugin
