package cmd

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func Test_nextDashboard(t *testing.T) {
	list := []string{"a.yml", "b.json", "c.toml"}
	testCases := []struct {
		name     string
		current  string
		step     int
		expected string
	}{
		{"next", "a.yml", 1, "b.json"},
		{"next of the last one", "c.toml", 1, "a.yml"},
		{"previous of the first one", "a.yml", -1, "c.toml"},
		{"current not in the list", "other.yml", 1, "a.yml"},
		{"same file with another path", "./b.json", 1, "c.toml"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if actual := nextDashboard(list, tc.current, tc.step); actual != tc.expected {
				t.Errorf("expected %q, actual %q", tc.expected, actual)
			}
		})
	}

	if actual := nextDashboard(nil, "a.yml", 1); actual != "a.yml" {
		t.Errorf("expected to keep the current dashboard without any other, actual %q", actual)
	}
}

func Test_findDashboardAndLoadConfig(t *testing.T) {
	dir := t.TempDir()
	files := map[string]string{
		"blog.yml":    "general:\n  refresh: 30\nprojects:\n  - name: blog\n",
		"github.json": `{"projects": [{"name": "github"}]}`,
		"host.toml":   "[[projects]]\nname = \"host\"\n",
		"broken.yml":  "projects: [\n",
	}
	for name, content := range files {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	t.Chdir(dir)

	for name, project := range map[string]string{"blog": "blog", "github": "github", "host.toml": "host", filepath.Join(dir, "blog.yml"): "blog"} {
		path, err := findDashboard(name)
		if err != nil {
			t.Fatalf("findDashboard(%q): %v", name, err)
		}

		cfg, err := loadConfig(path)
		if err != nil {
			t.Fatalf("loadConfig(%q): %v", path, err)
		}
		if len(cfg.Projects) != 1 || cfg.Projects[0].Name != project {
			t.Errorf("loadConfig(%q): expected the project %s, actual %+v", path, project, cfg.Projects)
		}
	}

	if _, err := findDashboard("missing"); err == nil || !strings.Contains(err.Error(), "can't find the dashboard missing") {
		t.Errorf("expected an error for a missing dashboard, actual %v", err)
	}

	if _, err := loadConfig("broken.yml"); err == nil {
		t.Error("expected an error for an invalid dashboard")
	}
}
