package cmd

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/PiercingXX/devdash/internal"
	"github.com/PiercingXX/devdash/internal/platform"
	"github.com/spf13/cobra"
)

var (
	// Used for flags
	cfgName string
	debug   bool

	rootCmd = &cobra.Command{
		Use:   "devdash",
		Short: "DevDash is a highly configurable terminal dashboard for developers and creators",
		Long:  `DevDash is a highly flexible terminal dashboard for developers and creators, which allows you to gather and refresh the data you really need from Google Analytics, Google Search Console, Github, and more.`,
		Run: func(cmd *cobra.Command, args []string) {
			run()
		},
	}
)

func init() {
	rootCmd.Flags().StringVarP(&cfgName, "config", "c", "", "A valid dashboard configuration")
	rootCmd.Flags().BoolVarP(&debug, "debug", "d", false, "Debug mode: fetch the data without displaying the widgets")
	rootCmd.AddCommand(listCmd())
	rootCmd.AddCommand(versionCmd())
	rootCmd.AddCommand(editCmd())
	rootCmd.AddCommand(generateCmd())
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println("Error: " + err.Error())
		os.Exit(1)
	}
}

// run the dashboard.
func run() {
	path, err := findDashboard(cfgName)
	if err != nil {
		fail(err)
	}

	cfg, err := loadConfig(path)
	if err != nil {
		fail(err)
	}

	if debug {
		fmt.Printf("Config file used: %s\n", path)
	}

	terminal, err := platform.NewTerminal()
	if err != nil {
		fail(err)
	}

	tui := internal.NewTUI(terminal)
	defer tui.Close()

	d := &dashboard{path: path}

	// Passing a time.Time to this channel reload the entire dashboard.
	hotReload := make(chan time.Time)

	tui.AddKHotReload(cfg.KHotReload(), hotReload)
	tui.AddKQuit(cfg.KQuit())
	tui.AddKAction(cfg.KNext(), func() {
		d.move(1)
		hotReload <- time.Now()
	})
	tui.AddKAction(cfg.KPrevious(), func() {
		d.move(-1)
		hotReload <- time.Now()
	})

	// Passing a bool to this channel stop the automatic reload of the dashboard.
	stopAutoReload := make(chan bool)
	autoReload(cfg.RefreshTime(), stopAutoReload, hotReload)

	// The editor is given the terminal until it exits.
	tui.AddKEdit(
		cfg.KEdit(),
		func() {
			stopReload(stopAutoReload)
			editDashboard(d.editor(), d.current())
			hotReload <- time.Now()
			autoReload(cfg.RefreshTime(), stopAutoReload, hotReload)
		},
	)

	build(d.current(), tui)

	go func() {
		for hr := range hotReload {
			tui.HotReload()
			build(d.current(), tui)
			if debug {
				fmt.Println("Last reload: " + hr.Format("2006-01-02 15:04:05"))
			}
		}
	}()

	tui.Loop()
}

// fail displays the error and exits.
func fail(err error) {
	fmt.Fprintln(os.Stderr, "Error:", err)
	os.Exit(1)
}

// dashboard currently displayed.
type dashboard struct {
	mu   sync.Mutex
	path string
}

func (d *dashboard) current() string {
	d.mu.Lock()
	defer d.mu.Unlock()

	return d.path
}

// move to the next dashboard of the list (step 1), or the previous one (step -1).
func (d *dashboard) move(step int) {
	d.mu.Lock()
	defer d.mu.Unlock()

	d.path = nextDashboard(dashboards(), d.path, step)
}

// editor of the current dashboard: the general editor option, or $EDITOR.
func (d *dashboard) editor() string {
	if cfg, err := loadConfig(d.current()); err == nil && cfg.General.Editor != "" {
		return cfg.General.Editor
	}

	return os.Getenv("EDITOR")
}

