package draft

import (
	"bytes"
	"context"
	"errors"
	"io"
	"net/http"
	"net/url"
	"testing"

	"github.com/Financial-Times/draft-content-suggestions/mocks"
	"github.com/Financial-Times/go-logger/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

const (
	//originIDcctTest       = "cct"
	contentTypeArticle    = "application/vnd.ft-upp-article+json"
	testBasicAuthUsername = "testUsername"
	testBasicAuthPassword = "testPassword"
	testTID               = "test_tid"
)

func TestDraftContentAPI_IsGTGSuccess(t *testing.T) {
	testServer := mocks.NewDraftContentTestServer(true)
	defer testServer.Close()

	contentAPI, err := NewContentAPI(testServer.URL+"/drafts/content", testServer.URL+"/__gtg", http.DefaultClient, http.DefaultClient, &draftContentValidatorResolver{})

	assert.NoError(t, err)

	msg, err := contentAPI.IsGTG(context.Background())
	assert.NoError(t, err)
	assert.Equal(t, "draft-content-public-read is healthy", msg)
}

func TestDraftContentAPI_IsGTGFailure503(t *testing.T) {
	testServer := mocks.NewDraftContentTestServer(false)
	defer testServer.Close()

	contentAPI, err := NewContentAPI(testServer.URL+"/drafts/content", testServer.URL+"/__gtg", http.DefaultClient, http.DefaultClient, &draftContentValidatorResolver{})

	assert.NoError(t, err)

	_, err = contentAPI.IsGTG(context.Background())
	assert.Error(t, err)
}

func TestDraftContentAPI_IsGTGFailureInvalidEndpoint(t *testing.T) {
	testServer := mocks.NewDraftContentTestServer(false)
	defer testServer.Close()

	contentAPI, err := NewContentAPI(testServer.URL+"/drafts/content", ":#", http.DefaultClient, http.DefaultClient, &draftContentValidatorResolver{})

	assert.NoError(t, err)

	_, err = contentAPI.IsGTG(context.Background())

	var urlErr *url.Error
	if assert.Error(t, err) && errors.As(err, &urlErr) {
		assert.Equal(t, "parse", urlErr.Op)
	}
}

func TestDraftContentAPI_IsGTGFailureRequestError(t *testing.T) {
	testServer := mocks.NewDraftContentTestServer(false)
	defer testServer.Close()

	contentAPI, err := NewContentAPI(testServer.URL+"/drafts/content", "__gtg", http.DefaultClient, http.DefaultClient, &draftContentValidatorResolver{})

	assert.NoError(t, err)

	_, err = contentAPI.IsGTG(context.Background())
	var urlErr *url.Error
	if assert.Error(t, err) && errors.As(err, &urlErr) {
		assert.Equal(t, "Get", urlErr.Op)
	}
}

func TestDraftContentAPI_FetchDraftContentSuccess(t *testing.T) {
	testServer := mocks.NewDraftContentTestServer(true)
	defer testServer.Close()

	contentAPI, err := NewContentAPI(testServer.URL+"/drafts/content", testServer.URL+"/__gtg", http.DefaultClient, http.DefaultClient, &draftContentValidatorResolver{})
	assert.NoError(t, err)

	content, err := contentAPI.FetchDraftContent(context.Background(), mocks.ValidMockContentUUID)

	assert.NoError(t, err)
	assert.True(t, content != nil)
	assert.True(t, len(content) > 0)
}

func TestDraftContentAPI_FetchValidatedContent(t *testing.T) {
	tests := []struct {
		name                     string
		retMockResolverValidator *MockValidator
		retMockResolverErr       error
		retMockValidatorBody     io.Reader
		retMockValidatorErr      error
		expectedValidatorErr     error
		expectedStatusCode       int
		expectedContentResult    []byte
	}{
		{
			name:                     "Successful fetch",
			retMockResolverValidator: &MockValidator{},
			retMockValidatorBody:     bytes.NewReader([]byte(`validated content`)),
			expectedContentResult:    []byte(`validated content`),
		},
		{
			name:                     "Unsuccessful fetch caused by missing validator",
			retMockResolverValidator: nil,
			retMockResolverErr:       ValidatorError{httpStatus: http.StatusBadRequest, msg: "Validator error"},
		},
		{
			name:                     "Unsuccessful fetch caused by validator error",
			retMockResolverValidator: &MockValidator{},
			retMockValidatorBody:     bytes.NewReader([]byte{}),
			retMockValidatorErr:      ValidatorError{httpStatus: http.StatusBadRequest, msg: "Validator error"},
			expectedValidatorErr:     ValidatorError{httpStatus: http.StatusBadRequest, msg: "Validator error"},
		},
		{
			name:                     "Unsuccessful fetch caused by validator error with unsupported media type",
			retMockResolverValidator: &MockValidator{},
			retMockValidatorBody:     bytes.NewReader([]byte{}),
			retMockValidatorErr:      ValidatorError{httpStatus: http.StatusUnsupportedMediaType, msg: "Validator error"},
			expectedValidatorErr:     ErrDraftContentTypeNotSupported,
		},
		{
			name:                     "Unsuccessful fetch caused by validator error with unprocessed entity",
			retMockResolverValidator: &MockValidator{},
			retMockValidatorBody:     bytes.NewReader([]byte{}),
			retMockValidatorErr:      ValidatorError{httpStatus: http.StatusUnprocessableEntity, msg: "Validator error"},
			expectedValidatorErr:     ErrDraftNotValid,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			testServer := mocks.NewDraftContentTestServer(true)
			defer testServer.Close()

			log := logger.NewUnstructuredLogger()

			mockResolver := &MockValidatorResolver{}
			if test.retMockResolverValidator != nil {
				test.retMockResolverValidator.On("Validate", mock.Anything, "", bytes.NewReader([]byte{}), "", log).Return(test.retMockValidatorBody, test.retMockValidatorErr).Once()
			}

			mockResolver.On("ValidatorForContentType", "").Return(test.retMockResolverValidator, test.retMockResolverErr).Once()
			contentAPI, err := NewContentAPI(testServer.URL+"/drafts/content", testServer.URL+"/__gtg", http.DefaultClient, http.DefaultClient, mockResolver)
			assert.NoError(t, err)
			content, err := contentAPI.FetchValidatedContent(context.Background(), bytes.NewReader([]byte{}), "", "", log)

			if test.retMockResolverErr != nil {
				assert.Equal(t, test.retMockResolverErr, err)
			}
			if test.retMockValidatorErr != nil {
				assert.Equal(t, test.expectedValidatorErr, err)
			}
			assert.Equal(t, test.expectedContentResult, content)
		})
	}
}

