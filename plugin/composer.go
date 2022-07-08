package plugin

import (
	"encoding/json"
	"io"
)

func VersionFromComposerJson(f io.Reader) string {
	type composerJson struct {
		Version string `json:"version"`
	}
	b, _ := io.ReadAll(f)
	c := composerJson{}
	_ = json.Unmarshal(b, &c)

	return c.Version
}
