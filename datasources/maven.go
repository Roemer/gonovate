package datasources

import (
	"encoding/xml"
	"gonovate/core"
	"log/slog"
	"net/url"
)

type MavenDatasource struct {
	datasourceBase
}

func NewMavenDatasource(logger *slog.Logger) IDatasource {
	newDatasource := &MavenDatasource{
		datasourceBase: datasourceBase{
			logger: logger,
			name:   core.DATASOURCE_TYPE_MAVEN,
		},
	}
	newDatasource.impl = newDatasource
	return newDatasource
}

func (ds *MavenDatasource) getReleases(packageSettings *core.PackageSettings, hostRules []*core.HostRule) ([]*core.ReleaseInfo, error) {
	repositoryUrl := "https://repo.maven.apache.org/maven2"

	// Get XML-Metadata
	// TODO: Build from packageSettings.PackageName
	packageMetadataUrl, err := url.JoinPath(repositoryUrl, "org/apache/maven/maven/maven-metadata.xml")
	if err != nil {
		return nil, err
	}

	// Download the file
	metadataFileBytes, err := core.HttpUtil.DownloadToMemory(packageMetadataUrl)
	if err != nil {
		return nil, err
	}

	// Parse the versions
	var mavenReleases *mavenReleases
	if err := xml.Unmarshal(metadataFileBytes, &mavenReleases); err != nil {
		return nil, err
	}

	// Convert all entries to objects
	releases := []*core.ReleaseInfo{}
	for _, entry := range mavenReleases.Versioning.Versions {
		versionString := entry
		releases = append(releases, &core.ReleaseInfo{
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
