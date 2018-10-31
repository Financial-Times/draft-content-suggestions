package suggestions

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/Financial-Times/draft-content-suggestions/commons"
	log "github.com/sirupsen/logrus"
)

func NewUmbrellaAPI(endpoint string, gtgEndpoint string, apiKey string, httpClient *http.Client) (UmbrellaAPI, error) {

	umbrellaAPI := &umbrellaAPI{endpoint, gtgEndpoint, apiKey, httpClient}

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
	commons.Endpoint
}

type umbrellaAPI struct {
	endpoint    string
	gtgEndpoint string
	apiKey      string
	httpClient  *http.Client
}

func (u *umbrellaAPI) FetchSuggestions(ctx context.Context, content []byte) (suggestion []byte, err error) {

	request, err := http.NewRequest(http.MethodPost, u.endpoint, bytes.NewBuffer(content))
	request.Header.Set("X-Api-Key", u.apiKey)

	if err != nil {
		return nil, err
	}

	response, err := u.httpClient.Do(request.WithContext(ctx))

	if err != nil {
		return nil, err
	}

	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return nil,
			errors.New(fmt.Sprintf("Suggestions Umbrella endpoint fails with response code: %v", response.StatusCode))
	}

	suggestion, err = ioutil.ReadAll(response.Body)

	if err != nil {
		return nil,
			errors.New(fmt.Sprintf("Failed reading the response body from Suggestions Umbrella endpoint %v", err))
	}

	return suggestion, nil
}

func (u *umbrellaAPI) Endpoint() string {
	return u.endpoint
}

func (u *umbrellaAPI) IsGTG(ctx context.Context) (string, error) {

	gtgReq, err := http.NewRequest(http.MethodGet, u.gtgEndpoint, nil)

	if err != nil {
		log.WithError(err).WithField("healthEndpoint", u.gtgEndpoint).Error("Error in creating GTG request to UPP suggestions API")
		return "", err
	}
	gtgReq.Header.Set("X-Api-Key", u.apiKey)

	response, err := u.httpClient.Do(gtgReq.WithContext(ctx))

	if err != nil {
		log.WithError(err).WithField("healthEndpoint", u.gtgEndpoint).Error("Error in GTG request to UPP suggestions API")
		return "", err
	}

	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		log.WithField("healthEndpoint", u.gtgEndpoint).
			WithField("status", response.StatusCode).
			Error("GTG for UPP suggestions API returned a non-200 HTTP status")
		return "", fmt.Errorf("GTG for UPP suggestions API returned a non-200 HTTP status: %v", response.StatusCode)
	}

	return "UPP suggestions API is healthy", nil
}

func (u *umbrellaAPI) IsValid() error {
	return commons.ValidateEndpoint(u.endpoint)
}
