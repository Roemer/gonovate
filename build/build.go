package main

import (
	"os"
	"path/filepath"

	"github.com/roemer/gotaskr"
	"github.com/roemer/gotaskr/execr"
)

// Internal variables
var outputDirectory = ".build-output"

func main() {
	os.Exit(gotaskr.Execute())
}

func init() {
	gotaskr.Task("Compile:Windows", func() error {
		os.Setenv("GOOS", "windows")
		os.Setenv("GOARCH", "amd64")
		return compile("exe")
	})

	gotaskr.Task("Compile:Linux", func() error {
		os.Setenv("GOOS", "linux")
		os.Setenv("GOARCH", "amd64")
		return compile("bin")
	})
}

func compile(ext string) error {
	return execr.Run(true, "go", "build", "-o", filepath.Join(outputDirectory, "gonovate"+"."+ext))
}
