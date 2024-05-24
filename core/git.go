package core

import (
	"bytes"
	"fmt"
	"os/exec"
	"strings"
)

type git struct{}

var Git git = git{}

func (g git) Run(arguments ...string) (string, string, error) {
	cmd := exec.Command("git", arguments...)
	var stdoutBuf, stderrBuf bytes.Buffer
	cmd.Stdout = &stdoutBuf
	cmd.Stderr = &stderrBuf
	err := cmd.Run()
	outStr, errStr := g.processOutputString(stdoutBuf.String()), g.processOutputString(stderrBuf.String())
	if err != nil {
		err = fmt.Errorf("git command failed: error: %w, stderr: %s", err, errStr)
	}
	return outStr, errStr, err
}

func (g git) processOutputString(value string) string {
	return strings.TrimRight(value, "\r\n")
}
