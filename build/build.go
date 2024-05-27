package main

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/roemer/gotaskr"
	"github.com/roemer/gotaskr/execr"
)

// Internal variables
var outputDirectory = ".build-output"
var version = "0.1.0"

func main() {
	os.Exit(gotaskr.Execute())
}

func init() {
	gotaskr.Task("Compile:Windows", func() error {
		os.Setenv("GOOS", "windows")
		os.Setenv("GOARCH", "amd64")

		path, err := compile(".exe")
		if err != nil {
			return err
		}
		return zipRelease(path)
	})

	gotaskr.Task("Compile:Linux", func() error {
		os.Setenv("GOOS", "linux")
		os.Setenv("GOARCH", "amd64")

		path, err := compile("")
		if err != nil {
			return err
		}
		return zipRelease(path)
	})

	gotaskr.Task("Compile:Mac", func() error {
		os.Setenv("GOOS", "darwin")
		os.Setenv("GOARCH", "amd64")

		path, err := compile(".dmg")
		if err != nil {
			return err
		}
		return zipRelease(path)
	})

	gotaskr.Task("Compile:MacArm", func() error {
		os.Setenv("GOOS", "darwin")
		os.Setenv("GOARCH", "arm64")

		path, err := compile(".dmg")
		if err != nil {
			return err
		}
		return zipRelease(path)
	})
}

func compile(ext string) (string, error) {
	outputFile := filepath.Join(outputDirectory, "gonovate"+ext)
	return outputFile, execr.Run(true, "go", "build", "-o", outputFile)
}

func zipRelease(file string) error {
	zipFilePath := filepath.Join(outputDirectory, fmt.Sprintf("gonovate-%s-%s-%s.zip", os.Getenv("GOOS"), version, os.Getenv("GOARCH")))

	a, err := os.Create(zipFilePath)
	if err != nil {
		return err
	}
	defer a.Close()

	return createFlatZip(a, file)
}

func createFlatZip(w io.Writer, files ...string) error {
	z := zip.NewWriter(w)
	for _, file := range files {
		src, err := os.Open(file)
		if err != nil {
			return err
		}
		info, err := src.Stat()
		if err != nil {
			return err
		}
		hdr, err := zip.FileInfoHeader(info)
		if err != nil {
			return err
		}
		hdr.Name = filepath.Base(file) // Write only the base name in the header
		dst, err := z.CreateHeader(hdr)
		if err != nil {
			return err
		}
		_, err = io.Copy(dst, src)
		if err != nil {
			return err
		}
		src.Close()
	}
	return z.Close()
}
