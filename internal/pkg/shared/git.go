package shared

import (
	"fmt"
)

type git struct{}

var Git git = git{}

func (g git) Run(arguments ...string) (string, string, error) {
	outStr, errStr, err := Execute.RunGetOutput(false, "git", arguments...)
	if err != nil {
		err = fmt.Errorf("git command failed: error: %w, stderr: %s", err, errStr)
	}
	return outStr, errStr, err
}
