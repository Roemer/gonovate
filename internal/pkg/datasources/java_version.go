package datasources

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"strings"

	"github.com/roemer/gonovate/internal/pkg/config"
	"github.com/roemer/gonovate/internal/pkg/shared"
)

type JavaVersionDatasource struct {
	datasourceBase
}

func NewJavaVersionDatasource(logger *slog.Logger, config *config.RootConfig) IDatasource {
	newDatasource := &JavaVersionDatasource{
		datasourceBase: datasourceBase{
			logger: logger,
			name:   shared.DATASOURCE_TYPE_GOVERSION,
			Config: config,
		},
	}
	newDatasource.impl = newDatasource
	return newDatasource
}

func (ds *JavaVersionDatasource) getReleases(dependency *shared.Dependency) ([]*shared.ReleaseInfo, error) {
	registryUrl := ds.getRegistryUrl("https://api.adoptium.net", dependency.RegistryUrls)

	// Get the type of java that is requested
	javaType := "jdk"
	if strings.HasSuffix(dependency.Name, "jre") {
		javaType = "jre"
	}

	// Build the url and get the data
	downloadUrl := fmt.Sprintf("%s/v3/info/release_versions", registryUrl)
	downloadUrl += fmt.Sprintf("?page_size=50&image_type=%s&project=jdk&release_type=ga&sort_method=DATE&sort_order=DESC", javaType)

	releases := []*shared.ReleaseInfo{}
	for {
		// Prepare the request
		req, err := http.NewRequest(http.MethodGet, downloadUrl, nil)
		if err != nil {
			return nil, err
		}
		// Perform the request
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()
		if resp.StatusCode != 200 {
			return nil, fmt.Errorf("failed with statuscode %d", resp.StatusCode)
		}

		// Parse the data as json
		var jsonData map[string]interface{}
		if err := json.NewDecoder(resp.Body).Decode(&jsonData); err != nil {
			return nil, err
		}

		// Convert all entries to objects
		for _, entry := range jsonData["versions"].([]interface{}) {
			versionString := entry.(map[string]interface{})["semver"].(string)
			releases = append(releases, &shared.ReleaseInfo{
				VersionString: versionString,
			})
		}

		// Check for the next page link
		if nextPageUrl, err := shared.HttpUtil.GetNextPageURL(resp); err != nil {
			return nil, err
		} else if nextPageUrl == nil {
			// No next page
			break
		} else {
			// There is a next page
			downloadUrl = nextPageUrl.String()
		}
	}

	return releases, nil
}
