package plugin

import (
	"strings"
	"testing"
)

func TestVersionFromComposerJson(t *testing.T) {
	testCases := []struct {
		content string
		want    string
	}{
		{
			content: "{\"version\":\"2.1.0\"}",
			want:    "2.1.0",
		},
		{
			content: "{\"version\":\"22\"}",
			want:    "22",
		},
		{
			content: "{}",
			want:    "",
		},
		{
			content: "",
			want:    "",
		},
		{
			content: "invalid json",
			want:    "",
		},
	}

	for _, testCase := range testCases {
		v := VersionFromComposerJson(strings.NewReader(testCase.content))
		if v != testCase.want {
			t.Errorf("%s contains version %s, but got %s", testCase.content, testCase.want, v)
		}
	}
}
