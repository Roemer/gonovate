package core

import (
	"fmt"
	"io"
	"net/http"
)

const (
	ContentTypeNone   = ""
	ContentTypeBinary = "application/octet-stream"
	ContentTypeForm   = "application/x-www-form-urlencoded"
	ContentTypeJSON   = "application/json"
	ContentTypeHTML   = "text/html; charset=utf-8"
	ContentTypeText   = "text/plain; charset=utf-8"
)

type HttpUtil struct {
}

func (h HttpUtil) DownloadToMemory(url string) ([]byte, error) {
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

func (h HttpUtil) GetTextFromBody(response *http.Response) (string, error) {
	body, err := io.ReadAll(response.Body)
	if err != nil {
		return "", err
	}
	return string(body), nil
}

func (h HttpUtil) AddBearerToRequest(request *http.Request, token string) {
	if len(token) > 0 {
		request.Header.Add("Authorization", fmt.Sprintf("Bearer %s", token))
	}
}

func (h HttpUtil) AddBasicAuth(request *http.Request, username, password string) {
	if len(username) > 0 && len(password) > 0 {
		request.SetBasicAuth(username, password)
	}
}
