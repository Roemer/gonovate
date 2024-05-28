package main

import (
	"archive/zip"
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/google/go-github/v62/github"
	"github.com/roemer/gotaskr"
	"github.com/roemer/gotaskr/execr"
	"github.com/roemer/gotaskr/log"
)

// Internal variables
var outputDirectory = ".build-output"
var version = "0.1.0"

func main() {
	os.Exit(gotaskr.Execute())
}

func init() {
	gotaskr.Task("Compile:All", func() error {
		return nil
	}).
		DependsOn("Compile:Windows").
		DependsOn("Compile:Linux").
		DependsOn("Compile:Mac").
		DependsOn("Compile:MacArm")

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

	gotaskr.Task("Release", func() error {
		log.Informationf("Creating new release for version %s", version)
		gitHubRepoParts := strings.Split(os.Getenv("GITHUB_REPOSITORY"), "/")
		gitHubOwner := gitHubRepoParts[0]
		gitHubRepo := gitHubRepoParts[1]
		gitHubToken := os.Getenv("GITHUB_TOKEN")

		// Create the client
		ctx := context.Background()
		client := github.NewClient(nil).WithAuthToken(gitHubToken)

		// Create the new release
		newRelease := &github.RepositoryRelease{
			Name:    github.String(version),
			Draft:   github.Bool(true),
			TagName: github.String(version),
		}
		release, _, err := client.Repositories.CreateRelease(ctx, gitHubOwner, gitHubRepo, newRelease)
		if err != nil {
			return err
		}
		log.Informationf("Created release: %s", *release.URL)

		// Upload the artifacts
		artifacts, err := os.ReadDir(outputDirectory)
		if err != nil {
			return err
		}
		for _, artifactPath := range artifacts {
			log.Informationf("Uploading artifact %s", artifactPath.Name())
			f, err := os.Open(filepath.Join(outputDirectory, artifactPath.Name()))
			if err != nil {
				return err
			}
			defer f.Close()
			_, _, err = client.Repositories.UploadReleaseAsset(ctx, gitHubOwner, gitHubRepo, *release.ID, &github.UploadOptions{
				Name: artifactPath.Name(),
			}, f)
			if err != nil {
				return err
			}
		}

		return nil
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

	if err := createFlatZip(a, file); err != nil {
		return err
	}
	return os.Remove(file)
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
