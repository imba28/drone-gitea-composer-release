package plugin

import (
	"archive/zip"
	"errors"
	"fmt"
	"github.com/sirupsen/logrus"
	"io"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"
)

var (
	ErrPackageExists = errors.New("the package name and/or version are invalid or a package with the same name and version already exist")
)

type Plugin struct {
	config Config
}

func (p Plugin) validFiles() ([]string, error) {
	ignoredPatternRegex, err := ignoredPatterns(p.config.IgnorePatterns)
	if err != nil {
		return nil, err
	}

	var files []string
	err = filepath.WalkDir(".", func(path string, info fs.DirEntry, err error) error {
		// ignore '.' and dot files
		if path[0] == '.' {
			return nil
		}

		for i := range ignoredPatternRegex {
			if ignoredPatternRegex[i].MatchString(path) {
				// if directory should be ignored do not descent into it
				if info.IsDir() {
					return filepath.SkipDir
				}
				return nil
			}
		}

		if !info.IsDir() {
			files = append(files, path)
		}

		return nil
	})

	return files, err
}

func (p Plugin) Execute() error {
	paths, err := p.validFiles()
	if err != nil {
		return err
	}

	logrus.Infof("Found %d files to process.", len(paths))
	file, err := os.CreateTemp(os.TempDir(), "package.*.zip")
	if err != nil {
		return errors.New("could not create temporary file")
	}
	defer os.Remove(file.Name())

	err = createPackage(paths, file)
	if err != nil {
		return err
	}
	// reset file pointer, so we can read from the start again
	_, err = file.Seek(0, 0)
	if err != nil {
		return err
	}

	err = p.uploadPackage(file)
	if err != nil {
		logrus.Infof("Could not create package version %s", p.config.Version)
		return err
	}

	logrus.Infof("Successfully created version %s of package", p.config.Version)
	return nil
}

func (p Plugin) uploadPackage(file io.Reader) error {
	url := fmt.Sprintf("%s/api/packages/%s/composer?version=%s", p.config.GiteaURL, p.config.Owner, p.config.Version)
	req, err := http.NewRequest(http.MethodPut, url, file)
	if err != nil {
		return err
	}

	// https://docs.gitea.io/en-us/api-usage/#oauth2-provider
	if p.config.GiteaToken == "x-oauth-basic" {
		req.Header.Set("Authorization", "Bearer "+p.config.GiteaUser)
	} else {
		req.SetBasicAuth(p.config.GiteaUser, p.config.GiteaToken)
	}

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusCreated {
		if res.StatusCode == http.StatusBadRequest {
			return ErrPackageExists
		}
		b, _ := io.ReadAll(res.Body)
		return errors.New(fmt.Sprintf("an error (%d) occurred while communicating whith gitea: %s", res.StatusCode, b))
	}

	return nil
}

func createPackage(paths []string, w io.Writer) error {
	archive := zip.NewWriter(w)
	for i := range paths {
		err := addFileToArchive(paths[i], archive)
		if err != nil {
			return err
		}
	}

	err := archive.Close()
	if err != nil {
		return err
	}
	return err
}

func addFileToArchive(path string, archive *zip.Writer) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()

	w, err := archive.Create(path)
	if err != nil {
		return err
	}
	if _, err := io.Copy(w, f); err != nil {
		return err
	}
	return nil
}

func New(config Config) Plugin {
	return Plugin{config: config}
}
