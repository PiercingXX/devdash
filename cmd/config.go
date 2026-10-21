package cmd

import (
	"bytes"
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

// Map config and return it with the config path
func mapConfig(cfgFile string) (config, string) {
	if cfgFile == "" {
		cfgFile = "default.yml"
		createConfig(dashPath(), cfgFile, defaultConfig(dashPath()))
	}

	// viper.AddConfigPath(home)
	viper.AddConfigPath(dashPath())
	viper.AddConfigPath(".")

	viper.SetConfigName(removeExt(cfgFile))
	err := viper.ReadInConfig()
	if err != nil {
		tryReadFile(cfgFile)
	}

	var cfg config
	if err := viper.Unmarshal(&cfg); err != nil {
		panic(err)
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

	return cfg, viper.ConfigFileUsed()
}

// removeExt of a dashboard's file name.
func removeExt(file string) string {
	if slices.Contains(dashboardExts, filepath.Ext(file)) {
		return strings.TrimSuffix(file, filepath.Ext(file))
	}

	return file
}

func createConfig(path string, filename string, template string) string {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		os.MkdirAll(path, 0755)
	}

	f := filepath.Join(path, filename)
	if _, err := os.Stat(f); os.IsNotExist(err) {
		file, _ := os.Create(f)
		defer file.Close()

		if file != nil {
			_, err := file.Write([]byte(template))
			if err != nil {
				panic(err)
			}
		}
	}

	return f
}

func tryReadFile(cfgFile string) {
	if _, err := os.Stat(cfgFile); os.IsNotExist(err) {
		panic(fmt.Errorf("config %s doesn't exist", cfgFile))
	}

	f, err := os.ReadFile(cfgFile)
	if err != nil {
		panic(fmt.Errorf("could not read file %s", cfgFile))
	}

	viper.SetConfigType(strings.Trim(filepath.Ext(cfgFile), "."))
	err = viper.ReadConfig(bytes.NewBuffer(f))
	if err != nil {
		panic(fmt.Errorf("could not read config %s data", string(f)))
	}
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
