package cmd

import (
	"fmt"
	"os"
	"os/exec"

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

	file, err := findDashboard(name)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
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
