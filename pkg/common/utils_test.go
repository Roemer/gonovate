package common

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTitle_Default(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)
	settings := &TitleBuilderSettings{
		DependencyName: "roemer/foo",
		NewRelease:     &ReleaseInfo{VersionString: "1.2.3"},
	}
	title, err := BuildTitle(settings)
	require.NoError(err)
	assert.Equal("Update 'roemer/foo' to '1.2.3'", title)
}

func TestTitle_WithGroupName(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)
	settings := &TitleBuilderSettings{
		GroupName: "core-deps",
	}
	title, err := BuildTitle(settings)
	require.NoError(err)
	assert.Equal("Update group 'core-deps'", title)
}

func TestBranchName_Default(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)
	settings := &BranchNameBuilderSettings{
		BaseBranch:     "main",
		DependencyName: "roemer/foo",
		NewRelease:     &ReleaseInfo{VersionString: "1.2.3"},
	}
	branchName, err := BuildBranchName(settings)
	require.NoError(err)
	assert.Equal("main-roemer-foo-1.2.3", branchName)
}

func TestBranchName_WithGroupName(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)
	settings := &BranchNameBuilderSettings{
		GroupName: "core-deps",
	}
	branchName, err := BuildBranchName(settings)
	require.NoError(err)
	assert.Equal("core-deps", branchName)
}

func TestTitle_CustomTemplate_Valid(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)
	settings := &TitleBuilderSettings{
		TitleTemplate:  "Upgrade {{.DependencyName}} -> {{.NewVersion}}",
		DependencyName: "acme/pkg",
		NewRelease:     &ReleaseInfo{VersionString: "2.0.0"},
	}
	title, err := BuildTitle(settings)
	require.NoError(err)
	assert.Equal("Upgrade acme/pkg -> 2.0.0", title)
}

func TestBranchName_CustomTemplate_Valid(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)
	settings := &BranchNameBuilderSettings{
		BranchNameTemplate: "{{.BaseBranch}}/{{.DependencyName}}/{{.NewVersion}}",
		BaseBranch:         "Main",
		DependencyName:     "roemer/foo",
		NewRelease:         &ReleaseInfo{VersionString: "1.2.3"},
	}
	branchName, err := BuildBranchName(settings)
	require.NoError(err)
	assert.Equal("main-roemer-foo-1.2.3", branchName)
}

func TestTitle_InvalidTemplateSyntax(t *testing.T) {
	require := require.New(t)
	settings := &TitleBuilderSettings{
		TitleTemplate:  "Update {{.DependencyName",
		DependencyName: "x",
		NewRelease:     &ReleaseInfo{VersionString: "1.0.0"},
	}
	_, err := BuildTitle(settings)
	require.Error(err)
	require.Contains(err.Error(), "title template")
}

func TestBranchName_InvalidTemplateSyntax(t *testing.T) {
	require := require.New(t)
	settings := &BranchNameBuilderSettings{
		BranchNameTemplate: "{{.BaseBranch",
		BaseBranch:         "main",
	}
	_, err := BuildBranchName(settings)
	require.Error(err)
	require.Contains(err.Error(), "branch-name template")
}

func TestTitle_TemplateExecutionError_MissingField(t *testing.T) {
	require := require.New(t)
	settings := &TitleBuilderSettings{
		TitleTemplate:  "Broken {{.UnknownField}}",
		DependencyName: "x",
		NewRelease:     &ReleaseInfo{VersionString: "1.0.0"},
	}
	_, err := BuildTitle(settings)
	require.Error(err)
	require.Contains(err.Error(), "title template")
}

func TestBranchName_TemplateExecutionError_MissingField(t *testing.T) {
	require := require.New(t)
	settings := &BranchNameBuilderSettings{
		BranchNameTemplate: "{{.NonExistent}}",
		BaseBranch:         "main",
	}
	_, err := BuildBranchName(settings)
	require.Error(err)
	require.Contains(err.Error(), "branch-name template")
}

func TestBranchName_UsesNormalizedFields(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)
	settings := &BranchNameBuilderSettings{
		BranchNameTemplate: "{{.BaseBranch}}-{{.DependencyName}}-{{.NewVersion}}",
		BaseBranch:         "Main",
		DependencyName:     "Gonovate/Repo",
		NewRelease:         &ReleaseInfo{VersionString: "V1.2.3"},
	}
	branchName, err := BuildBranchName(settings)
	require.NoError(err)
	assert.Equal("main-gonovate-repo-v1.2.3", branchName)
}

func TestBranchName_UsesRawFields(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)
	longBase := "ThisIsAVeryLongBranchNameThatExceedsTwentyChars"
	// Using normalized field (pre-truncated to 20 chars)
	normSettings := &BranchNameBuilderSettings{
		BranchNameTemplate: "{{.BaseBranch}}-x",
		BaseBranch:         longBase,
	}
	normBranch, err := BuildBranchName(normSettings)
	require.NoError(err)
	// Using raw field (preserves full value before final normalization)
	rawSettings := &BranchNameBuilderSettings{
		BranchNameTemplate: "{{.BaseBranchRaw}}-x",
		BaseBranch:         longBase,
	}
	rawBranch, err := BuildBranchName(rawSettings)
	require.NoError(err)

	// Normalized version should be shorter (pre-truncated) and not contain the full long base
	assert.Equal("thisisaver-ntychars-x", normBranch)
	assert.Equal(NormalizeString(longBase, 0)+"-x", rawBranch)
}