func TestDraftContentAPI_FetchDraftContentMissing(t *testing.T) {

	testServer := mocks.NewDraftContentTestServer(true)
	defer testServer.Close()

	contentAPI, err := NewContentAPI(testServer.URL+"/drafts/content", testServer.URL+"/__gtg", http.DefaultClient, http.DefaultClient, &draftContentValidatorResolver{})
	assert.NoError(t, err)

	content, err := contentAPI.FetchDraftContent(context.Background(), mocks.MissingMockContentUUID)

	assert.NoError(t, err)
	assert.True(t, content == nil)
}

func TestDraftContentAPI_FetchDraftContentFailure(t *testing.T) {

	contentAPI, err := NewContentAPI("http://localhost/drafts/content", "http://localhost/__gtg", http.DefaultClient, http.DefaultClient, &draftContentValidatorResolver{})
	assert.NoError(t, err)

	content, err := contentAPI.FetchDraftContent(context.Background(), mocks.ValidMockContentUUID)

	assert.Error(t, err)
	assert.True(t, content == nil)
}

func TestDraftContentAPI_FetchDraftContentUnmappable(t *testing.T) {

	testServer := mocks.NewDraftContentTestServer(false)
	defer testServer.Close()

	contentAPI, err := NewContentAPI(testServer.URL+"/drafts/content", testServer.URL+"/__gtg", http.DefaultClient, http.DefaultClient, &draftContentValidatorResolver{})
	assert.NoError(t, err)

	content, err := contentAPI.FetchDraftContent(context.Background(), mocks.UnprocessableContentUUID)

	assert.Error(t, err)
	assert.True(t, errors.Is(err, ErrDraftNotMappable))
	assert.True(t, content == nil)

}

func TestDraftContentAPI_FetchDraftContentNon200(t *testing.T) {

	testServer := mocks.NewDraftContentTestServer(true)
	defer testServer.Close()

	contentAPI, err := NewContentAPI(testServer.URL+"/drafts/content", testServer.URL+"/__gtg", http.DefaultClient, http.DefaultClient, &draftContentValidatorResolver{})
	assert.NoError(t, err)

	content, err := contentAPI.FetchDraftContent(context.Background(), mocks.FailsRetrivalContentUuid)

	assert.Error(t, err)
	assert.EqualError(t, err, "error in draft content retrival status=500")
	assert.True(t, content == nil)
}
