package internal

import (
	"bytes"
	"testing"
	"text/template"

	"go.yaml.in/yaml/v3"
)

func Test_templatesAreValidYAML(t *testing.T) {
	testCases := []struct {
		name     string
		template string
		data     any
	}{
		{
			name:     "blog",
			template: Blog(),
			data:     CreateBlogConfig("key.json", "https://example.com", "123456"),
		},
		{
			name:     "blog without values",
			template: Blog(),
			data:     CreateBlogConfig("", "", ""),
		},
		{
			name:     "github project",
			template: GitHubProject(),
			data:     CreateGitHubProjectConfig("token", "owner", "repo"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tpl, err := template.New(tc.name).Parse(tc.template)
			if err != nil {
				t.Fatal(err)
			}

			b := bytes.Buffer{}
			if err := tpl.Execute(&b, tc.data); err != nil {
				t.Fatal(err)
			}

			var cfg map[string]any
			if err := yaml.Unmarshal(b.Bytes(), &cfg); err != nil {
				t.Fatalf("invalid YAML: %v\n%s", err, b.String())
			}

			if _, ok := cfg["projects"]; !ok {
				t.Errorf("expected a projects key in\n%s", b.String())
			}
		})
	}
}

func Test_CreateBlogConfig(t *testing.T) {
	c := CreateBlogConfig("key.json", "https://example.com", "123456")
	if c.Address != "address: https://example.com" || c.PropertyID != "property_id: 123456" {
		t.Errorf("unexpected config %+v", c)
	}
}
