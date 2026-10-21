package internal

import (
	"fmt"
	"slices"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/PiercingXX/devdash/internal/platform"
)

// Problem found in a dashboard.
type Problem struct {
	Path    string
	Message string
}

func (p Problem) String() string {
	return p.Path + ": " + p.Message
}

var (
	generalKeys     = []string{"refresh", "editor", "keys"}
	keyNames        = []string{"quit", "hot_reload", "edit", "next", "previous"}
	projectKeys     = []string{"name", "name_options", "services", "themes", "widgets"}
	nameOptionsKeys = []string{optionSize, optionHeight, optionBold, optionColor, optionTextColor, optionBorderColor}
	elementKeys     = []string{"name", "options", "theme"}

	// serviceKeys are the options of each service.
	serviceKeys = map[string][]string{
		"google_analytics":      {"keyfile", "property_id", "view_id"},
		"google_search_console": {"keyfile", "address"},
		"monitor":               {"address"},
		"github":                {"token", "owner", "repository"},
		"feedly":                {"address"},
		"git":                   {"path"},
		"remote_host":           {"username", "address", "known_hosts"},
	}

	// widgetServices are the services needed by the widgets of each service ID.
	widgetServices = map[string]string{
		"ga":     "google_analytics",
		"gsc":    "google_search_console",
		"mon":    "monitor",
		"github": "github",
		"feedly": "feedly",
		"git":    "git",
		"rh":     "remote_host",
	}

	units = []string{"b", "kb", "mb", "gb", "tb", "pb", "eb"}
)

// CheckDashboard finds the problems of a dashboard, given as its raw configuration.
func CheckDashboard(cfg map[string]any) []Problem {
	c := &checker{}

	c.unknownKeys("", cfg, []string{"general", "projects"})
	if g, ok := c.object("general", cfg["general"]); ok {
		c.unknownKeys("general", g, generalKeys)
		if r, ok := g["refresh"]; ok {
			c.positiveInt("general.refresh", r)
		}
		if k, ok := c.object("general.keys", g["keys"]); ok {
			c.unknownKeys("general.keys", k, keyNames)
		}
	}

	projects, ok := cfg["projects"].([]any)
	if !ok || len(projects) == 0 {
		c.add("projects", "a dashboard needs at least one project")
	}
	for i, p := range projects {
		c.project(fmt.Sprintf("projects[%d]", i), p)
	}

	return c.problems
}

type checker struct {
	problems []Problem
}

func (c *checker) add(path string, format string, args ...any) {
	c.problems = append(c.problems, Problem{Path: path, Message: fmt.Sprintf(format, args...)})
}

func (c *checker) object(path string, v any) (map[string]any, bool) {
	if v == nil {
		return nil, false
	}

	m, ok := v.(map[string]any)
	if !ok {
		c.add(path, "needs to be an object, not %v", v)
	}

	return m, ok
}

// list of a value which can be a single object, like a column.
func list(v any) []any {
	if l, ok := v.([]any); ok {
		return l
	}
	if v == nil {
		return nil
	}

	return []any{v}
}

func (c *checker) unknownKeys(path string, m map[string]any, known []string) {
	for _, k := range sortedKeys(m) {
		if !slices.Contains(known, k) {
			c.add(join(path, k), "unknown option%s", suggestion(k, known))
		}
	}
}

func (c *checker) project(path string, v any) {
	p, ok := c.object(path, v)
	if !ok {
		return
	}

	c.unknownKeys(path, p, projectKeys)
	if no, ok := c.object(path+".name_options", p["name_options"]); ok {
		c.unknownKeys(path+".name_options", no, nameOptionsKeys)
		c.optionValues(path+".name_options", no)
	}

	services, _ := c.object(path+".services", p["services"])
	for _, name := range sortedKeys(services) {
		known, ok := serviceKeys[name]
		if !ok {
			c.add(join(path+".services", name), "unknown service%s", suggestion(name, sortedKeys(serviceKeys)))
			continue
		}
		if s, ok := c.object(join(path+".services", name), services[name]); ok {
			c.unknownKeys(join(path+".services", name), s, known)
		}
	}
	if ga, ok := services["google_analytics"].(map[string]any); ok {
		if _, ok := ga["property_id"]; !ok {
			c.add(path+".services.google_analytics", "property_id is required (view_id was used by Universal Analytics, which Google shut down)")
		}
	}

	themes, _ := c.object(path+".themes", p["themes"])
	all := AllOptions()
	for _, name := range sortedKeys(themes) {
		if t, ok := c.object(join(path+".themes", name), themes[name]); ok {
			c.unknownKeys(join(path+".themes", name), t, all)
			c.optionValues(join(path+".themes", name), t)
		}
	}

	for i, r := range list(p["widgets"]) {
		rp := fmt.Sprintf("%s.widgets[%d]", path, i)
		row, ok := c.object(rp, r)
		if !ok {
			continue
		}
		c.unknownKeys(rp, row, []string{"row"})

		width := 0
		for j, co := range list(row["row"]) {
			cp := fmt.Sprintf("%s.row[%d]", rp, j)
			column, ok := c.object(cp, co)
			if !ok {
				continue
			}
			c.unknownKeys(cp, column, []string{"col"})

			for _, w := range list(column["col"]) {
				col, ok := c.object(cp+".col", w)
				if !ok {
					continue
				}
				c.unknownKeys(cp+".col", col, []string{optionSize, "elements"})
				width += c.size(cp+".col.size", col[optionSize])

				for k, e := range list(col["elements"]) {
					c.element(fmt.Sprintf("%s.col.elements[%d]", cp, k), e, services, themes)
				}
			}
		}
		if width > 12 {
			c.add(rp, "the sizes of the columns add up to %d, more than 12", width)
		}
	}
}

