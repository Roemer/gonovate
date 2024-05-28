package core

const (
	PLATFORM_TYPE_GIT    string = "git"
	PLATFORM_TYPE_GITHUB string = "github"
	PLATFORM_TYPE_GITLAB string = "gitlab"
	PLATFORM_TYPE_NOOP   string = "noop"
)

const (
	MANAGER_TYPE_INLINE string = "inline"
	MANAGER_TYPE_REGEX  string = "regex"
)

const (
	DATASOURCE_TYPE_ARTIFACTORY string = "artifactory"
	DATASOURCE_TYPE_DOCKER      string = "docker"
	DATASOURCE_TYPE_GOVERSION   string = "go-version"
	DATASOURCE_TYPE_MAVEN       string = "maven"
	DATASOURCE_TYPE_NODEJS      string = "nodejs"
	DATASOURCE_TYPE_NPM         string = "npm"
)

const (
	UPDATE_TYPE_MAJOR string = "major"
	UPDATE_TYPE_MINOR string = "minor"
	UPDATE_TYPE_PATCH string = "patch"
)
