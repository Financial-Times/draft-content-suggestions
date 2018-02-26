package draft

import (
	"context"
	"net/http"
	"testing"

	"github.com/Financial-Times/draft-content-suggestions/mocks"
	"github.com/stretchr/testify/assert"
)

func TestDraftContentAPI_IsHealthySuccess(t *testing.T) {

	testServer := mocks.NewDraftContentTestServer(true)
	defer testServer.Close()

	contentAPI, err := NewContentAPI(testServer.URL+"/drafts/content", testServer.URL+"/__health", http.DefaultClient)

	assert.NoError(t, err)

	_, err = contentAPI.IsGTG(context.Background())
	assert.NoError(t, err)
}
func TestDraftContentAPI_IsHealthyFailure(t *testing.T) {
	testServer := mocks.NewDraftContentTestServer(false)
	defer testServer.Close()

	contentAPI, err := NewContentAPI(testServer.URL+"/drafts/content", testServer.URL+"/__gtg", http.DefaultClient)

	assert.NoError(t, err)

	_, err = contentAPI.IsGTG(context.Background())
	assert.Error(t, err)
}
func TestDraftContentAPI_FetchDraftContentSuccess(t *testing.T) {

	testServer := mocks.NewDraftContentTestServer(true)
	defer testServer.Close()

	contentAPI, err := NewContentAPI(testServer.URL+"/drafts/content", testServer.URL+"/__gtg", http.DefaultClient)
	assert.NoError(t, err)

	content, err := contentAPI.FetchDraftContent(context.Background(), mocks.ValidMockContentUUID)

	assert.NoError(t, err)
	assert.True(t, content != nil)
	assert.True(t, len(content) > 0)
}
func TestDraftContentAPI_FetchDraftContentMissing(t *testing.T) {

	testServer := mocks.NewDraftContentTestServer(true)
	defer testServer.Close()

	contentAPI, err := NewContentAPI(testServer.URL+"/drafts/content", testServer.URL+"/__gtg", http.DefaultClient)
	assert.NoError(t, err)

	content, err := contentAPI.FetchDraftContent(context.Background(), mocks.MissingMockContentUUID)

	assert.NoError(t, err)
	assert.True(t, content == nil)
}
func TestDraftContentAPI_FetchDraftContentFailure(t *testing.T) {

	testServer := mocks.NewDraftContentTestServer(true)
	testServer.Close()

	contentAPI, err := NewContentAPI(testServer.URL+"/drafts/content", testServer.URL+"/__gtg", http.DefaultClient)
	assert.NoError(t, err)

	content, err := contentAPI.FetchDraftContent(context.Background(), mocks.ValidMockContentUUID)

	assert.Error(t, err)
	assert.True(t, content == nil)
}
