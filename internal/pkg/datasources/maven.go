package datasources

import (
	"encoding/xml"
	"log/slog"
	"net/url"
	"strings"

	"github.com/roemer/gonovate/internal/pkg/config"
	"github.com/roemer/gonovate/internal/pkg/shared"
)

type MavenDatasource struct {
	datasourceBase
}

func NewMavenDatasource(logger *slog.Logger, config *config.RootConfig) IDatasource {
	newDatasource := &MavenDatasource{
		datasourceBase: datasourceBase{
			logger: logger,
			name:   shared.DATASOURCE_TYPE_MAVEN,
			Config: config,
		},
	}
	newDatasource.impl = newDatasource
	return newDatasource
}

func (ds *MavenDatasource) getReleases(dependency *shared.Dependency) ([]*shared.ReleaseInfo, error) {
	registryUrl := ds.getRegistryUrl("https://repo.maven.apache.org/maven2", dependency.RegistryUrls)

	// Get XML-Metadata
	packageParts := strings.Split(dependency.Name, ":")
	group := packageParts[0]
	pkg := packageParts[1]
	packageMetadataUrl, err := url.JoinPath(registryUrl, strings.ReplaceAll(group, ".", "/"), pkg, "maven-metadata.xml")
	if err != nil {
		return nil, err
	}

	// Download the file
	metadataFileBytes, err := shared.HttpUtil.DownloadToMemory(packageMetadataUrl)
	if err != nil {
		return nil, err
	}

	// Parse the versions
	var mavenReleases *mavenReleases
	if err := xml.Unmarshal(metadataFileBytes, &mavenReleases); err != nil {
		return nil, err
	}

	// Convert all entries to objects
	releases := []*shared.ReleaseInfo{}
	for _, entry := range mavenReleases.Versioning.Versions {
		versionString := entry
		releases = append(releases, &shared.ReleaseInfo{
			VersionString: versionString,
		})
	}
	return releases, nil
}

type mavenReleases struct {
	GroupId    string `xml:"groupId"`
	ArtifactId string `xml:"artifactId"`
	Versioning struct {
		Latest   string   `xml:"latest"`
		Release  string   `xml:"release"`
		Versions []string `xml:"versions>version"`
	} `xml:"versioning"`
}
