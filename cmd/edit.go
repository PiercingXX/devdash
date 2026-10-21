package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/spf13/cobra"
)

var editor string

func editCmd() *cobra.Command {
	editCmd := &cobra.Command{
		Use:   "edit [dashboard]",
		Short: "Edit a dashboard with a shell editor (default: the default dashboard)",
		Run: func(cmd *cobra.Command, args []string) {
			edit(args)
		},
	}

	editCmd.Flags().StringVarP(&editor, "editor", "e", "$EDITOR", "Path of your favorite editor")

	return editCmd
}

func edit(args []string) {
	name := "default"
	if len(args) > 0 {
		name = args[0]
	}

	file := findDashboard(name)
	if file == "" {
		fmt.Fprintf(os.Stdout, "The dashboard %s doesn't exist in %s or in the current directory\n", name, dashPath())
		return
	}

	editDashboard(os.ExpandEnv(editor), file)
}

// editDashboard opens the dashboard's configuration in the editor, and waits for the editor to exit.
func editDashboard(editor string, config string) {
	cmd := exec.Command(editor, config)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		fmt.Println(err)
	}
}

// findDashboard returns the path of a dashboard, given with or without its extension.
func findDashboard(name string) string {
	for _, f := range dashboards() {
		base := filepath.Base(f)
		if name == base || name == removeExt(base) {
			return f
		}
	}

	return ""
}
