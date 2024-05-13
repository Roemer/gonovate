package datasources

import (
	"bytes"
	"encoding/json"
	"fmt"
	"gonovate/core"
	"log/slog"
	"net/http"
	"net/url"
	"regexp"
	"strings"

	"github.com/roemer/gover"
)

type DockerDatasource struct {
	datasourcesBase
}

func NewDockerDatasource(logger *slog.Logger) *DockerDatasource {
	newDatasource := &DockerDatasource{}
	newDatasource.logger = logger
	return newDatasource
}

func (ds *DockerDatasource) SearchPackageUpdate(packageName string, currentVersion string, packageSettings *core.PackageSettings, hostRules []*core.HostRule) (string, bool, error) {
	// Default to Docker registry
	baseUrlString := "https://index.docker.io/v2"
	if packageSettings != nil && len(packageSettings.RegistryUrls) > 0 {
		baseUrlString = packageSettings.RegistryUrls[0]
		ds.logger.Debug(fmt.Sprintf("Using custom registry url: %s", baseUrlString))
	}
	baseUrl, err := url.Parse(baseUrlString)
	if err != nil {
		return "", false, err
	}

	// Get a host rule if any was defined
	relevantHostRule := core.FilterHostConfigsForHost(baseUrl.Host, hostRules)

	// Different handling for different sites
	var tags []string
	if strings.Contains(baseUrl.Host, "hub.docker.com") {
		tags, err = ds.getTagsForDockerHub(baseUrl, packageName, relevantHostRule)
		if err != nil {
			return "", false, err
		}
	} else if strings.Contains(baseUrl.Host, "index.docker.io") {
		tags, err = ds.getTagsForDocker(baseUrl, packageName, relevantHostRule)
		if err != nil {
			return "", false, err
		}
	} else if strings.Contains(baseUrl.Host, "ghcr.io") {
		tags, err = ds.getTagsForGhcr(baseUrl, packageName, relevantHostRule)
		if err != nil {
			return "", false, err
		}
	} else if strings.Contains(baseUrl.Host, "gcr.io") {
		tags, err = ds.getTagsForGcr(baseUrl, packageName, relevantHostRule)
		if err != nil {
			return "", false, err
		}
	} else if strings.Contains(baseUrl.Host, "quay.io") {
		// For quay we need a special token
		tags, err = ds.getTagsForQuay(baseUrl, packageName, relevantHostRule)
		if err != nil {
			return "", false, err
		}
	} else {
		// For everything else we just use a bearer token (if provided), eg. Artifactory
		bearerToken := ""
		if relevantHostRule != nil {
			bearerToken = relevantHostRule.TokendExpanded()
		}
		tags, err = ds.getTagsWithToken(baseUrl, packageName, bearerToken)
		if err != nil {
			return "", false, err
		}
	}

	// Search for an updated version

	// Convert all entries to objects
	ignoreNoneMatching := false
	if packageSettings.IgnoreNonMatching != nil {
		ignoreNoneMatching = *packageSettings.IgnoreNonMatching
	}
	// TODO: Define default versioning
	versionRegex := regexp.MustCompile(packageSettings.Versioning)
	versions := []*gover.Version{}
	for _, tag := range tags {
		version, err := gover.ParseVersionFromRegex(tag, versionRegex)
		if err != nil {
			if ignoreNoneMatching {
				continue
			}
			return "", false, err
		}
		versions = append(versions, version)
	}

	curr, err := gover.ParseVersionFromRegex(currentVersion, versionRegex)
	if err != nil {
		return "", false, err
	}
	refVersion := ds.getReferenceVersionForUpdateType(packageSettings.MaxUpdateType, curr)
	// Search for an update
	maxValidVersion := gover.FindMax(versions, refVersion, false)

	if maxValidVersion.Equals(curr) {
		ds.logger.Debug("Found no new version")
		return "", false, nil
	}

	ds.logger.Info(fmt.Sprintf("Found a new version: %s", maxValidVersion.Original))

	return maxValidVersion.Original, true, nil
}

