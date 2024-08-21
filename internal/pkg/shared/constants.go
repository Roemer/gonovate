package shared

type PlatformType string

const (
	PLATFORM_TYPE_GIT    PlatformType = "git"
	PLATFORM_TYPE_GITHUB PlatformType = "github"
	PLATFORM_TYPE_GITLAB PlatformType = "gitlab"
	PLATFORM_TYPE_NOOP   PlatformType = "noop"
)

type ManagerType string

const (
	MANAGER_TYPE_INLINE     ManagerType = "inline"
	MANAGER_TYPE_REGEX      ManagerType = "regex"
	MANAGER_TYPE_DOCKERFILE ManagerType = "dockerfile"
	MANAGER_TYPE_GOMOD      ManagerType = "go-mod"
)

type DatasourceType string

const (
	DATASOURCE_TYPE_ARTIFACTORY     DatasourceType = "artifactory"
	DATASOURCE_TYPE_DOCKER          DatasourceType = "docker"
	DATASOURCE_TYPE_GITHUB_RELEASES DatasourceType = "github-releases"
	DATASOURCE_TYPE_GITHUB_TAGS     DatasourceType = "github-tags"
	DATASOURCE_TYPE_GOMOD           DatasourceType = "go-mod"
	DATASOURCE_TYPE_GOVERSION       DatasourceType = "go-version"
	DATASOURCE_TYPE_GRADLEVERSION   DatasourceType = "gradle-version"
	DATASOURCE_TYPE_JAVAVERSION     DatasourceType = "java-version"
	DATASOURCE_TYPE_MAVEN           DatasourceType = "maven"
	DATASOURCE_TYPE_NODEJS          DatasourceType = "nodejs"
	DATASOURCE_TYPE_NPM             DatasourceType = "npm"
)

type UpdateType string

const (
	UPDATE_TYPE_MAJOR UpdateType = "major"
	UPDATE_TYPE_MINOR UpdateType = "minor"
	UPDATE_TYPE_PATCH UpdateType = "patch"
)
