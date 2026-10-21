package cmd

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/spf13/cobra"
)

// dashboardExts are the extensions of the dashboard configurations.
var dashboardExts = []string{".yml", ".yaml", ".json", ".toml"}

var extension bool

func listCmd() *cobra.Command {
	listCmd := &cobra.Command{
		Use:   "list",
		Short: "List the dashboards of $XDG_CONFIG_HOME/devdash and of the current directory",
		Run: func(cmd *cobra.Command, args []string) {
			runList()
		},
	}

	listCmd.Flags().BoolVarP(&extension, "extension", "e", false, "Display file extensions")

	return listCmd
}

func runList() {
	for _, f := range dashboards() {
		name := filepath.Base(f)
		if !extension {
			name = removeExt(name)
		}
		fmt.Fprintln(os.Stdout, name)
	}
}

// dashboards returns the paths of the dashboards in the config directory, then in the current directory.
func dashboards() []string {
	paths := []string{}
	for _, dir := range []string{dashPath(), "."} {
		entries, err := os.ReadDir(dir)
		if err != nil {
			continue
		}

		for _, e := range entries {
			p := filepath.Join(dir, e.Name())
			if !e.IsDir() && slices.Contains(dashboardExts, filepath.Ext(e.Name())) && isDashboard(p) {
				paths = append(paths, p)
			}
		}
	}

	return paths
}

// isDashboard if the file has a "projects" key.
func isDashboard(path string) bool {
	file, err := os.Open(path)
	if err != nil {
		return false
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		if strings.Contains(scanner.Text(), "projects") {
			return true
		}
	}

	return false
}
