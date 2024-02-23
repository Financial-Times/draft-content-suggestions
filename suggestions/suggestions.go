package suggestions

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"

	"github.com/Financial-Times/draft-content-suggestions/endpointessentials"
)

const (
	TestUsername = "username"
	TestPassword = "password"
	OriginHeader = "X-Origin"
	Origin       = "PAC"
)

func NewUmbrellaAPI(endpoint string, gtgEndpoint string, username string, password string, httpClient *http.Client, healthHTTPClient *http.Client) (UmbrellaAPI, error) {
	umbrellaAPI := &umbrellaAPI{endpoint, gtgEndpoint, username, password, httpClient, healthHTTPClient}

	err := umbrellaAPI.IsValid()
	if err != nil {
		return nil, err
	}

	return umbrellaAPI, nil
}

type UmbrellaAPI interface {
	// FetchSuggestions
	// Makes a API request to Suggestions Umbrella and returns the
	// []byte body
	FetchSuggestions(ctx context.Context, content []byte) (suggestion []byte, err error)

	// Embedded Endpoint interface, check its godoc
	endpointessentials.Endpoint
}

type umbrellaAPI struct {
	endpoint         string
	gtgEndpoint      string
	username         string
	password         string
	httpClient       *http.Client
	healthHTTPClient *http.Client
}

func (u *umbrellaAPI) FetchSuggestions(ctx context.Context, content []byte) (suggestion []byte, err error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, u.endpoint, bytes.NewBuffer(content))
	if err != nil {
		return nil, err
	}

	req.SetBasicAuth(u.username, u.password)
	req.Header.Set(OriginHeader, Origin)

	res, err := u.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("suggestions Umbrella endpoint fail: %s", res.Status)
	}

	suggestion, err = io.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("failed reading the response body from Suggestions Umbrella endpoint: %w", err)
	}

	return suggestion, nil
}

func (u *umbrellaAPI) Endpoint() string {
	return u.endpoint
}

func (u *umbrellaAPI) IsGTG(ctx context.Context) (string, error) {
	req, err := http.NewRequest(http.MethodGet, u.gtgEndpoint, nil)
	if err != nil {
		return "", fmt.Errorf("error creating GTG request: %w", err)
	}

	req.SetBasicAuth(u.username, u.password)

	response, err := u.healthHTTPClient.Do(req.WithContext(ctx))
	if err != nil {
		return "", fmt.Errorf("error sending GTG request: %w", err)
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return "", fmt.Errorf("received non-200 HTTP response: %d", response.StatusCode)
	}

	return "UPP suggestions API is healthy", nil
}

func (u *umbrellaAPI) IsValid() error {
	return endpointessentials.ValidateEndpoint(u.endpoint)
}
