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

	defer draftContentTestServer.Close()
	defer umbrellaTestServer.Close()

	contentAPI, _ := draft.NewContentAPI(draftContentTestServer.URL+"/drafts/content", draftContentTestServer.URL+"/__gtg", http.DefaultClient)
	umbrellaAPI, _ := suggestions.NewUmbrellaAPI(umbrellaTestServer.URL, "12345", http.DefaultClient)

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

func TestRequestHandlerSLATimeout(t *testing.T) {
	draftContentTestServer := mocks.NewDraftContentTestServer(true, 200*time.Millisecond)
	umbrellaTestServer := mocks.NewUmbrellaTestServer(true, 200*time.Millisecond)

	defer draftContentTestServer.Close()
	defer umbrellaTestServer.Close()

	contentAPI, _ := draft.NewContentAPI(draftContentTestServer.URL+"/drafts/content", draftContentTestServer.URL+"/__gtg", fthttp.NewClientWithDefaultTimeout("", ""))
	umbrellaAPI, _ := suggestions.NewUmbrellaAPI(umbrellaTestServer.URL, "12345", http.DefaultClient)

	requestHandler := requestHandler{contentAPI, umbrellaAPI, 100 * time.Millisecond}

	r := mux.NewRouter()
	r.HandleFunc("/drafts/content/{uuid}/suggestions", requestHandler.draftContentSuggestionsRequest)
	ts := httptest.NewServer(r)

	defer ts.Close()

	resp, err := http.Get(ts.URL + "/drafts/content/" + mocks.ValidMockContentUUID + "/suggestions")

	defer resp.Body.Close()

	assert.NoError(t, err)
	assert.Equal(t, http.StatusGatewayTimeout, resp.StatusCode)

}
