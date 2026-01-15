package datasources

import (
	"errors"
	"fmt"
	"log/slog"
	"regexp"
	"strings"

	"github.com/roemer/gonovate/pkg/common"
	"github.com/roemer/gover"
)

type datasourceBase struct {
	datasourceType common.DatasourceType
	logger         *slog.Logger
	impl           common.IDatasource
	settings       *common.DatasourceSettings
}

func newDatasourceBase(datasourceType common.DatasourceType, settings *common.DatasourceSettings) *datasourceBase {
	return &datasourceBase{
		datasourceType: datasourceType,
		logger:         settings.Logger.With(slog.String("datasource", string(datasourceType))),
		settings:       settings,
	}
}

func GetDatasource(datasourceType common.DatasourceType, settings *common.DatasourceSettings) (common.IDatasource, error) {
	switch datasourceType {
	case common.DATASOURCE_TYPE_ANTVERSION:
		return NewAntVersionDatasource(settings), nil
	case common.DATASOURCE_TYPE_ARTIFACTORY:
		return NewArtifactoryDatasource(settings), nil
	case common.DATASOURCE_TYPE_BROWSERVERSION:
		return NewBrowserVersionDatasource(settings), nil
	case common.DATASOURCE_TYPE_DOCKER:
		return NewDockerDatasource(settings), nil
	case common.DATASOURCE_TYPE_GITHUB_RELEASES:
		return NewGitHubReleasesDatasource(settings), nil
	case common.DATASOURCE_TYPE_GITHUB_TAGS:
		return NewGitHubTagsDatasource(settings), nil
	case common.DATASOURCE_TYPE_GITLAB_PACKAGES:
		return NewGitLabPackagesDatasource(settings), nil
	case common.DATASOURCE_TYPE_GOMOD:
		return NewGoModDatasource(settings), nil
	case common.DATASOURCE_TYPE_GOVERSION:
		return NewGoVersionDatasource(settings), nil
	case common.DATASOURCE_TYPE_GRADLEVERSION:
		return NewGradleVersionDatasource(settings), nil
	case common.DATASOURCE_TYPE_HELM:
		return NewHelmDatasource(settings), nil
	case common.DATASOURCE_TYPE_JAVAVERSION:
		return NewJavaVersionDatasource(settings), nil
	case common.DATASOURCE_TYPE_MAVEN:
		return NewMavenDatasource(settings), nil
	case common.DATASOURCE_TYPE_NODEJS:
		return NewNodeJsDatasource(settings), nil
	case common.DATASOURCE_TYPE_NPM:
		return NewNpmDatasource(settings), nil
	}
	return nil, fmt.Errorf("no datasource defined for '%s'", datasourceType)
}

func (ds *datasourceBase) SearchDependencyUpdate(dependency *common.Dependency) (*common.ReleaseInfo, error) {
	ds.logger.Info(fmt.Sprintf("Searching an update for '%s'", dependency.Name))
	skipVersionCheck := false
	if dependency.SkipVersionCheck != nil {
		skipVersionCheck = *dependency.SkipVersionCheck
	}

	// Check if the dependency has a digest
	hasDigest := dependency.HasDigest()

	// Prepare some state variables
	var newRelease *common.ReleaseInfo
	var versionDiffers bool

	// Special condition for when the update version check should be skipped (eg. Docker latest)
	if skipVersionCheck {
		if !hasDigest {
			// No version check and no digest, the dependency cannot have an update
			ds.logger.Warn("Version check is disabled and no digest is defined, skipping check")
			return nil, nil
		}
		// The new release keeps the current version
		newRelease = &common.ReleaseInfo{
			VersionString: dependency.Version,
		}
		versionDiffers = false
	} else {
		// Search for an updated release
		updatedRelease, currentVersion, err := ds.searchUpdatedVersion(dependency)
		if err != nil {
			return nil, err
		}
		if updatedRelease != nil {
			// Assign the new release
			versionDiffers = !updatedRelease.Version.Equals(currentVersion)
			newRelease = updatedRelease
		} else {
			// Use the current version, there is still possibly a digest that can change
			newRelease = &common.ReleaseInfo{
				VersionString: dependency.Version,
			}
			versionDiffers = false
		}
	}

	// Special handling for digest (eg. for Docker)
	digestDiffers := false
	// Check if the dependency has a digest set
	if hasDigest {
		// Get the digest for the maximum valid release from the datasource
		newDigest, err := ds.impl.GetDigest(dependency, newRelease.VersionString)
		if err != nil {
			return nil, err
		}
		// Check if the digest differs
		if dependency.Digest != newDigest {
			digestDiffers = true
		}
		// Make sure the digest is assigned
		newRelease.Digest = newDigest
	}

	// Check if the version and digest is the same
	if !versionDiffers && !digestDiffers {
		ds.logger.Info("No update found")
		return nil, nil
	}

	// It is not the same, return the new version
	changeList := []string{}
	if versionDiffers {
		changeList = append(changeList, newRelease.Version.Raw)
	}
	if digestDiffers {
		changeList = append(changeList, "Digest Changed")
	}
	ds.logger.Info(fmt.Sprintf("Update found: %s", strings.Join(changeList, " / ")))
	return newRelease, nil
}

