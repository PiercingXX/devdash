package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"slices"
	"strings"

	"github.com/PiercingXX/devdash/internal"
	"github.com/adrg/xdg"
	"github.com/spf13/viper"
)

const (
	// keys
	kQuit      = "C-c"
	kHotReload = "C-r"
	kEdit      = "C-e"
	kNext      = "C-n"
	kPrevious  = "C-p"
)

type config struct {
	General  General   `mapstructure:"general"`
	Projects []Project `mapstructure:"projects"`
}

type General struct {
	Keys    map[string]string `mapstructure:"keys"`
	Refresh int64             `mapstructure:"refresh"`
	Editor  string            `mapstructure:"editor"`
}

// RefreshTime return the duration before refreshing the data of all widgets, in seconds.
func (c config) RefreshTime() int64 {
	if c.General.Refresh == 0 {
		return 60
	}

	return c.General.Refresh
}

type Project struct {
	Name        string                       `mapstructure:"name"`
	NameOptions map[string]string            `mapstructure:"name_options"`
	Services    Services                     `mapstructure:"services"`
	Themes      map[string]map[string]string `mapstructure:"themes"`
	Widgets     []Row                        `mapstructure:"widgets"`
}

// Row is constituted of columns
type Row struct {
	Row []Column `mapstructure:"row"`
}

// Column is constituted of widgets
type Column struct {
	Col []Widgets `mapstructure:"col"`
}

type Widgets struct {
	Size     string            `mapstructure:"size"`
	Elements []internal.Widget `mapstructure:"elements"`
}

type Services struct {
	GoogleAnalytics     GoogleAnalytics `mapstructure:"google_analytics"`
	GoogleSearchConsole SearchConsole   `mapstructure:"google_search_console"`
	Monitor             Monitor         `mapstructure:"monitor"`
	Github              Github          `mapstructure:"github"`
	Feedly              Feedly          `mapstructure:"feedly"`
	Git                 Git             `mapstructure:"git"`
	RemoteHost          RemoteHost      `mapstructure:"remote_host"`
}

type GoogleAnalytics struct {
	Keyfile    string `mapstructure:"keyfile"`
	PropertyID string `mapstructure:"property_id"`
	// ViewID was used by Universal Analytics, shut down by Google in favor of GA4.
	ViewID string `mapstructure:"view_id"`
}

type SearchConsole struct {
	Keyfile string `mapstructure:"keyfile"`
	Address string `mapstructure:"address"`
}

type Monitor struct {
	Address string `mapstructure:"address"`
}

type Github struct {
	Token      string `mapstructure:"token"`
	Owner      string `mapstructure:"owner"`
	Repository string `mapstructure:"repository"`
}

type Feedly struct {
	Address string `mapstructure:"address"`
}

type RemoteHost struct {
	Username string `mapstructure:"username"`
	Address  string `mapstructure:"address"`
}

type Git struct {
	Path string `mapstructure:"path"`
}

func (g GoogleAnalytics) empty() bool {
	return g == GoogleAnalytics{}
}

func (m Monitor) empty() bool {
	return m == Monitor{}
}
func (s SearchConsole) empty() bool {
	return s == SearchConsole{}
}

func (g Github) empty() bool {
	return g == Github{}
}

func (f Feedly) empty() bool {
	return f == Feedly{}
}

func (g Git) empty() bool {
	return g == Git{}
}

func (g RemoteHost) empty() bool {
	return g == RemoteHost{}
}

// OrderWidgets add the widgets to a three dimensional slice.
// First dimension: index of the rows (ir or indexRows).
// Second dimension: index of the columns (ic or indexColumn).
// Third dimension: index of the widget.
func (p Project) OrderWidgets() ([][][]internal.Widget, [][]string) {
	rows := make([][][]internal.Widget, len(p.Widgets))
	sizes := make([][]string, len(p.Widgets))
	for ir, r := range p.Widgets {
		for ic, c := range r.Row {
			rows[ir] = append(rows[ir], []internal.Widget{}) // add columns to rows
			for _, ws := range c.Col {
				// keep sizes of columns and good order of widgets in a separate slice
				sizes[ir] = append(sizes[ir], ws.Size)

				// add widgets to columns
				rows[ir][ic] = append(rows[ir][ic], ws.Elements...)
			}
		}
	}

	return rows, sizes
}

func dashPath() string {
	return filepath.Join(xdg.ConfigHome, "devdash")
}

