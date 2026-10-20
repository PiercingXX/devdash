package cmd

// TODO to test

import (
	"bufio"
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

var extension bool

func listCmd() *cobra.Command {
	listCmd := &cobra.Command{
		Use:   "list",
		Short: "List your devdash boards",
		// TODO write some help
		Run: func(cmd *cobra.Command, args []string) {
			runList()
		},
	}

	listCmd.Flags().BoolVarP(&extension, "extension", "e", false, "Display file extensions")

	return listCmd
}

func runList() {
	if _, err := os.Stat(dashPath()); err != nil {
		log.Fatal(err)
	}

	fs := getConfigFiles()
	for _, f := range fs {
		s := strings.Split(f.Name(), ".")
		// TODO erk to refactor
		if !f.IsDir() && len(s) > 1 && (s[1] == "json" || s[1] == "toml" || s[1] == "yaml" || s[1] == "yml") {
			if extension {
				fmt.Fprintln(os.Stdout, f.Name())
			} else {
				fmt.Fprintln(os.Stdout, s[0])
			}
		}
	}
}

func getConfigFiles() []fs.DirEntry {
	homeFiles, err := os.ReadDir(dashPath())
	if err != nil {
		log.Fatal(err)
	}
	currentFiles, err := os.ReadDir(".")
	if err != nil {
		log.Fatal(err)
	}

	fs := []fs.DirEntry{}
	for _, v := range homeFiles {
		f, isDash := isDashboard(v, dashPath())
		if isDash {
			fs = append(fs, f)
		}
	}

	for _, v := range currentFiles {
		f, isDash := isDashboard(v, ".")
		if isDash {
			fs = append(fs, f)
		}
	}

	return fs
}

func isDashboard(entry fs.DirEntry, path string) (fs.DirEntry, bool) {
	file, err := os.Open(path + string(filepath.Separator) + entry.Name())
	if err != nil {
		panic(err)
	}

	defer file.Close()

	scanner := bufio.NewScanner(file)

	scanner.Split(bufio.ScanLines)

	for scanner.Scan() {
		if strings.Contains(scanner.Text(), "projects:") {
			return entry, true
		}
	}

	return nil, false
}
