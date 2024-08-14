package datasources

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"regexp"
	"strings"

	"github.com/roemer/gonovate/internal/pkg/config"
	"github.com/roemer/gonovate/internal/pkg/shared"
)

type DockerDatasource struct {
	datasourceBase
}

func NewDockerDatasource(logger *slog.Logger, config *config.RootConfig) IDatasource {
	newDatasource := &DockerDatasource{
		datasourceBase: datasourceBase{
			logger: logger,
			name:   shared.DATASOURCE_TYPE_DOCKER,
			Config: config,
		},
	}
	newDatasource.impl = newDatasource
	return newDatasource
}

var dockerIoRegex = regexp.MustCompile(`^(https?://)?([a-zA-Z-_0-9\.]*docker\.io)($|/)`)
var httpSchemeRegex = regexp.MustCompile(`^https?://(.*)`)

func (ds *DockerDatasource) getReleases(dependency *shared.Dependency) ([]*shared.ReleaseInfo, error) {
	// Prepare the registry host
	customRegistryUrl := ""
	if len(dependency.RegistryUrls) > 0 {
		customRegistryUrl = dependency.RegistryUrls[0]
	}
	registryUrl, imagePath, err := getDockerRegistry(dependency.Name, customRegistryUrl)
	if err != nil {
		return nil, err
	}

	// Parse the registry url
	baseUrl, err := url.Parse(registryUrl)
	if err != nil {
		return nil, err
	}
	// Add the v2 endpoint
	baseUrl = baseUrl.JoinPath("v2")

	// Get a host rule if any was defined
	relevantHostRule := ds.Config.FilterHostConfigsForHost(baseUrl.Host)

	// Different handling for different sites
	var tags []string
	if strings.Contains(baseUrl.Host, "index.docker.io") {
		tags, err = ds.getTagsForDocker(baseUrl, imagePath, relevantHostRule)
		if err != nil {
			return nil, err
		}
	} else if strings.Contains(baseUrl.Host, "ghcr.io") {
		tags, err = ds.getTagsForGhcr(baseUrl, imagePath, relevantHostRule)
		if err != nil {
			return nil, err
		}
	} else if strings.Contains(baseUrl.Host, "gcr.io") {
		tags, err = ds.getTagsForGcr(baseUrl, imagePath, relevantHostRule)
		if err != nil {
			return nil, err
		}
	} else if strings.Contains(baseUrl.Host, "quay.io") {
		// For quay we need a special token
		tags, err = ds.getTagsForQuay(baseUrl, imagePath, relevantHostRule)
		if err != nil {
			return nil, err
		}
	} else {
		// For everything else we just use a bearer token (if provided), eg. Artifactory
		bearerToken := ""
		if relevantHostRule != nil {
			bearerToken = relevantHostRule.TokendExpanded()
		}
		tags, err = ds.getTagsWithToken(baseUrl, imagePath, bearerToken)
		if err != nil {
			return nil, err
		}
	}
	ds.logger.Debug(fmt.Sprintf("Found %d tag(s)", len(tags)))
	releases := []*shared.ReleaseInfo{}
	for _, tag := range tags {
		releases = append(releases, &shared.ReleaseInfo{
			VersionString: tag,
		})
	}
	return releases, nil
}

// Processes the package name and registry url and returns the concrete host and image path
func getDockerRegistry(dependencyName string, registryUrl string) (string, string, error) {
	// Makes sure that the given url (if not empty) has a http/https scheme or it appends https
	if registryUrl != "" && !httpSchemeRegex.MatchString(registryUrl) {
		registryUrl = "https://" + registryUrl
	}
	if registryUrl == "" {
		// Default to the Docker registry
		registryUrl = "https://index.docker.io"
	}

	// Make sure all *.docker.io registries point to index.docker.io
	registryUrl = normalizeDockerIo(registryUrl)
	dependencyName = normalizeDockerIo(dependencyName)

	// If the dependencyName equals the passed registryUrl, move all path parts from the registryUrl to the dependencyName
	simplifiedRegistryUrl := ensureTrailingSlash(removeScheme(registryUrl))
	if simplifiedRegistryUrl != "" && strings.HasPrefix(dependencyName, simplifiedRegistryUrl) {
		var err error
		registryUrl, dependencyName, err = moveRegistryPathToPackage(registryUrl, strings.Replace(dependencyName, simplifiedRegistryUrl, "", 1))
		if err != nil {
			return "", "", err
		}
	} else {
		// Split the dependencyName into parts
		split := strings.Split(dependencyName, "/")
		// Check if the dependencyName seems to contain a host (eg. with a . or a :)
		if len(split) > 1 && strings.ContainsAny(split[0], ".:") {
			// It does, so use it as the host with https
			registryUrl = fmt.Sprintf("https://%s", split[0])
			dependencyName = strings.Join(split[1:], "/")
		} else {
			var err error
			registryUrl, dependencyName, err = moveRegistryPathToPackage(registryUrl, dependencyName)
			if err != nil {
				return "", "", err
			}
		}
	}

	// Special handling for docker.io: if the dependencyName is a single entry, add "library/"
	if dockerIoRegex.MatchString(registryUrl) {
		if !strings.Contains(dependencyName, "/") {
			dependencyName = "library/" + dependencyName
		}
	}

	return registryUrl, dependencyName, nil
}

