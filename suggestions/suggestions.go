package suggestions

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/Financial-Times/draft-content-suggestions/commons"
	"github.com/satori/go.uuid"
)

func NewUmbrellaAPI(endpoint string, apiKey string, httpClient *http.Client) (UmbrellaAPI, error) {

	umbrellaAPI := &umbrellaAPI{endpoint, apiKey, httpClient}

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
	endpoint   string
	apiKey     string
	httpClient *http.Client
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
	newUUID := uuid.NewV4()

	c := make(map[string]interface{})
	c["uuid"] = newUUID.String()
	content, err := json.Marshal(c)

	_, err = u.FetchSuggestions(ctx, content)

	if err != nil {
		return "", err
	}

	return "suggestions umbrella service is healthy", nil
}

func (u *umbrellaAPI) IsValid() error {
	return commons.ValidateEndpoint(u.endpoint)
}
