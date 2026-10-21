package cmd

import (
	"os"
	"path/filepath"
	"slices"
	"testing"
)

func Test_dashboards(t *testing.T) {
	dir := t.TempDir()
	files := map[string]string{
		"blog.yml":      "projects:\n  - name: blog\n",
		"my.site.yaml":  "projects:\n  - name: site\n",
		"github.json":   `{"projects": [{"name": "github"}]}`,
		"host.toml":     "[[projects]]\nname = \"host\"\n",
		"other.yml":     "foo: bar\n",
		"projects.md":   "projects:\n",
		"keyfile.json":  `{"type": "service_account"}`,
		"dir.yml/a.yml": "projects:\n",
	}
	for name, content := range files {
		p := filepath.Join(dir, name)
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(p, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	t.Chdir(dir)

	actual := []string{}
	for _, d := range dashboards() {
		if filepath.Dir(d) == "." {
			actual = append(actual, d)
		}
	}
	slices.Sort(actual)

	expected := []string{"blog.yml", "github.json", "host.toml", "my.site.yaml"}
	if !slices.Equal(actual, expected) {
		t.Errorf("Expected %v, actual %v", expected, actual)
	}
}

func Test_removeExt(t *testing.T) {
	for in, expected := range map[string]string{
		"blog.yml":     "blog",
		"my.site.yaml": "my.site",
		"blog.json":    "blog",
		"blog.toml":    "blog",
		"blog":         "blog",
		"blog.md":      "blog.md",
		"a.json.yml":   "a.json",
	} {
		if actual := removeExt(in); actual != expected {
			t.Errorf("removeExt(%q): expected %q, actual %q", in, expected, actual)
		}
	}
}
