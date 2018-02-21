package suggestions

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/Financial-Times/draft-content-suggestions/commons"
	"github.com/Financial-Times/draft-content-suggestions/draft"
	"github.com/satori/go.uuid"
	"io"
	"net/http"
)

func NewUmbrellaAPI(endpoint string, httpClient *http.Client) (UmbrellaAPI, error) {

	umbrellaAPI := &umbrellaAPI{endpoint, httpClient}

	err := umbrellaAPI.IsValid()

	if err != nil {
		return nil, err
	}

	return umbrellaAPI, nil
}

type UmbrellaAPI interface {
	// FetchSuggestions
	// Makes a API request to Suggestions Umbrella and directly returns the
	// response io.ReadCloser for possible pipelined streaming.
	FetchSuggestions(ctx context.Context, content *draft.Content) (suggestion io.ReadCloser, err error)

	// Embedded Endpoint interface, check its godoc
	commons.Endpoint
}

type umbrellaAPI struct {
	endpoint   string
	httpClient *http.Client
}

func (u *umbrellaAPI) FetchSuggestions(ctx context.Context, content *draft.Content) (suggestion io.ReadCloser, err error) {
	jsonBytes, err := json.Marshal(content)

	if err != nil {
		return nil, err
	}

	request, err := commons.NewHttpRequest(ctx, http.MethodPost, u.endpoint, bytes.NewBuffer(jsonBytes))

	if err != nil {
		return nil, err
	}

	response, err := u.httpClient.Do(request)

	if err != nil {
		return nil, err
	}

	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return nil,
			errors.New(fmt.Sprintf("Suggestions Umbrella endpoint fails with response code: %v", response.StatusCode))
	}

	return response.Body, nil
}

func (u *umbrellaAPI) Endpoint() string {
	return u.endpoint
}

func (u *umbrellaAPI) IsHealthy(ctx context.Context) (string, error) {
	newUUID, err := uuid.NewV4()

	if err != nil {
		return "", errors.New(fmt.Sprintln())
	}

	c := &draft.Content{UUID: newUUID.String()}

	_, err = u.FetchSuggestions(ctx, c)

	if err != nil {
		return "", err
	}

	return "suggestions umbrella service is healthy", nil
}

func (u *umbrellaAPI) IsValid() error {
	return commons.ValidateEndpoint(u.endpoint)
}
