package datasources

import (
	"encoding/json"
	"fmt"
	"gonovate/core"
	"log/slog"
	"net/url"
	"regexp"

	"github.com/roemer/gover"
)

type NodeJsDatasource struct {
	datasourcesBase
}

func NewNodeJsDatasource(logger *slog.Logger) *NodeJsDatasource {
	newDatasource := &NodeJsDatasource{}
	newDatasource.logger = logger
	return newDatasource
}

func (ds *NodeJsDatasource) SearchPackageUpdate(packageName string, currentVersion string, packageSettings *core.PackageSettings, hostRules []*core.HostRule) (string, bool, error) {
	baseUrl := "https://nodejs.org/dist"
	if len(packageSettings.RegistryUrls) > 0 {
		baseUrl = packageSettings.RegistryUrls[0]
		ds.logger.Debug(fmt.Sprintf("Using custom registry url: %s", baseUrl))
	}
	indexFilePath := "index.json"
	useUnstable := false
	if packageSettings.UseUnstable != nil {
		useUnstable = *packageSettings.UseUnstable
	}

	// Download the index file
	downloadUrl, err := url.JoinPath(baseUrl, indexFilePath)
	if err != nil {
		return "", false, err
	}
	indexFileBytes, err := core.HttpUtil.DownloadToMemory(downloadUrl)
	if err != nil {
		return "", false, err
	}

	// Parse the data as json
	var jsonData []map[string]interface{}
	if err := json.Unmarshal(indexFileBytes, &jsonData); err != nil {
		return "", false, err
	}

	// Convert all entries to objects
	versionRegex := regexp.MustCompile(`(?m:^v(?P<d1>\d+)(?:\.(?P<d2>\d+))?(?:\.(?P<d3>\d+))?$)`)
	allVersions := []*gover.Version{}
	ltsVersions := []*gover.Version{}
	for _, entry := range jsonData {
		versionString := entry["version"].(string)
		ltsValue := entry["lts"]
		version := gover.MustParseVersionFromRegex(versionString, versionRegex)
		allVersions = append(allVersions, version)
		if ltsValue != false {
			ltsVersions = append(ltsVersions, version)
		}
	}

	curr, err := gover.ParseVersionFromRegex(currentVersion, versionRegex)
	if err != nil {
		return "", false, err
	}
	refVersion := ds.getReferenceVersionForUpdateType(packageSettings.MaxUpdateType, curr)
	// Search for an update
	var maxValidVersion *gover.Version
	if useUnstable {
		maxValidVersion = gover.FindMax(allVersions, refVersion, true)
	} else {
		maxValidVersion = gover.FindMax(ltsVersions, refVersion, true)
	}

	if maxValidVersion.Equals(curr) {
		ds.logger.Debug("Found no new version")
		return "", false, nil
	}

	ds.logger.Info(fmt.Sprintf("Found a new version: %s", maxValidVersion.Original))

	return maxValidVersion.Original, true, nil
}
