package main

import (
	"bytes"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	logger "github.com/Financial-Times/go-logger/v2"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/Financial-Times/draft-content-suggestions/config"
	"github.com/Financial-Times/draft-content-suggestions/draft"
	"github.com/Financial-Times/draft-content-suggestions/mocks"
	"github.com/Financial-Times/draft-content-suggestions/suggestions"
)

func TestGetDraftSuggestionsForContent(t *testing.T) {
	tests := []struct {
		name                       string
		UUID                       string
		retMockSuggestionsResponse []byte
		retMockSuggestionsErr      error
		retMockContentAPIResponse  []byte
		retMockContentAPIError     error
		expectedStatus             int
		expectedError              error
		payload                    []byte
		expectedContentResult      []byte
	}{
		{
			name:                       "Successful fetch",
			UUID:                       "36320eb6-5617-4d12-9750-1907690e74db",
			expectedStatus:             http.StatusOK,
			payload:                    []byte(`test`),
			retMockContentAPIResponse:  []byte(`test`),
			retMockSuggestionsResponse: []byte(`test`),
			expectedContentResult:      []byte(`test`),
		},
		{
			name:           "Invalid uuid",
			UUID:           "invalid uuid",
			expectedStatus: http.StatusBadRequest,
			expectedContentResult: []byte(`{"message":"Invalid UUID"}
`),
		},
		{
			name:           "Empty payload",
			UUID:           "36320eb6-5617-4d12-9750-1907690e74db",
			expectedStatus: http.StatusBadRequest,
			expectedContentResult: []byte(`{"message":"content body is missing from the request"}
`),
		},
		{
			name:                   "FetchValidatedContent error case",
			UUID:                   "36320eb6-5617-4d12-9750-1907690e74db",
			expectedStatus:         http.StatusBadRequest,
			payload:                []byte(`test`),
			retMockContentAPIError: errors.New("simulated error"),
			expectedContentResult: []byte(`{"message":"failed while validating content: simulated error"}
`),
		},
		{
			name:                      "FetchSuggestions error case",
			UUID:                      "36320eb6-5617-4d12-9750-1907690e74db",
			expectedStatus:            http.StatusServiceUnavailable,
			payload:                   []byte(`test`),
			retMockContentAPIResponse: []byte(`test`),
			retMockSuggestionsErr:     errors.New("simulated error"),
			expectedContentResult: []byte(`{"message":"Suggestions umbrella api access has failed"}
`),
		},
	}

	retMockSuggestions := &suggestions.MockSuggestionsUmbrellaAPI{}
	retMockContentAPI := &draft.MockDraftContentAPI{}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			log := logger.NewUnstructuredLogger()

			rh := requestHandler{retMockContentAPI, retMockSuggestions, log}

			r := mux.NewRouter()
			r.HandleFunc("/drafts/content/{uuid}/suggestions", rh.getDraftSuggestionsForContent)
			ts := httptest.NewServer(r)

			defer ts.Close()
			req, err := http.NewRequest(http.MethodPost, ts.URL+strings.Replace("/drafts/content/{uuid}/suggestions", "{uuid}", test.UUID, 1), bytes.NewReader(test.payload))
			if err != nil {
				t.Fatal(err)
			}

			retMockContentAPI.On("FetchValidatedContent", mock.Anything, bytes.NewReader(test.payload), test.UUID, "", log).Return(test.retMockContentAPIResponse, test.retMockContentAPIError).Once()
			defer retMockContentAPI.On("FetchValidatedContent", mock.Anything, bytes.NewReader(test.payload), test.UUID, "", log).Unset()
			retMockSuggestions.On("FetchSuggestions", mock.Anything, test.payload).Return(test.retMockSuggestionsResponse, test.retMockSuggestionsErr).Once()
			defer retMockSuggestions.On("FetchSuggestions", mock.Anything, test.payload).Unset()

			resp, err := http.DefaultClient.Do(req)
			if err != nil {
				if test.expectedError == nil {
					t.Fatalf("unexpected error occurred: %v", err)
				}

				if err.Error() != test.expectedError.Error() {
					t.Fatalf("expected error: %v, got: %v", test.expectedError, err)
				}

				return
			}
			defer resp.Body.Close()

			if test.expectedError != nil {
				t.Fatalf("expected error did not occur: %v", test.expectedError)
			}

			if resp.StatusCode != test.expectedStatus {
				t.Fatalf("expected status code: %v, but got: %v", test.expectedStatus, resp.StatusCode)
			}

			respBody, err := io.ReadAll(resp.Body)
			if err != nil {
				t.Fatal(err)
			}

			if string(respBody) != string(test.expectedContentResult) {
				t.Errorf("expected result: %s, but got: %s", test.expectedContentResult, string(respBody))
			}
		})
	}
}

func TestRequestHandlerSuccess(t *testing.T) {
	resp, err := handleTestRequest("/drafts/content/" + mocks.ValidMockContentUUID + "/suggestions")

	resp.Body.Close()

	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestRequestHandlerContentNotFound(t *testing.T) {
	resp, err := handleTestRequest("/drafts/content/" + mocks.MissingMockContentUUID + "/suggestions")

	resp.Body.Close()

	assert.NoError(t, err)
	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
}

func TestRequestHandlerContentNotMappable(t *testing.T) {

	resp, err := handleTestRequest("/drafts/content/" + mocks.UnprocessableContentUUID + "/suggestions")

	resp.Body.Close()

	assert.NoError(t, err)
	assert.Equal(t, http.StatusUnprocessableEntity, resp.StatusCode)
}

func TestRequestHandlerContentInvalidUUID(t *testing.T) {

	resp, err := handleTestRequest("/drafts/content/invaliduuid/suggestions")

	resp.Body.Close()

	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func handleTestRequest(urlpath string) (resp *http.Response, err error) {
	draftContentTestServer := mocks.NewDraftContentTestServer(true)
	umbrellaTestServer := mocks.NewUmbrellaTestServer(true)

	defer draftContentTestServer.Close()
	defer umbrellaTestServer.Close()

	log := logger.NewUPPLogger("Test", "PANIC")
	validatorConfig, err := config.ReadConfig("config.local.yml")
	if err != nil {
		log.WithError(err).Fatal("unable to read r/w YAML configuration")
	}

	contentTypeMapping := draft.BuildContentTypeMapping(validatorConfig, http.DefaultClient, log)
	resolver := draft.NewDraftContentValidatorResolver(contentTypeMapping)
	contentAPI, _ := draft.NewContentAPI(draftContentTestServer.URL+"/drafts/content", draftContentTestServer.URL+"/__gtg", http.DefaultClient, http.DefaultClient, resolver)
	umbrellaAPI, _ := suggestions.NewUmbrellaAPI(umbrellaTestServer.URL, umbrellaTestServer.URL+"/__gtg", suggestions.TestUsername, suggestions.TestPassword, http.DefaultClient, http.DefaultClient)

	rh := requestHandler{contentAPI, umbrellaAPI, log}

	r := mux.NewRouter()
	r.HandleFunc("/drafts/content/{uuid}/suggestions", rh.draftContentSuggestionsRequest)
	ts := httptest.NewServer(r)

	defer ts.Close()

	return http.Get(ts.URL + urlpath)
}