func (c *checker) size(path string, v any) int {
	if v == nil {
		c.add(path, "a column needs a size")
		return 0
	}

	s, err := MapSize(fmt.Sprint(v))
	if err != nil || s < 1 || s > 12 {
		c.add(path, "the size %v needs to be a number from 1 to 12, or XXS, XS, S, M, L, XL, XXL", v)
		return 0
	}

	return s
}

func (c *checker) element(path string, v any, services map[string]any, themes map[string]any) {
	e, ok := c.object(path, v)
	if !ok {
		return
	}
	c.unknownKeys(path, e, elementKeys)

	name, _ := e["name"].(string)
	known, ok := WidgetOptions(name)
	if !ok {
		c.add(path, "unknown widget %q%s", name, suggestion(name, WidgetNames()))
		return
	}
	path += " (" + name + ")"

	w := Widget{Name: name}
	if s, ok := widgetServices[w.serviceID()]; ok {
		if _, ok := services[s]; !ok {
			c.add(path, "the widget needs the service %s in the project", s)
		}
	}

	if t, ok := e["theme"]; ok {
		if _, ok := themes[fmt.Sprint(t)]; !ok {
			c.add(path+".theme", "unknown theme %q", t)
		}
	}

	if o, ok := c.object(path+".options", e["options"]); ok {
		c.unknownKeys(path+".options", o, known)
		c.optionValues(path+".options", o)
	}
}

// optionValues checks the values of the options which have a known format.
func (c *checker) optionValues(path string, options map[string]any) {
	for _, k := range sortedKeys(options) {
		v := fmt.Sprint(options[k])
		p := join(path, k)

		switch {
		case k == optionColor || strings.HasSuffix(k, "_color"):
			if _, ok := colorLookUp[v]; !ok {
				c.add(p, "unknown color %q%s", v, suggestion(v, sortedKeys(colorLookUp)))
			}
		case slices.Contains([]string{optionHeight, optionBarGap, optionBarWidth, optionRowLimit, optionCharLimit}, k):
			c.positiveInt(p, options[k])
		case slices.Contains([]string{optionBold, optionMultiline, optionGlobal, optionRate}, k):
			if _, err := strconv.ParseBool(v); err != nil {
				c.add(p, "needs to be true or false, not %q", v)
			}
		case k == optionStartDate:
			if _, _, err := platform.ConvertDates(time.Now(), v, "today"); err != nil {
				c.add(p, "invalid date %q: use YYYY-MM-DD or an alias like 7_days_ago", v)
			}
		case k == optionEndDate:
			if _, _, err := platform.ConvertDates(time.Now(), "today", v); err != nil {
				c.add(p, "invalid date %q: use YYYY-MM-DD or an alias like 7_days_ago", v)
			}
		case k == optionTimePeriod:
			c.oneOf(p, v, []string{"day", "month", "year"})
		case k == optionScope:
			c.oneOf(p, v, []string{"owner", "all"})
		case k == optionUnit:
			c.oneOf(p, strings.ToLower(v), units)
		case k == optionSize:
			c.size(p, options[k])
		}
	}
}

func (c *checker) positiveInt(path string, v any) {
	if i, err := strconv.Atoi(fmt.Sprint(v)); err != nil || i < 0 {
		c.add(path, "needs to be a positive number, not %q", fmt.Sprint(v))
	}
}

func (c *checker) oneOf(path string, v string, values []string) {
	if !slices.Contains(values, v) {
		c.add(path, "%q needs to be one of %s", v, strings.Join(values, ", "))
	}
}

func join(path string, key string) string {
	if path == "" {
		return key
	}

	return path + "." + key
}

func sortedKeys[V any](m map[string]V) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	return keys
}

// suggestion of the closest known value, if any is close enough.
func suggestion(value string, known []string) string {
	normalize := func(s string) string {
		return strings.NewReplacer("_", "", "-", "", ".", "").Replace(strings.ToLower(s))
	}

	best, distance := "", 3
	for _, k := range known {
		if normalize(k) == normalize(value) {
			return fmt.Sprintf(" (did you mean %s?)", k)
		}
		if d := levenshtein(value, k); d < distance {
			best, distance = k, d
		}
	}

	if best == "" {
		return ""
	}

	return fmt.Sprintf(" (did you mean %s?)", best)
}

func levenshtein(a, b string) int {
	ra, rb := []rune(a), []rune(b)
	prev := make([]int, len(rb)+1)
	for j := range prev {
		prev[j] = j
	}

	for i := 1; i <= len(ra); i++ {
		cur := make([]int, len(rb)+1)
		cur[0] = i
		for j := 1; j <= len(rb); j++ {
			cost := 1
			if ra[i-1] == rb[j-1] {
				cost = 0
			}
			cur[j] = min(prev[j]+1, cur[j-1]+1, prev[j-1]+cost)
		}
		prev = cur
	}

	return prev[len(rb)]
}
