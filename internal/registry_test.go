package internal

import (
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"testing"
)

// The registry needs to list every widget the services handle.
func Test_registryHasEveryWidget(t *testing.T) {
	files, err := filepath.Glob("*_widget.go")
	if err != nil {
		t.Fatal(err)
	}

	constant := regexp.MustCompile(`= "((?:rh|ga|gsc|github|mon|feedly|git|display|json)\.[a-z_]+)"`)
	found := []string{}
	for _, f := range files {
		src, err := os.ReadFile(f)
		if err != nil {
			t.Fatal(err)
		}
		for _, m := range constant.FindAllStringSubmatch(string(src), -1) {
			found = append(found, m[1])
		}
	}

	names := WidgetNames()
	for _, n := range found {
		if !slices.Contains(names, n) {
			t.Errorf("the widget %s isn't in the registry", n)
		}
	}
	if len(found) != len(widgetOptions) {
		t.Errorf("expected %d widgets in the registry, actual %d", len(found), len(widgetOptions))
	}
}

func Test_WidgetOptions(t *testing.T) {
	o, ok := WidgetOptions("lh.box_net_io")
	if !ok || !slices.Contains(o, "rate") || !slices.Contains(o, "bold") || slices.Contains(o, "bar_width") {
		t.Errorf("expected the options of a host box, actual %v", o)
	}

	o, _ = WidgetOptions("ga.bar_devices")
	if !slices.Contains(o, "first_color") || slices.Contains(o, "bar_color") {
		t.Errorf("expected the options of a stacked bar, actual %v", o)
	}

	if _, ok := WidgetOptions("ga.foo"); ok {
		t.Error("expected an unknown widget")
	}
}

// The published schema needs to be generated again when the widgets change: go run . schema > doc/devdash.schema.json
func Test_publishedSchema(t *testing.T) {
	published, err := os.ReadFile("../doc/devdash.schema.json")
	if err != nil {
		t.Fatal(err)
	}

	s, err := Schema()
	if err != nil {
		t.Fatal(err)
	}

	if string(published) != string(s) {
		t.Error("doc/devdash.schema.json is outdated: run go run . schema > doc/devdash.schema.json")
	}
}
