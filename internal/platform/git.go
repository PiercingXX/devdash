package platform

import (
	"bytes"
	"fmt"
	"os/exec"
	"strings"
)

const git = "git"

type Git struct {
	Path string
}

func NewGit(path string) *Git {
	return &Git{
		Path: path,
	}
}

func (g *Git) Branches() ([][]string, error) {
	cmd := exec.Command(
		git,
		"for-each-ref",
		"--sort=committerdate",
		"refs/heads/",
		"--format=%(refname:short),%(authorname),%(authoremail),%(authordate:short)",
	)
	cmd.Dir = g.Path

	cmdOutput := &bytes.Buffer{}
	cmd.Stdout = cmdOutput

	err := cmd.Run()
	if err != nil {
		return nil, fmt.Errorf("can't run %v: %w", git+" "+strings.Join(cmd.Args, " "), err)
	}

	output := cmdOutput.Bytes()
	return formatBranches(string(output)), nil
}

func formatBranches(data string) [][]string {
	data = strings.Replace(data, "<", "", -1)
	data = strings.Replace(data, ">", "", -1)
	result := [][]string{{"Branch", "Last commit by", "Email", "Last commit date"}}
	d := strings.Split(data, "\n")
	for i := 0; i < len(d)-1; i++ {
		result = append(result, strings.Split(d[i], ","))
	}

	return result
}