// nextDashboard in the list, after moving from the current one. The current dashboard is kept if the list is empty.
func nextDashboard(list []string, current string, step int) string {
	if len(list) == 0 {
		return current
	}

	abs := func(p string) string {
		a, err := filepath.Abs(p)
		if err != nil {
			return p
		}
		return a
	}

	for i, p := range list {
		if abs(p) == abs(current) {
			return list[((i+step)%len(list)+len(list))%len(list)]
		}
	}

	return list[0]
}

func autoReload(refresh int64, stopAutoReload <-chan bool, hotReload chan<- time.Time) {
	go func() {
		ticker := time.NewTicker(time.Duration(refresh) * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-stopAutoReload:
				return
			case tick := <-ticker.C:
				hotReload <- tick
			}
		}
	}()
}

func stopReload(stopAutoReload chan<- bool) {
	stopAutoReload <- true
}

// showError alone on the dashboard.
func showError(tui *internal.Tui, err error) {
	internal.DisplayError(tui, err)()
	_ = tui.AddCol("12")
	tui.AddRow()
	tui.Render()
}

// build every services present in the configuration
func build(path string, tui *internal.Tui) {
	cfg, err := loadConfig(path)
	if err != nil {
		showError(tui, err)
		return
	}

	for _, p := range cfg.Projects {
		rows, sizes := p.OrderWidgets()
		project := internal.NewProject(p.Name, p.NameOptions, rows, sizes, p.Themes, tui)

		gaService := p.Services.GoogleAnalytics
		if !gaService.empty() {
			gaWidget, err := internal.NewGaWidget(gaService.Keyfile, gaService.PropertyID)
			if gaService.PropertyID == "" {
				err = errors.New("google_analytics: property_id is required (view_id was used by Universal Analytics, which Google shut down; use the ID of your GA4 property instead)")
			}
			if err != nil {
				internal.DisplayError(tui, err)()
			} else {
				project.WithGa(gaWidget)
			}
		}

		gscService := p.Services.GoogleSearchConsole
		if !gscService.empty() {
			gscWidget, err := internal.NewGscWidget(gscService.Keyfile, gscService.Address)
			if err != nil {
				internal.DisplayError(tui, err)()
			} else {
				project.WithGoogleSearchConsole(gscWidget)
			}
		}

		monService := p.Services.Monitor
		if !monService.empty() {
			monWidget, err := internal.NewMonitorWidget(monService.Address)
			if err != nil {
				internal.DisplayError(tui, err)()
			} else {
				project.WithMonitor(monWidget)
			}
		}

		githubService := p.Services.Github
		if !githubService.empty() {
			githubWidget, err := internal.NewGithubWidget(
				githubService.Token,
				githubService.Owner,
				githubService.Repository,
			)
			if err != nil {
				internal.DisplayError(tui, err)()
			} else {
				project.WithGithub(githubWidget)
			}
		}

		feedlyService := p.Services.Feedly
		if !feedlyService.empty() {
			feedlyWidget := internal.NewFeedlyWidget(feedlyService.Address)
			project.WithFeedly(feedlyWidget)
		}

		gitService := p.Services.Git
		if !gitService.empty() {
			gitWidget := internal.NewGitWidget(gitService.Path)
			project.WithGit(gitWidget)
		}

		remoteHostService := p.Services.RemoteHost
		if !remoteHostService.empty() {
			remoteHostWidget, err := internal.NewHostWidget(
				remoteHostService.Username,
				remoteHostService.Address,
			)
			if err != nil {
				internal.DisplayError(tui, err)()
			} else {
				project.WithRemoteHost(remoteHostWidget)
			}
		}

		localhost, err := internal.NewHostWidget("localhost", "localhost")
		if err != nil {
			internal.DisplayError(tui, err)()
		}
		project.WithLocalhost(localhost)

		renderFuncs := project.CreateWidgets()
		if !debug {
			project.Render(renderFuncs)
		}
	}
}
