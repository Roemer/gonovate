package datasources

import (
	"errors"
	"fmt"
	"gonovate/core"
	"log/slog"
	"regexp"

	"github.com/roemer/gover"
)

type IDatasource interface {
	getReleases(packageSettings *core.PackageSettings, hostRules []*core.HostRule) ([]*core.ReleaseInfo, error)
	SearchPackageUpdate(currentVersion string, packageSettings *core.PackageSettings, hostRules []*core.HostRule) (*core.ReleaseInfo, *gover.Version, error)
}

type datasourceBase struct {
	logger *slog.Logger
	name   string
	impl   IDatasource
}

func (ds *datasourceBase) SearchPackageUpdate(currentVersion string, packageSettings *core.PackageSettings, hostRules []*core.HostRule) (*core.ReleaseInfo, *gover.Version, error) {
	// Setup
	cacheIdentifier := ds.name + "|" + packageSettings.PackageName
	allowUnstable := false
	if packageSettings.AllowUnstable != nil {
		allowUnstable = *packageSettings.AllowUnstable
	}
	ignoreNoneMatching := false
	if packageSettings.IgnoreNonMatching != nil {
		ignoreNoneMatching = *packageSettings.IgnoreNonMatching
	}
	if packageSettings.Versioning == "" {
		return nil, nil, fmt.Errorf("empty 'versioning' regexp")
	}
	versionRegex, err := regexp.Compile(packageSettings.Versioning)
	if err != nil {
		return nil, nil, fmt.Errorf("failed parsing the 'versioning' regexp '%s': %w", packageSettings.Versioning, err)
	}
	var extractVersionRegex *regexp.Regexp
	if packageSettings != nil && len(packageSettings.ExtractVersion) > 0 {
		extractVersionRegex, err = regexp.Compile(packageSettings.ExtractVersion)
		if err != nil {
			return nil, nil, fmt.Errorf("failed parsing the 'extractVersion' regexp '%s': %w", packageSettings.ExtractVersion, err)
		}
	}

	// Try get releases from the cache
	avaliableReleases := core.DatasourceCache.GetCache(ds.name, cacheIdentifier)
	if avaliableReleases == nil {
		// No data in cache, fetch new data
		ds.logger.Debug("Lookup releases from remote")
		releases, err := ds.impl.getReleases(packageSettings, hostRules)
		if err != nil {
			return nil, nil, err
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
						return nil, nil, fmt.Errorf("could not extract version from '%s'", release.VersionString)
					}
				}
				// Continue with only the matched part
				release.VersionString = m[1]
			}
			version, err := gover.ParseVersionFromRegex(release.VersionString, versionRegex)
			if err != nil {
				if errors.Is(err, gover.ErrNoMatch) && ignoreNoneMatching {
					ds.logger.Debug(fmt.Sprintf("Ignoring non matching version: %s", release.VersionString))
					continue
				}
				return nil, nil, fmt.Errorf("failed parsing the version from '%s': %w", release.VersionString, err)
			}
			release.Version = version
			avaliableReleases = append(avaliableReleases, release)
		}
		// Store in cache
		core.DatasourceCache.SetCache(ds.name, cacheIdentifier, avaliableReleases)
	} else {
		ds.logger.Debug("Returned releases from cache")
	}

	if len(avaliableReleases) == 0 {
		ds.logger.Warn("No releases found to check for versions")
		return nil, nil, nil
	}

	// Parse the current version
	curr, err := gover.ParseVersionFromRegex(currentVersion, versionRegex)
	if err != nil {
		return nil, nil, fmt.Errorf("failed parsing the current version '%s': %w", currentVersion, err)
	}
	// Get the reference version to search
	refVersion, err := getReferenceVersionForUpdateType(packageSettings.MaxUpdateType, curr)
	if err != nil {
		return nil, nil, err
	}

	// Search for an update
	maxValidRelease := gover.FindMaxGeneric(avaliableReleases, func(x *core.ReleaseInfo) *gover.Version { return x.Version }, refVersion, !allowUnstable)

	if maxValidRelease == nil {
		ds.logger.Warn("No valid releases found within the desired limits")
		return nil, nil, nil
	}

	// Check if the version is the same
	if maxValidRelease.Version.Equals(curr) {
		ds.logger.Debug("Found no new version")
		return nil, nil, nil
	}

	// The current version somehow is bigger than the maximum found version
	if maxValidRelease.Version.LessThan(curr) {
		ds.logger.Warn(fmt.Sprintf("Max found version is less than the current version: %s < %s", maxValidRelease.VersionString, curr.Raw))
		return nil, nil, nil
	}

	// It is not the same, return the new version
	ds.logger.Info(fmt.Sprintf("Found a new version: %s", maxValidRelease.Version.Raw))
	return maxValidRelease, curr, nil
}

func GetDatasource(logger *slog.Logger, datasource string) (IDatasource, error) {
	switch datasource {
	case core.DATASOURCE_TYPE_ARTIFACTORY:
		return NewArtifactoryDatasource(logger), nil
	case core.DATASOURCE_TYPE_DOCKER:
		return NewDockerDatasource(logger), nil
	case core.DATASOURCE_TYPE_GOVERSION:
		return NewGoVersionDatasource(logger), nil
	case core.DATASOURCE_TYPE_NODEJS:
		return NewNodeJsDatasource(logger), nil
	case core.DATASOURCE_TYPE_NPM:
		return NewNpmDatasource(logger), nil
	}
	return nil, fmt.Errorf("no datasource defined for '%s'", datasource)
}

func getReferenceVersionForUpdateType(updateType string, currentVersion *gover.Version) (*gover.Version, error) {
	if updateType == core.UPDATE_TYPE_MAJOR {
		return gover.EmptyVersion, nil
	}
	if updateType == core.UPDATE_TYPE_MINOR {
		return gover.ParseSimple(currentVersion.Major()), nil
	}
	if updateType == core.UPDATE_TYPE_PATCH {
		return gover.ParseSimple(currentVersion.Major(), currentVersion.Minor()), nil
	}
	return nil, fmt.Errorf("missing updateType")
}
