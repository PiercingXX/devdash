package cmd

import (
	"bytes"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
	"text/template"

	"github.com/PiercingXX/devdash/internal"
)

func Test_checkDashboard(t *testing.T) {
	path := filepath.Join(t.TempDir(), "dashboard.yml")
	content := `
general:
  refresh: 600
  keyz:
    quit: q
projects:
  - name: test
    services:
      google_analytics:
        keyfile: k.json
        view_id: 123
      githb:
        owner: me
    themes:
      bar:
        bar_colour: red
    widgets:
      - row:
          - col:
              size: 8
              elements:
                - name: ga.bar_users
                  options:
                    startDate: 7_days_ago
                    color: purple
                    height: big
                    end_date: someday
                - name: github.bar_stars
                - name: ga.bar_user
                - name: lh.box_net_io
                  theme: ocean
                  options:
                    rate: yes
          - col:
              size: XL
              elements:
                - name: display.box
`
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	problems, err := checkDashboard(path)
	if err != nil {
		t.Fatal(err)
	}

	actual := []string{}
	for _, p := range problems {
		actual = append(actual, p.String())
	}
	expected := []string{
		"general.keyz: unknown option (did you mean keys?)",
		"projects[0].services.githb: unknown service (did you mean github?)",
		"projects[0].services.google_analytics: property_id is required (view_id was used by Universal Analytics, which Google shut down)",
		"projects[0].themes.bar.bar_colour: unknown option (did you mean bar_color?)",
		"projects[0].widgets[0].row[0].col.elements[0] (ga.bar_users).options.startdate: unknown option (did you mean start_date?)",
		"projects[0].widgets[0].row[0].col.elements[0] (ga.bar_users).options.color: unknown color \"purple\"",
		"projects[0].widgets[0].row[0].col.elements[0] (ga.bar_users).options.end_date: invalid date \"someday\": use YYYY-MM-DD or an alias like 7_days_ago",
		"projects[0].widgets[0].row[0].col.elements[0] (ga.bar_users).options.height: needs to be a positive number, not \"big\"",
		"projects[0].widgets[0].row[0].col.elements[1] (github.bar_stars): the widget needs the service github in the project",
		"projects[0].widgets[0].row[0].col.elements[2]: unknown widget \"ga.bar_user\" (did you mean ga.bar_users?)",
		"projects[0].widgets[0].row[0].col.elements[3] (lh.box_net_io).theme: unknown theme \"ocean\"",
		"projects[0].widgets[0].row[0].col.elements[3] (lh.box_net_io).options.rate: needs to be true or false, not \"yes\"",
		"projects[0].widgets[0]: the sizes of the columns add up to 18, more than 12",
	}

	if strings.Join(actual, "\n") != strings.Join(expected, "\n") {
		t.Errorf("expected:\n%s\nactual:\n%s", strings.Join(expected, "\n"), strings.Join(actual, "\n"))
	}
}

// The examples and the dashboards of the documentation need to be valid.
func Test_checkExamplesAndDocumentation(t *testing.T) {
	examples, _ := filepath.Glob("../example/*.yml")
	docs, _ := filepath.Glob("../doc/*.md")
	services, _ := filepath.Glob("../doc/services/*.md")
	dir := t.TempDir()

	// The YAML blocks with projects in the documentation are complete dashboards.
	block := regexp.MustCompile("(?s)```yaml\n(.*?)```")
	for _, d := range append(docs, services...) {
		src, err := os.ReadFile(d)
		if err != nil {
			t.Fatal(err)
		}
		for i, b := range block.FindAllStringSubmatch(string(src), -1) {
			if !strings.Contains(b[1], "projects:") {
				continue
			}
			f := filepath.Join(dir, strings.TrimSuffix(filepath.Base(d), ".md")+"-"+string(rune('a'+i))+".yml")
			if err := os.WriteFile(f, []byte(b[1]), 0o644); err != nil {
				t.Fatal(err)
			}
			examples = append(examples, f)
		}
	}

	if len(examples) < 10 {
		t.Fatalf("expected the examples and the dashboards of the documentation, actual %v", examples)
	}

	for _, e := range examples {
		problems, err := checkDashboard(e)
		if err != nil {
			t.Errorf("%s: %v", e, err)
		}
		for _, p := range problems {
			t.Errorf("%s: %s", e, p)
		}
	}
}

func Test_checkDefaultConfig(t *testing.T) {
	path := filepath.Join(t.TempDir(), "default.yml")
	if err := os.WriteFile(path, []byte(defaultConfig("/home/me/.config/devdash")), 0o644); err != nil {
		t.Fatal(err)
	}

	problems, err := checkDashboard(path)
	if err != nil || len(problems) > 0 {
		t.Errorf("expected a valid default dashboard, actual %v, %v", problems, err)
	}
}

func Test_checkTemplates(t *testing.T) {
	templates := map[string]string{
		"blog":    renderTemplate(t, internal.Blog(), internal.CreateBlogConfig("k.json", "https://example.com", "123")),
		"project": renderTemplate(t, internal.GitHubProject(), internal.CreateGitHubProjectConfig("token", "me", "repo")),
	}

	for name, content := range templates {
		path := filepath.Join(t.TempDir(), name+".yml")
		if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}

		problems, err := checkDashboard(path)
		if err != nil {
			t.Errorf("%s: %v", name, err)
		}
		for _, p := range problems {
			t.Errorf("%s: %s", name, p)
		}
	}
}

func renderTemplate(t *testing.T, tpl string, data any) string {
	t.Helper()
	b := &bytes.Buffer{}
	if err := template.Must(template.New("t").Parse(tpl)).Execute(b, data); err != nil {
		t.Fatal(err)
	}

	return b.String()
}