// Docker Hub does not implement the Docker Registry API and has a different way...
func (ds *DockerDatasource) getTagsForDockerHub(baseUrl *url.URL, packageName string, hostRule *core.HostRule) ([]string, error) {
	// Get an authentication token
	loginUrl := baseUrl.JoinPath("users/login")
	payloadBuf := new(bytes.Buffer)
	if err := json.NewEncoder(payloadBuf).Encode(struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}{
		Username: hostRule.UsernameExpanded(),
		Password: hostRule.PasswordExpanded(),
	}); err != nil {
		return nil, err
	}
	resp, err := http.DefaultClient.Post(loginUrl.String(), core.ContentTypeJSON, payloadBuf)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	var tokenObj struct {
		Token string `json:"token"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&tokenObj); err != nil {
		return nil, err
	}

	// Get the tags with the token
	tagListUrl := baseUrl.JoinPath("repositories", packageName, "tags")
	tagListUrl.RawQuery = url.Values{
		"page_size": {"1000"},
	}.Encode()
	req, err := http.NewRequest(http.MethodGet, tagListUrl.String(), nil)
	if err != nil {
		return nil, err
	}
	core.HttpUtil.AddBearerToRequest(req, tokenObj.Token)
	resp, err = http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("failed with statuscode %d", resp.StatusCode)
	}

	var tagsObj struct {
		Count   int    `json:"count"`
		Next    string `json:"next"`
		Results []*struct {
			Name string `json:"name"`
		} `json:"results"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&tagsObj); err != nil {
		return nil, err
	}

	// TODO: Loop while Next contains an url

	// Build the tag list to return
	tags := []string{}
	for _, image := range tagsObj.Results {
		tags = append(tags, image.Name)
	}

	return tags, nil
}

func (ds *DockerDatasource) getTagsForDocker(baseUrl *url.URL, packageName string, hostRule *core.HostRule) ([]string, error) {
	token, err := ds.getJwtToken("https://auth.docker.io/token?service=registry.docker.io&scope=repository:%s:pull", packageName, hostRule)
	if err != nil {
		return nil, err
	}
	return ds.getTagsWithToken(baseUrl, packageName, token)
}

func (ds *DockerDatasource) getTagsForGhcr(baseUrl *url.URL, packageName string, hostRule *core.HostRule) ([]string, error) {
	token, err := ds.getJwtToken("https://ghcr.io/token?service=ghcr.io&scope=repository:%s:pull", packageName, hostRule)
	if err != nil {
		return nil, err
	}
	return ds.getTagsWithToken(baseUrl, packageName, token)
}

func (ds *DockerDatasource) getTagsForGcr(baseUrl *url.URL, packageName string, hostRule *core.HostRule) ([]string, error) {
	token, err := ds.getJwtToken("https://gcr.io/v2/token?service=gcr.io&scope=repository:%s:pull", packageName, hostRule)
	if err != nil {
		return nil, err
	}
	return ds.getTagsWithToken(baseUrl, packageName, token)
}

func (ds *DockerDatasource) getTagsForQuay(baseUrl *url.URL, packageName string, hostRule *core.HostRule) ([]string, error) {
	token, err := ds.getJwtToken("https://quay.io/v2/auth?service=quay.io&scope=repository:%s:pull", packageName, hostRule)
	if err != nil {
		return nil, err
	}
	return ds.getTagsWithToken(baseUrl, packageName, token)
}

// Creates a request to get a token and returns the token. Uses basic auth uf username/password is provided.
func (ds *DockerDatasource) getJwtToken(authUrl string, packageName string, hostRule *core.HostRule) (string, error) {
	// Get a token
	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf(authUrl, packageName), nil)
	if err != nil {
		return "", err
	}
	if hostRule != nil && len(hostRule.Username) > 0 && len(hostRule.Password) > 0 {
		// Add basic authentication for eg. private images
		core.HttpUtil.AddBasicAuth(req, hostRule.Username, hostRule.Password)
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
func (ds *DockerDatasource) getTagsWithToken(baseUrl *url.URL, packageName string, bearerToken string) ([]string, error) {
	// Build the initial url
	tagListUrl := baseUrl.JoinPath(packageName, "tags/list")
	tagListUrl.RawQuery = url.Values{
		"n": {"1000"},
	}.Encode()

	// Loop (we might have multiple pages)
	currentUrl := tagListUrl
	allTags := []string{}
	for {
		// Prepare the request
		req, err := http.NewRequest(http.MethodGet, currentUrl.String(), nil)
		if err != nil {
			return nil, err
		}
		if len(bearerToken) > 0 {
			core.HttpUtil.AddBearerToRequest(req, bearerToken)
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
		if nextPageUrl, err := core.HttpUtil.GetNextPageURL(resp); err != nil {
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
