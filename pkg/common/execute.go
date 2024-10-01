package shared

import (
	"bytes"
	"io"
	"os"
	"os/exec"
	"strings"
)

type execute struct{}

var Execute execute = execute{}

// Run runs an executable with the given arguments and returns the output (stdout and stderr separate).
func (e execute) RunGetOutput(outputToConsole bool, executable string, arguments ...string) (string, string, error) {
	cmd := exec.Command(executable, arguments...)
	return e.RunCommandGetOutput(outputToConsole, cmd)
}

// Run runs an executable with the given arguments and returns the output (stdout and stderr combined).
func (e execute) RunGetCombinedOutput(outputToConsole bool, executable string, arguments ...string) (string, error) {
	cmd := exec.Command(executable, arguments...)
	return e.RunCommandGetCombinedOutput(outputToConsole, cmd)
}

func (e execute) RunCommandGetOutput(outputToConsole bool, cmd *exec.Cmd) (string, string, error) {
	var stdoutBuf, stderrBuf bytes.Buffer
	if outputToConsole {
		cmd.Stdout = io.MultiWriter(os.Stdout, &stdoutBuf)
		cmd.Stderr = io.MultiWriter(os.Stderr, &stderrBuf)
	} else {
		cmd.Stdout = &stdoutBuf
		cmd.Stderr = &stderrBuf
	}
	err := cmd.Run()
	outStr, errStr := e.processOutputString(stdoutBuf.String()), e.processOutputString(stderrBuf.String())
	return outStr, errStr, err
}

func (e execute) RunCommandGetCombinedOutput(outputToConsole bool, cmd *exec.Cmd) (string, error) {
	var outBuf bytes.Buffer
	if outputToConsole {
		cmd.Stdout = io.MultiWriter(os.Stdout, &outBuf)
		cmd.Stderr = io.MultiWriter(os.Stderr, &outBuf)
	} else {
		cmd.Stdout = &outBuf
		cmd.Stderr = &outBuf
	}
	err := cmd.Run()
	outStr := e.processOutputString(outBuf.String())
	return outStr, err
}

func (g execute) processOutputString(value string) string {
	return strings.TrimRight(value, "\r\n")
}
