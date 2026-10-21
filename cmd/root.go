package cmd

import (
	"errors"
	"fmt"
	"os"
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
	terminal, err := platform.NewTerminal()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	// Create the TUI.
	tui := internal.NewTUI(terminal)
	defer tui.Close()

	// Map dashboard config to a struct Config.
	cfg, cfgFile := mapConfig(cfgName)
	if debug {
		fmt.Fprintf(os.Stdout, "Config file used: %s", cfgFile)
	}

	// Passing a time.Time to this channel reload the entire dashboard.
	hotReload := make(chan time.Time)

	// Add keystrokes to reload and quit
	tui.AddKHotReload(cfg.KHotReload(), hotReload)
	tui.AddKQuit(cfg.KQuit())

	// Passing a bool to this channel stop the automatic reload of the dashboard.
	stopAutoReload := make(chan bool)
	autoReload(cfg.RefreshTime(), stopAutoReload, hotReload)

	editor := os.Getenv("EDITOR")
	if cfg.General.Editor != "" {
		editor = cfg.General.Editor
	}

	// Add keystroke (managed by TUI) to edit the configuration in a CLI editor.
	// Wrap edit config in lambda to defer the execution.
	tui.AddKEdit(
		cfg.KEdit(),
		func() {
			stopReload(stopAutoReload)
			editDashboard(editor, cfgFile)
			hotReload <- time.Now()
			autoReload(cfg.RefreshTime(), stopAutoReload, hotReload)
		},
	)

	// First display.
	build(cfgName, tui)

	// Automatic reload
	go func() {
		for hr := range hotReload {
			tui.HotReload()
			build(cfgName, tui)
			if debug {
				fmt.Println("Last reload: " + hr.Format("2006-01-02 15:04:05"))
			}
		}
	}()

	tui.Loop()
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

// build every services present in the configuration
func build(file string, tui *internal.Tui) {
	cfg, _ := mapConfig(file)
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
				fmt.Println(err)
				internal.DisplayError(tui, err)()
			} else {
				project.WithRemoteHost(remoteHostWidget)
			}
		}

		localhost, err := internal.NewHostWidget("localhost", "localhost")
		if err != nil {
			fmt.Println(err)
			internal.DisplayError(tui, err)()
		}
		project.WithLocalhost(localhost)

		renderFuncs := project.CreateWidgets()
		if !debug {
			project.Render(renderFuncs)
		}
	}
}
