package cmd

import (
	"os"

	"github.com/PiercingXX/devdash/internal"
	"github.com/spf13/cobra"
)

func schemaCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "schema",
		Short: "Print the JSON schema of the dashboards, for the completion and validation of editors",
		Run: func(cmd *cobra.Command, args []string) {
			s, err := internal.Schema()
			if err != nil {
				fail(err)
			}
			os.Stdout.Write(s)
		},
	}
}
