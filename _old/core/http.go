package core

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"
)

const (
	ContentTypeNone   = ""
	ContentTypeBinary = "application/octet-stream"
	ContentTypeForm   = "application/x-www-form-urlencoded"
	ContentTypeJSON   = "application/json"
	ContentTypeHTML   = "text/html; charset=utf-8"
	ContentTypeText   = "text/plain; charset=utf-8"
)

// Unexported type
type httpUtil struct{}

// exported global variable
var HttpUtil httpUtil

func (h httpUtil) DownloadToMemory(url string) ([]byte, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to download file '%s'. Status code: %d", url, resp.StatusCode)
	}
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read body: %w", err)
	}
	return bodyBytes, nil
}

func (h httpUtil) GetTextFromBody(response *http.Response) (string, error) {
	body, err := io.ReadAll(response.Body)
	if err != nil {
		return "", err
	}
	return string(body), nil
}

func (h httpUtil) AddBearerToRequest(request *http.Request, token string) {
	if len(token) > 0 {
		request.Header.Add("Authorization", fmt.Sprintf("Bearer %s", token))
	}
}

func (h httpUtil) AddBasicAuth(request *http.Request, username, password string) {
	if len(username) > 0 && len(password) > 0 {
		request.SetBasicAuth(username, password)
	}
}

// Gets the "next" link from the "Link" header from a response.
func (h httpUtil) GetNextPageURL(resp *http.Response) (*url.URL, error) {
	// See if we have the link header
	linkHeaderRaw := resp.Header.Get("Link")
	if linkHeaderRaw == "" {
		return nil, nil
	}

	// Make sure we have a request url (needet to resolve the link)
	if resp.Request == nil || resp.Request.URL == nil {
		return nil, nil
	}

	// Prepare the regular expression to parse
	linkRegex, err := regexp.Compile(`\s*<(.*)>; *rel="(.*)"\s*`)
	if err != nil {
		return nil, err
	}

	// Split it in case there are multiple links inside the header
	linksRaw := strings.Split(linkHeaderRaw, ",")
	for _, linkRaw := range linksRaw {
		if matches := linkRegex.FindStringSubmatch(linkRaw); matches != nil {
			if matches[2] != "next" {
				continue
			}
			linkURL, err := url.Parse(matches[1])
			if err != nil {
				return nil, err
			}
			// Resolve and return the url
			return resp.Request.URL.ResolveReference(linkURL), nil
		}
	}

	// Nothing found, return
	return nil, nil
}
