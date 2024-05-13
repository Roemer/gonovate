package datasources

import (
	"fmt"
	"gonovate/core"
	"log/slog"
	"regexp"

	"github.com/roemer/gover"
)

type datasourceBase struct {
	logger *slog.Logger
	name   string
}

type datasource interface {
	GetLogger() *slog.Logger
	GetName() string
	GetReleases(packageSettings *core.PackageSettings, hostRules []*core.HostRule) ([]*core.ReleaseInfo, error)
}

func (ds *datasourceBase) GetLogger() *slog.Logger {
	return ds.logger
}

func (ds *datasourceBase) GetName() string {
	return ds.name
}

func SearchPackageUpdate(ds datasource, currentVersion string, packageSettings *core.PackageSettings, hostRules []*core.HostRule) (string, bool, error) {
	// Setup
	name := ds.GetName()
	logger := ds.GetLogger()
	cacheIdentifier := name + "|" + packageSettings.PackageName
	allowUnstable := false
	if packageSettings.AllowUnstable != nil {
		allowUnstable = *packageSettings.AllowUnstable
	}
	ignoreNoneMatching := false
	if packageSettings.IgnoreNonMatching != nil {
		ignoreNoneMatching = *packageSettings.IgnoreNonMatching
	}
	versionRegex, err := regexp.Compile(packageSettings.Versioning)
	if err != nil {
		return "", false, fmt.Errorf("failed parsing the 'versioning' regexp '%s': %w", packageSettings.Versioning, err)
	}
	var extractVersionRegex *regexp.Regexp
	if packageSettings != nil && len(packageSettings.ExtractVersion) > 0 {
		extractVersionRegex, err = regexp.Compile(packageSettings.ExtractVersion)
		if err != nil {
			return "", false, fmt.Errorf("failed parsing the 'extractVersion' regexp '%s': %w", packageSettings.ExtractVersion, err)
		}
	}

	// Try get releases from the cache
	avaliableReleases := core.DatasourceCache.GetCache(name, cacheIdentifier)
	if avaliableReleases == nil {
		// No data in cache, fetch new data
		logger.Debug("Lookup releases from remote")
		releases, err := ds.GetReleases(packageSettings, hostRules)
		if err != nil {
			return "", false, err
		}
		// Convert the raw strings to versions
		avaliableReleases = []*core.ReleaseInfo{}
		for _, release := range releases {
			// Extract the version number from the raw string if needed
			if extractVersionRegex != nil {
				m := extractVersionRegex.FindStringSubmatch(release.VersionString)
				if m == nil {
					if ignoreNoneMatching {
						continue
					} else {
						return "", false, fmt.Errorf("could not extract version from '%s'", release.VersionString)
					}
				}
				// Continue with only the matched part
				release.VersionString = m[1]
			}
			version, err := gover.ParseVersionFromRegex(release.VersionString, versionRegex)
			if err != nil {
				if ignoreNoneMatching {
					continue
				}
				return "", false, fmt.Errorf("failed parsing the version from '%s': %w", release.VersionString, err)
			}
			release.Version = version
			avaliableReleases = append(avaliableReleases, release)
		}
		// Store in cache
		core.DatasourceCache.SetCache(name, cacheIdentifier, avaliableReleases)
	} else {
		logger.Debug("Returned releases from cache")
	}

	// Parse the current version
	curr, err := gover.ParseVersionFromRegex(currentVersion, versionRegex)
	if err != nil {
		return "", false, fmt.Errorf("failed parsing the current version '%s: %w", currentVersion, err)
	}
	// Get the reference version to search
	refVersion := getReferenceVersionForUpdateType(packageSettings.MaxUpdateType, curr)

	// Search for an update
	maxValidVersion := gover.FindMaxGeneric(avaliableReleases, func(x *core.ReleaseInfo) *gover.Version { return x.Version }, refVersion, !allowUnstable)

	// Check if the version is the same
	if maxValidVersion.Equals(curr) {
		logger.Debug("Found no new version")
		return "", false, nil
	}

	// It is not the same, return the new version
	logger.Info(fmt.Sprintf("Found a new version: %s", maxValidVersion.Original))
	return maxValidVersion.Original, true, nil
}

func GetDatasource(logger *slog.Logger, datasource string) (datasource, error) {
	if datasource == core.DATASOURCE_TYPE_ARTIFACTORY {
		return NewArtifactoryDatasource(logger), nil
	}
	if datasource == core.DATASOURCE_TYPE_DOCKER {
		return NewDockerDatasource(logger), nil
	}
	if datasource == core.DATASOURCE_TYPE_NODEJS {
		return NewNodeJsDatasource(logger), nil
	}
	return nil, fmt.Errorf("no datasource defined for '%s'", datasource)
}

func getReferenceVersionForUpdateType(updateType string, currentVersion *gover.Version) *gover.Version {
	if updateType == core.UPDATE_TYPE_MAJOR {
		return gover.EmptyVersion
	}
	if updateType == core.UPDATE_TYPE_MINOR {
		return gover.ParseSimple(currentVersion.Major())
	}
	if updateType == core.UPDATE_TYPE_PATCH {
		return gover.ParseSimple(currentVersion.Major(), currentVersion.Minor())
	}
	return nil
}
