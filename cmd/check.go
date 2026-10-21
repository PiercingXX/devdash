package cmd

import (
	"fmt"
	"os"

	"github.com/PiercingXX/devdash/internal"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func checkCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "check [dashboard]",
		Short: "Check a dashboard: unknown options, widgets and services, invalid values (default: the default dashboard)",
		Args:  cobra.MaximumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			name := "default"
			if len(args) > 0 {
				name = args[0]
			}

			path, err := findDashboard(name)
			if err != nil {
				fail(err)
			}

			problems, err := checkDashboard(path)
			if err != nil {
				fail(err)
			}

			for _, p := range problems {
				fmt.Println(p)
			}
			if len(problems) > 0 {
				fmt.Fprintf(os.Stderr, "%s: %d problem(s)\n", path, len(problems))
				os.Exit(1)
			}

			fmt.Printf("%s: no problem found\n", path)
		},
	}
}

// checkDashboard reads a dashboard and returns its problems.
func checkDashboard(path string) ([]internal.Problem, error) {
	if _, err := loadConfig(path); err != nil {
		return nil, err
	}

	v := viper.New()
	v.SetConfigFile(path)
	if err := v.ReadInConfig(); err != nil {
		return nil, err
	}

	return internal.CheckDashboard(v.AllSettings()), nil
}