// findDashboard returns the path of a dashboard: a path to a file, or a name with or without extension,
// searched in the config directory then in the current directory.
// Without a name, the default dashboard is created if it doesn't exist.
func findDashboard(name string) (string, error) {
	if name == "" {
		return createConfig(dashPath(), "default.yml", defaultConfig(dashPath()))
	}

	if fileExists(name) {
		return name, nil
	}

	for _, dir := range []string{dashPath(), "."} {
		candidates := []string{filepath.Join(dir, name)}
		for _, ext := range dashboardExts {
			candidates = append(candidates, filepath.Join(dir, name+ext))
		}

		for _, c := range candidates {
			if fileExists(c) {
				return c, nil
			}
		}
	}

	return "", fmt.Errorf("can't find the dashboard %s in %s or in the current directory", name, dashPath())
}

// loadConfig of a dashboard file.
func loadConfig(path string) (config, error) {
	v := viper.New()
	v.SetConfigFile(path)
	if ext := strings.TrimPrefix(filepath.Ext(path), "."); ext == "yml" {
		v.SetConfigType("yaml")
	}

	var cfg config
	if err := v.ReadInConfig(); err != nil {
		return cfg, fmt.Errorf("can't read the dashboard %s: %w", path, err)
	}

	if err := v.Unmarshal(&cfg); err != nil {
		return cfg, fmt.Errorf("invalid dashboard %s: %w", path, err)
	}

	prefix := "DEVDASH"
	for k := range cfg.Projects {
		if cfg.Projects[k].Services.GoogleAnalytics.Keyfile == "" {
			cfg.Projects[k].Services.GoogleAnalytics.Keyfile = os.Getenv(prefix + "_GA_KEYFILE")
		}

		if cfg.Projects[k].Services.GoogleSearchConsole.Keyfile == "" {
			cfg.Projects[k].Services.GoogleSearchConsole.Keyfile = os.Getenv(prefix + "_GSC_KEYFILE")
		}

		if cfg.Projects[k].Services.Github.Token == "" {
			cfg.Projects[k].Services.Github.Token = os.Getenv(prefix + "_GITHUB_TOKEN")
		}
	}

	return cfg, nil
}

// removeExt of a dashboard's file name.
func removeExt(file string) string {
	if slices.Contains(dashboardExts, filepath.Ext(file)) {
		return strings.TrimSuffix(file, filepath.Ext(file))
	}

	return file
}

func fileExists(path string) bool {
	f, err := os.Stat(path)
	return err == nil && !f.IsDir()
}

// createConfig with a template if it doesn't exist, and return its path.
func createConfig(path string, filename string, template string) (string, error) {
	if err := os.MkdirAll(path, 0o755); err != nil {
		return "", err
	}

	f := filepath.Join(path, filename)
	if _, err := os.Stat(f); err == nil {
		return f, nil
	}

	return f, os.WriteFile(f, []byte(template), 0o644)
}

// Keyboard events
func (c config) KQuit() string {
	if ok := c.General.Keys["quit"]; ok != "" {
		return c.General.Keys["quit"]
	}

	return kQuit
}

func (c config) KHotReload() string {
	if ok := c.General.Keys["hot_reload"]; ok != "" {
		return c.General.Keys["hot_reload"]
	}

	return kHotReload
}

func (c config) KEdit() string {
	if ok := c.General.Keys["edit"]; ok != "" {
		return c.General.Keys["edit"]
	}

	return kEdit
}

func (c config) KNext() string {
	if k := c.General.Keys["next"]; k != "" {
		return k
	}

	return kNext
}

func (c config) KPrevious() string {
	if k := c.General.Keys["previous"]; k != "" {
		return k
	}

	return kPrevious
}

// defaultConfig is the dashboard created on the first run. The host widgets need Linux.
func defaultConfig(dashPath string) string {
	host := ""
	if runtime.GOOS == "linux" {
		host = `
          - col:
              size: 4
              elements:
                - name: lh.gauge_cpu_rate
                  options:
                    title: " CPU (since boot) "
                    bar_color: green
                - name: lh.gauge_memory_rate
                  options:
                    bar_color: yellow
                - name: lh.gauge_swap_rate
                  options:
                    bar_color: red
          - col:
              size: 4
              elements:
                - name: lh.box_uptime
                - name: lh.box_load
                - name: lh.box_processes
      - row:
          - col:
              size: 12
              elements:
                - name: lh.table_disk`
	}

	return fmt.Sprintf(`---
general:
  refresh: 600
  keys:
    quit: "C-c"
    hot_reload: "C-r"
    edit: "C-e"

projects:
  - name: Default dashboard - %s
    services:
      monitor:
        address: "https://github.com"
    widgets:
      - row:
          - col:
              size: 4
              elements:
                - name: display.box
                  options:
                    title: " Welcome to DevDash "
                    content: "CTRL+e edits this dashboard, CTRL+r reloads it, CTRL+c quits.\n\nDocumentation and examples:\ngithub.com/PiercingXX/devdash"
                    multiline: true
                    color: cyan
                - name: mon.box_availability
                  options:
                    title: " github.com "
                    color: green%s
`, filepath.Join(dashPath, "default.yml"), host)
}
