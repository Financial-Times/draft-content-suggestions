package draft

import (
	"context"
	"errors"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/Financial-Times/draft-content-suggestions/commons"
)


var (
	ErrDraftNotMappable = errors.New("draft content is invalid for mapping status 422")
	ErrRetrievingContent=errors.New("error in draft content retrival status non-200")
	)

func NewContentAPI(endpoint string, healthEndpoint string, httpClient *http.Client) (contentAPI ContentAPI, err error) {

	if !strings.HasSuffix(endpoint, "/") {
		endpoint += "/"
	}

	contentAPI = &draftContentAPI{
		endpoint,
		healthEndpoint,
		httpClient,
	}

	err = contentAPI.IsValid()

	if err != nil {
		return nil, err
	}

	return contentAPI, nil

}

// ContentApi for accessing to draft-content-api endpoint
type ContentAPI interface {
	FetchDraftContent(ctx context.Context, uuid string) (content []byte, err error)
	commons.Endpoint
}

type draftContentAPI struct {
	endpoint       string
	healthEndpoint string
	httpClient     *http.Client
}

func (d *draftContentAPI) FetchDraftContent(ctx context.Context, uuid string) ([]byte, error) {

	requestPath := d.endpoint + uuid
	request, err := http.NewRequest(http.MethodGet, requestPath, nil)

	if err != nil {
		return nil, err
	}

	response, err := d.httpClient.Do(request.WithContext(ctx))

	if err != nil {
		return nil, err
	}

	defer response.Body.Close()

	if response.StatusCode == http.StatusNotFound {
		return nil, nil
	}

	if response.StatusCode == http.StatusUnprocessableEntity{
		return nil, ErrDraftNotMappable
	}

	if response.StatusCode!=http.StatusOK{
		return nil, ErrRetrievingContent
	}

	bytes, err := ioutil.ReadAll(response.Body)

	if err != nil {
		return nil, err
	}

	return bytes, nil
}

func (d *draftContentAPI) Endpoint() string {
	return d.endpoint
}

func (d *draftContentAPI) IsGTG(ctx context.Context) (string, error) {
	request, err := http.NewRequest(http.MethodGet, d.healthEndpoint, nil)

	if err != nil {
		return "", err
	}

	response, err := d.httpClient.Do(request.WithContext(ctx))

	if err != nil {
		return "", err
	}

	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return "", errors.New("draft-content-api endpoint is unhealthy")
	}

	return "draft-content-api is healthy", nil

}

func (d *draftContentAPI) IsValid() error {
	return commons.ValidateEndpoint(d.endpoint)
}
