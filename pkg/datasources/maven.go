package datasources

import (
	"encoding/xml"
	"net/url"
	"strings"

	"github.com/roemer/gonovate/pkg/common"
)

type MavenDatasource struct {
	*datasourceBase
}

func NewMavenDatasource(settings *common.DatasourceSettings) common.IDatasource {
	newDatasource := &MavenDatasource{
		datasourceBase: newDatasourceBase(common.DATASOURCE_TYPE_MAVEN, settings),
	}
	newDatasource.impl = newDatasource
	return newDatasource
}

func (ds *MavenDatasource) GetReleases(dependency *common.Dependency) ([]*common.ReleaseInfo, error) {
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
	metadataFileBytes, err := common.HttpUtil.DownloadToMemory(packageMetadataUrl)
	if err != nil {
		return nil, err
	}

	// Parse the versions
	var mavenReleases *mavenReleases
	if err := xml.Unmarshal(metadataFileBytes, &mavenReleases); err != nil {
		return nil, err
	}

	// Convert all entries to objects
	releases := []*common.ReleaseInfo{}
	for _, entry := range mavenReleases.Versioning.Versions {
		versionString := entry
		releases = append(releases, &common.ReleaseInfo{
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
