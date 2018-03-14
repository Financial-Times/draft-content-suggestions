package main

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/Financial-Times/draft-content-suggestions/draft"
	"github.com/Financial-Times/draft-content-suggestions/mocks"
	"github.com/Financial-Times/draft-content-suggestions/suggestions"
	"github.com/Financial-Times/go-ft-http/fthttp"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
)

func TestRequestHandlerSuccess(t *testing.T) {

	draftContentTestServer := mocks.NewDraftContentTestServer(true, 0)
	umbrellaTestServer := mocks.NewUmbrellaTestServer(true, 0)

	defer draftContentTestServer.Server.Close()
	defer umbrellaTestServer.Server.Close()

	contentAPI, _ := draft.NewContentAPI(draftContentTestServer.Server.URL+"/drafts/content", draftContentTestServer.Server.URL+"/__gtg", http.DefaultClient)
	umbrellaAPI, _ := suggestions.NewUmbrellaAPI(umbrellaTestServer.Server.URL, "12345", http.DefaultClient)

	requestHandler := requestHandler{contentAPI, umbrellaAPI, 8 * time.Second}

	r := mux.NewRouter()
	r.HandleFunc("/drafts/content/{uuid}/suggestions", requestHandler.draftContentSuggestionsRequest)
	ts := httptest.NewServer(r)

	defer ts.Close()

	resp, err := http.Get(ts.URL + "/drafts/content/" + mocks.ValidMockContentUUID + "/suggestions")

	defer resp.Body.Close()

	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestRequestHandlerDraftContentSLATimeout(t *testing.T) {
	draftContentTestServer := mocks.NewDraftContentTestServer(true, 200*time.Millisecond)
	umbrellaTestServer := mocks.NewUmbrellaTestServer(true, 0)

	draftContentTestServer.On("EndpointCalled")

	defer draftContentTestServer.Server.Close()
	defer umbrellaTestServer.Server.Close()

	contentAPI, _ := draft.NewContentAPI(draftContentTestServer.Server.URL+"/drafts/content", draftContentTestServer.Server.URL+"/__gtg", fthttp.NewClientWithDefaultTimeout("", ""))
	umbrellaAPI, _ := suggestions.NewUmbrellaAPI(umbrellaTestServer.Server.URL+"/content/suggest", "12345", http.DefaultClient)

	requestHandler := requestHandler{contentAPI, umbrellaAPI, 100 * time.Millisecond}

	r := mux.NewRouter()
	r.HandleFunc("/drafts/content/{uuid}/suggestions", requestHandler.draftContentSuggestionsRequest)
	ts := httptest.NewServer(r)

	defer ts.Close()

	resp, err := http.Get(ts.URL + "/drafts/content/" + mocks.ValidMockContentUUID + "/suggestions")

	defer resp.Body.Close()

	assert.NoError(t, err)
	assert.Equal(t, http.StatusGatewayTimeout, resp.StatusCode)
	assert.True(t, umbrellaTestServer.AssertNotCalled(t, "EndpointCalled"))
	draftContentTestServer.AssertExpectations(t)

}
func TestRequestHandlerUmbrellaApiSLATimeout(t *testing.T) {
	draftContentTestServer := mocks.NewDraftContentTestServer(true, 0*time.Millisecond)
	umbrellaTestServer := mocks.NewUmbrellaTestServer(true, 1000*time.Millisecond)

	draftContentTestServer.On("EndpointCalled")
	umbrellaTestServer.On("EndpointCalled")

	defer draftContentTestServer.Server.Close()
	defer umbrellaTestServer.Server.Close()

	contentAPI, _ := draft.NewContentAPI(draftContentTestServer.Server.URL+"/drafts/content", draftContentTestServer.Server.URL+"/__gtg", fthttp.NewClientWithDefaultTimeout("", ""))
	umbrellaAPI, _ := suggestions.NewUmbrellaAPI(umbrellaTestServer.Server.URL+"/content/suggest", "12345", fthttp.NewClientWithDefaultTimeout("", ""))

	requestHandler := requestHandler{contentAPI, umbrellaAPI, 100 * time.Millisecond}

	r := mux.NewRouter()
	r.HandleFunc("/drafts/content/{uuid}/suggestions", requestHandler.draftContentSuggestionsRequest)
	ts := httptest.NewServer(r)

	defer ts.Close()

	resp, err := http.Get(ts.URL + "/drafts/content/" + mocks.ValidMockContentUUID + "/suggestions")

	defer resp.Body.Close()

	assert.NoError(t, err)
	assert.Equal(t, http.StatusGatewayTimeout, resp.StatusCode)
	assert.True(t, draftContentTestServer.AssertNotCalled(t, "EndpointCalled"))
	umbrellaTestServer.AssertExpectations(t)
}