// Makes sure that docker.io urls (like just docker.io or registry-1.docker.io) are replaced with index.docker.io
func normalizeDockerIo(value string) string {
	return dockerIoRegex.ReplaceAllString(value, "${1}index.docker.io${3}")
}

// Move all path parts (if any) from the registryUrl to the dependencyName
func moveRegistryPathToPackage(registryUrl, dependencyName string) (string, string, error) {
	fullUrl, err := url.JoinPath(registryUrl, dependencyName)
	if err != nil {
		return "", "", err
	}
	url, err := url.Parse(fullUrl)
	if err != nil {
		return "", "", err
	}
	return fmt.Sprintf("%s://%s", url.Scheme, url.Host), url.Path[1:], nil
}

func removeScheme(url string) string {
	return httpSchemeRegex.ReplaceAllString(url, "$1")
}

func ensureTrailingSlash(url string) string {
	if strings.HasPrefix(url, "/") {
		return url
	}
	return url + "/"
}

func (ds *DockerDatasource) getTagsForDocker(baseUrl *url.URL, dependencyName string, hostRule *config.HostRule) ([]string, error) {
	token, err := ds.getJwtToken("https://auth.docker.io/token?service=registry.docker.io&scope=repository:%s:pull", dependencyName, hostRule)
	if err != nil {
		return nil, err
	}
	return ds.getTagsWithToken(baseUrl, dependencyName, token)
}

func (ds *DockerDatasource) getTagsForGhcr(baseUrl *url.URL, dependencyName string, hostRule *config.HostRule) ([]string, error) {
	token, err := ds.getJwtToken("https://ghcr.io/token?service=ghcr.io&scope=repository:%s:pull", dependencyName, hostRule)
	if err != nil {
		return nil, err
	}
	return ds.getTagsWithToken(baseUrl, dependencyName, token)
}

func (ds *DockerDatasource) getTagsForGcr(baseUrl *url.URL, dependencyName string, hostRule *config.HostRule) ([]string, error) {
	token, err := ds.getJwtToken("https://gcr.io/v2/token?service=gcr.io&scope=repository:%s:pull", dependencyName, hostRule)
	if err != nil {
		return nil, err
	}
	return ds.getTagsWithToken(baseUrl, dependencyName, token)
}

func (ds *DockerDatasource) getTagsForQuay(baseUrl *url.URL, dependencyName string, hostRule *config.HostRule) ([]string, error) {
	token, err := ds.getJwtToken("https://quay.io/v2/auth?service=quay.io&scope=repository:%s:pull", dependencyName, hostRule)
	if err != nil {
		return nil, err
	}
	return ds.getTagsWithToken(baseUrl, dependencyName, token)
}

// Creates a request to get a token and returns the token. Uses basic auth uf username/password is provided.
func (ds *DockerDatasource) getJwtToken(authUrl string, dependencyName string, hostRule *config.HostRule) (string, error) {
	// Get a token
	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf(authUrl, dependencyName), nil)
	if err != nil {
		return "", err
	}
	if hostRule != nil && len(hostRule.Username) > 0 && len(hostRule.Password) > 0 {
		// Add basic authentication for eg. private images
		shared.HttpUtil.AddBasicAuth(req, hostRule.UsernameExpanded(), hostRule.PasswordExpanded())
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	var tokenObj struct {
		Token string `json:"token"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&tokenObj); err != nil {
		return "", err
	}
	return tokenObj.Token, nil
}

// Gets the tags according to the v2 api spec. It uses a  bearer (token) if one is given.
func (ds *DockerDatasource) getTagsWithToken(baseUrl *url.URL, dependencyName string, bearerToken string) ([]string, error) {
	// Build the initial url
	tagListUrl := baseUrl.JoinPath(dependencyName, "tags/list")
	tagListUrl.RawQuery = url.Values{
		"n": {"1000"},
	}.Encode()

	// Loop (we might have multiple pages)
	currentUrl := tagListUrl
	ds.logger.Debug(fmt.Sprintf("Fetching Docker tags from url: %s", currentUrl))
	allTags := []string{}
	for {
		// Prepare the request
		req, err := http.NewRequest(http.MethodGet, currentUrl.String(), nil)
		if err != nil {
			return nil, err
		}
		if len(bearerToken) > 0 {
			shared.HttpUtil.AddBearerToRequest(req, bearerToken)
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
		// Parse the objects
		var tagsObj struct {
			Name string   `json:"name"`
			Tags []string `json:"tags"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&tagsObj); err != nil {
			return nil, err
		}
		allTags = append(allTags, tagsObj.Tags...)
		// Check for the next page link
		if nextPageUrl, err := shared.HttpUtil.GetNextPageURL(resp); err != nil {
			return nil, err
		} else if nextPageUrl == nil {
			// No next page
			break
		} else {
			// There is a next page
			currentUrl = nextPageUrl
		}
	}

	return allTags, nil
}