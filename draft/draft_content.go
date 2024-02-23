package draft

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/Financial-Times/draft-content-suggestions/config"
	"github.com/Financial-Times/draft-content-suggestions/endpointessentials"
	"github.com/Financial-Times/go-logger/v2"
	tidutils "github.com/Financial-Times/transactionid-utils-go"
)

var (
	ErrDraftNotMappable             = errors.New("draft content is invalid for mapping")
	ErrDraftNotValid                = errors.New("draft content is invalid")
	ErrDraftContentTypeNotSupported = errors.New("draft content-type is invalid")
)

func NewContentAPI(endpoint string, healthEndpoint string, httpClient *http.Client, healthHTTPClient *http.Client, resolver ContentValidatorResolver) (contentAPI ContentAPI, err error) {
	if !strings.HasSuffix(endpoint, "/") {
		endpoint += "/"
	}

	contentAPI = &draftContentAPI{
		endpoint,
		healthEndpoint,
		httpClient,
		healthHTTPClient,
		resolver,
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
	FetchValidatedContent(ctx context.Context, body io.Reader, contentUUID string, contentType string, log *logger.UPPLogger) ([]byte, error)
	endpointessentials.Endpoint
}

type draftContentAPI struct {
	endpoint         string
	healthEndpoint   string
	httpClient       *http.Client
	healthHTTPClient *http.Client
	resolver         ContentValidatorResolver
}

func (d *draftContentAPI) FetchDraftContent(ctx context.Context, uuid string) ([]byte, error) {
	requestPath := d.endpoint + uuid
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, requestPath, nil)
	if err != nil {
		return nil, err
	}

	response, err := d.httpClient.Do(request)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	if response.StatusCode == http.StatusNotFound {
		return nil, nil
	}
	if response.StatusCode == http.StatusUnprocessableEntity {
		return nil, ErrDraftNotMappable
	}
	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("error in draft content retrival status=%v", response.StatusCode)
	}

	bytes, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}

	return bytes, nil
}

func (d *draftContentAPI) FetchValidatedContent(ctx context.Context, body io.Reader, contentUUID string, contentType string, log *logger.UPPLogger) ([]byte, error) {
	tid, _ := tidutils.GetTransactionIDFromContext(ctx)
	readLog := log.WithField(tidutils.TransactionIDHeader, tid).WithField("uuid", contentUUID)

	var validatedContent io.ReadCloser

	validator, resolverErr := d.resolver.ValidatorForContentType(contentType)

	if resolverErr != nil {
		readLog.WithError(resolverErr).Error("Unable to validate content")
		return nil, resolverErr
	}

	validatedContent, err := validator.Validate(ctx, contentUUID, body, contentType, log)
	if err != nil {
		readLog.WithError(err).Warn("Validator error")
		var validatorError ValidatorError
		if errors.As(err, &validatorError) {
			switch validatorError.StatusCode() {
			case http.StatusNotFound:
				fallthrough
			case http.StatusUnsupportedMediaType:
				err = ErrDraftContentTypeNotSupported
			case http.StatusUnprocessableEntity:
				err = ErrDraftNotValid
			}
		}
		return nil, err
	}
	defer validatedContent.Close()

	bytes, err := io.ReadAll(validatedContent)
	if err != nil {
		return nil, err
	}

	return bytes, err
}

func BuildContentTypeMapping(validatorConfig *config.Config, httpClient *http.Client, log *logger.UPPLogger) map[string]ContentValidator {
	contentTypeMapping := map[string]ContentValidator{}

	for contentType, cfg := range validatorConfig.ContentTypes {
		var service ContentValidator

		switch cfg.Validator {
		case "generic":
			service = NewDraftContentValidatorService(cfg.Endpoint, httpClient)
		default:
			log.WithField("Validator", cfg.Validator).Fatal("Unknown validator")
		}
		contentTypeMapping[contentType] = service

		log.
			WithField("Content-Type", contentType).
			WithField("Endpoint", cfg.Endpoint).
			WithField("Validator", cfg.Validator).
			Info("added validator service")
	}

	return contentTypeMapping
}

func (d *draftContentAPI) Endpoint() string {
	return d.endpoint
}

func (d *draftContentAPI) IsGTG(ctx context.Context) (string, error) {
	req, err := http.NewRequest(http.MethodGet, d.healthEndpoint, nil)
	if err != nil {
		return "", fmt.Errorf("error in creating GTG request: %w", err)
	}

	response, err := d.healthHTTPClient.Do(req.WithContext(ctx))
	if err != nil {
		return "", fmt.Errorf("error in GTG request: %w", err)
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return "", fmt.Errorf("non-200 HTTP response (%v) on GTG request", response.StatusCode)
	}

	return "draft-content-public-read is healthy", nil
}

func (d *draftContentAPI) IsValid() error {
	return endpointessentials.ValidateEndpoint(d.endpoint)
}
