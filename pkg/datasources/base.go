package datasources

import (
	"errors"
	"fmt"
	"log/slog"
	"regexp"
	"slices"
	"strings"
	"time"

	"github.com/roemer/gonovate/pkg/cache"
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
	case common.DATASOURCE_TYPE_GIT_TAGS:
		return NewGitTagsDatasource(settings), nil
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

func (ds *datasourceBase) SearchDependencyUpdates(dependency *common.Dependency) ([]*common.ReleaseInfo, error) {
	ds.logger.Info(fmt.Sprintf("Searching an update for '%s'", dependency.Name))
	skipVersionCheck := false
	if dependency.SkipVersionCheck != nil {
		skipVersionCheck = *dependency.SkipVersionCheck
	}

	// The process is as follows:
	// 1. If skipVersionCheck is true, only check for a digest update for the current version if a digest is defined.
	// 2. If skipVersionCheck is false, search for new releases.
	// 2.1 If there are no updates, also check for a digest update for the current version if a digest is defined.
	// 2.2 If there are updates, use them but also check for digest updates for them.

	// Check if the dependency has a digest
	hasDigest := dependency.HasDigest()

	// Prepare some state variables
	updates := []*common.ReleaseInfo{}
	var currentVersion *gover.Version = nil

	// Special condition for when the update version check should be skipped (eg. Docker latest)
	if skipVersionCheck {
		if !hasDigest {
			// No version check and no digest, the dependency cannot have an update
			ds.logger.Warn("Version check is disabled and no digest is defined, skipping check")
			return nil, nil
		}
		// Use the current version, the digest might still be updated
		updates = append(updates, &common.ReleaseInfo{
			VersionString: dependency.Version,
		})
	} else {
		// Search for new releases for the dependency
		foundUpdates, cv, err := ds.searchUpdatedVersion(dependency)
		if err != nil {
			return nil, err
		}
		currentVersion = cv
		if len(foundUpdates) > 0 {
			// Assign the new release
			updates = append(updates, foundUpdates...)
		} else {
			// Use the current version, the digest might still be updated
			updates = append(updates, &common.ReleaseInfo{
				VersionString: dependency.Version,
			})
		}
	}

	// Special handling for digest (eg. for Docker) if a digest is set
	if hasDigest {
		for _, newUpdate := range updates {
			// Get the digest for the new release version from the datasource
			newDigest, err := ds.impl.GetDigest(dependency, newUpdate.VersionString)
			if err != nil {
				return nil, err
			}
			// Make sure the digest is assigned
			newUpdate.Digest = newDigest
		}
	}

	// Verify if there is a version or digest update
	changedUpdates := []*common.ReleaseInfo{}
	for _, newUpdate := range updates {
		versionDiffers := false
		digestDiffers := false

		if newUpdate.Version != nil && currentVersion != nil {
			versionDiffers = !newUpdate.Version.Equals(currentVersion)
		}
		if hasDigest {
			digestDiffers = newUpdate.Digest != dependency.Digest
		}

		// Check if the version and digest is the same
		if !versionDiffers && !digestDiffers {
			// If so, skip this release
			continue
		}

		// Check if a release with the same version/digest already exists
		// And if so, use the one with the lowest update-type
		sameUpdateIndex := slices.IndexFunc(changedUpdates, func(update *common.ReleaseInfo) bool {
			return update.VersionString == newUpdate.VersionString && update.Digest == newUpdate.Digest
		})
		if sameUpdateIndex >= 0 {
			existingUpdate := changedUpdates[sameUpdateIndex]
			// If the existing update has a higher update type than the new update, replace it with the new update
			if newUpdate.UpdateType.IsLessSignificant(existingUpdate.UpdateType) {
				changedUpdates[sameUpdateIndex] = newUpdate
			}
			continue
		}

		// Else add the update to the list of changed updates
		changedUpdates = append(changedUpdates, newUpdate)

		// Write an info about the new version
		changeList := []string{}
		if versionDiffers {
			changeList = append(changeList, newUpdate.Version.Raw)
		}
		if digestDiffers {
			changeList = append(changeList, "Digest Changed")
		}
		if newUpdate.UpdateType != "" {
			changeList = append(changeList, string(newUpdate.UpdateType))
		}
		ds.logger.Info(fmt.Sprintf("Update found: %s", strings.Join(changeList, " / ")))
	}

	// Write an info if no update at all was found
	if len(changedUpdates) == 0 {
		ds.logger.Info("No update found")
		return nil, nil
	}

	// Return the new updates
	return changedUpdates, nil
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

// Searches for new releases for each update type for the dependency.
func (ds *datasourceBase) searchUpdatedVersion(dependency *common.Dependency) ([]*common.ReleaseInfo, *gover.Version, error) {
	// Ensure that at least one update type is configured
	if len(dependency.UpdateTypes) == 0 {
		return nil, nil, fmt.Errorf("no update types configured for dependency")
	}

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
	cacheIdentifier := fmt.Sprintf("rel/%s/%s", ds.datasourceType, cache.NormalizeFilePath(dependency.Name, false))
	if ds.settings.Cache != nil {
		// Fetch from cache
		if releasesFromCache, exists, err := ds.settings.Cache.Get(cacheIdentifier); err != nil {
			// Cache failed
			return nil, nil, err
		} else if exists {
			rawReleases = releasesFromCache
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
		if ds.settings.Cache != nil {
			if err := ds.settings.Cache.Set(cacheIdentifier, rawReleases, time.Hour); err != nil {
				return nil, nil, err
			}
		}
	}

	// Convert the raw releases to parsed versions
	availableReleases := []*common.ReleaseInfo{}
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
		availableReleases = append(availableReleases, release)
	}

	if len(availableReleases) == 0 {
		ds.logger.Warn("No releases found to check for versions")
		return nil, nil, nil
	}

	// Parse the current version
	currentVersion, err := gover.ParseVersionFromRegex(dependency.Version, versionRegex)
	if err != nil {
		return nil, nil, fmt.Errorf("failed parsing the current version '%s': %w", dependency.Version, err)
	}

	// Search for new releases for each update type and collect them in a list
	updates := []*common.ReleaseInfo{}
	for _, currentUpdateType := range dependency.UpdateTypes {
		// Get the reference version for the update type
		refVersion, err := getReferenceVersionForUpdateType(currentUpdateType, currentVersion)
		if err != nil {
			return nil, nil, err
		}

		// Search for an update
		newUpdate := gover.FindMaxGeneric(availableReleases, func(x *common.ReleaseInfo) *gover.Version { return x.Version }, refVersion, !allowUnstable)

		// Early exit if no update was found at all
		if newUpdate == nil {
			ds.logger.Warn("No valid releases found within the desired limits")
			continue
		}

		// The current version somehow is bigger than the maximum found version
		if newUpdate.Version.LessThan(currentVersion) {
			ds.logger.Warn(fmt.Sprintf("Max found version is less than the current version: %s < %s", newUpdate.VersionString, currentVersion.Raw))
			continue
		}

		// Check if an existing update with the same version already exists
		existingUpdateIndex := slices.IndexFunc(updates, func(release *common.ReleaseInfo) bool {
			return release.Version.Equals(newUpdate.Version)
		})
		if existingUpdateIndex >= 0 {
			existingUpdate := updates[existingUpdateIndex]
			// If the existing update has a higher update type than the new update, update the update type
			if currentUpdateType.IsLessSignificant(existingUpdate.UpdateType) {
				updates[existingUpdateIndex].UpdateType = currentUpdateType
			}
			continue
		}

		// A new update was found, add a clone of it to the list of updates (because the same releaseInfo can be used for multiple update types)
		newUpdateClone := newUpdate.Clone()
		newUpdateClone.UpdateType = currentUpdateType
		updates = append(updates, newUpdateClone)
	}

	return updates, currentVersion, nil
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
