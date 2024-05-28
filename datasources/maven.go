package datasources

import (
	"encoding/xml"
	"fmt"
	"gonovate/core"
	"log/slog"
	"net/url"
	"strings"
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
	if len(packageSettings.RegistryUrls) > 0 {
		repositoryUrl = packageSettings.RegistryUrls[0]
		ds.logger.Debug(fmt.Sprintf("Using custom registry url: %s", repositoryUrl))
	}

	// Get XML-Metadata
	packageParts := strings.Split(packageSettings.PackageName, ":")
	group := packageParts[0]
	pkg := packageParts[1]
	packageMetadataUrl, err := url.JoinPath(repositoryUrl, strings.ReplaceAll(group, ".", "/"), pkg, "maven-metadata.xml")
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
