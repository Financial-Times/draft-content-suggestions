package main

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Financial-Times/draft-content-suggestions/draft"
	"github.com/Financial-Times/draft-content-suggestions/mocks"
	"github.com/Financial-Times/draft-content-suggestions/suggestions"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"time"
)

func TestRequestHandlerSuccess(t *testing.T) {

	draftContentTestServer := mocks.NewDraftContentTestServer(true)
	umbrellaTestServer := mocks.NewUmbrellaTestServer(true)

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

}