func (ds *datasourceBase) GetDigest(dependency *common.Dependency, releaseVersion string) (string, error) {
	return "", fmt.Errorf("datasource does not support digests")
}

func (ds *datasourceBase) GetAdditionalData(dependency *common.Dependency, newRelease *common.ReleaseInfo, dataType string) (string, error) {
	if value, ok := newRelease.AdditionalData[dataType]; ok {
		return value, nil
	}
	return "", fmt.Errorf("additional data for '%s' not found in dependency '%s'", dataType, dependency.Name)
}

// Searches for a new version update for the given dependency
func (ds *datasourceBase) searchUpdatedVersion(dependency *common.Dependency) (*common.ReleaseInfo, *gover.Version, error) {
	// Setup everything for the releases lookup
	allowUnstable := false
	if dependency.AllowUnstable != nil {
		allowUnstable = *dependency.AllowUnstable
	}
	ignoreNoneMatching := false
	if dependency.IgnoreNonMatching != nil {
		ignoreNoneMatching = *dependency.IgnoreNonMatching
	}
	if dependency.Versioning == "" {
		return nil, nil, fmt.Errorf("empty 'versioning' regexp")
	}
	versionRegex, err := regexp.Compile(dependency.Versioning)
	if err != nil {
		return nil, nil, fmt.Errorf("failed parsing the 'versioning' regexp '%s': %w", dependency.Versioning, err)
	}
	var extractVersionRegex *regexp.Regexp
	if dependency != nil && len(dependency.ExtractVersion) > 0 {
		extractVersionRegex, err = regexp.Compile(dependency.ExtractVersion)
		if err != nil {
			return nil, nil, fmt.Errorf("failed parsing the 'extractVersion' regexp '%s': %w", dependency.ExtractVersion, err)
		}
	}

	// Try get releases from the cache or look them up from remote
	var rawReleases []*common.ReleaseInfo = nil
	cache := ds.settings.Cache
	cacheIdentifier := dependency.Name
	if cache != nil {
		// Fetch from cache
		rawReleases, err = cache.Get(ds.datasourceType, cacheIdentifier)
		if err != nil {
			// Cache failed
			return nil, nil, err
		}
	}
	if rawReleases != nil {
		ds.logger.Debug("Returned releases from cache")
	} else {
		// No data from cache, fetch new data
		ds.logger.Debug("Lookup releases from remote")
		rawReleases, err = ds.impl.GetReleases(dependency)
		if err != nil {
			return nil, nil, err
		}
		// Store in cache
		if cache != nil {
			if err := cache.Set(ds.datasourceType, cacheIdentifier, rawReleases); err != nil {
				return nil, nil, err
			}
		}
	}

	// Convert the raw releases to parsed versions
	avaliableReleases := []*common.ReleaseInfo{}
	for _, release := range rawReleases {
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

	if len(avaliableReleases) == 0 {
		ds.logger.Warn("No releases found to check for versions")
		return nil, nil, nil
	}

	// Parse the current version
	currentVersion, err := gover.ParseVersionFromRegex(dependency.Version, versionRegex)
	if err != nil {
		return nil, nil, fmt.Errorf("failed parsing the current version '%s': %w", dependency.Version, err)
	}
	// Get the reference version to search
	refVersion, err := getReferenceVersionForUpdateType(dependency.MaxUpdateType, currentVersion)
	if err != nil {
		return nil, nil, err
	}

	// Search for an update
	newRelease := gover.FindMaxGeneric(avaliableReleases, func(x *common.ReleaseInfo) *gover.Version { return x.Version }, refVersion, !allowUnstable)

	// Early exit if no release was found at all
	if newRelease == nil {
		ds.logger.Warn("No valid releases found within the desired limits")
		return nil, nil, nil
	}

	// The current version somehow is bigger than the maximum found version
	if newRelease.Version.LessThan(currentVersion) {
		ds.logger.Warn(fmt.Sprintf("Max found version is less than the current version: %s < %s", newRelease.VersionString, currentVersion.Raw))
		return nil, nil, nil
	}

	return newRelease, currentVersion, nil
}

func (ds *datasourceBase) getRegistryUrl(baseUrl string, customRegistryUrls []string) string {
	if len(customRegistryUrls) > 0 {
		baseUrl = customRegistryUrls[0]
		ds.logger.Debug(fmt.Sprintf("Using custom registry url: %s", baseUrl))
	}
	baseUrl = strings.TrimSuffix(baseUrl, "/")
	return baseUrl
}

func (ds *datasourceBase) getHostRuleForHost(host string) *common.HostRule {
	if ds.settings != nil {
		for _, hostRule := range ds.settings.HostRules {
			if strings.Contains(host, hostRule.MatchHost) {
				return hostRule
			}
		}
	}
	return nil
}

func getReferenceVersionForUpdateType(updateType common.UpdateType, currentVersion *gover.Version) (*gover.Version, error) {
	if updateType == common.UPDATE_TYPE_MAJOR {
		return gover.EmptyVersion, nil
	}
	if updateType == common.UPDATE_TYPE_MINOR {
		return gover.ParseSimple(currentVersion.Major()), nil
	}
	if updateType == common.UPDATE_TYPE_PATCH {
		return gover.ParseSimple(currentVersion.Major(), currentVersion.Minor()), nil
	}
	return nil, fmt.Errorf("missing updateType")
}
