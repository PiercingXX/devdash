package cmd

import (
	"fmt"
	"os"
	"runtime"

	"github.com/spf13/cobra"
)

var (
	current   string
	buildDate string
)

func versionCmd() *cobra.Command {
	versionCmd := &cobra.Command{
		Use:   "version",
		Short: "Display the version of DevDash",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Fprintln(os.Stdout, version())
		},
	}

	return versionCmd
}

func version() string {
	program := "devdash"

	osArch := runtime.GOOS + "/" + runtime.GOARCH

	date := buildDate
	if date == "" {
		date = "unknown"
	}

	return fmt.Sprintf("%s %s %s BuildDate=%s", program, current, osArch, date)
}
