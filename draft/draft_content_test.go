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

	mockDraftContent := &Content{
		UUID:   "6f14ea94-690f-3ed4-98c7-b926683c735a",
		Body:   "<body><content data-embedded=\"true\" id=\"c0cc4ca2-0b43-11e8-24ad-bec2279df517\" type=\"http://www.ft.com/ontology/content/ImageSet\"></content><p>US stocks see-sawed in early trading on Tuesday, as volatility on global markets intensified, breaking an extended period of calm for investors.xxxx</body>",
		Title:  "Wall Street stocks xxx",
		Byline: "Eric Platt in New York, Michael Hunter and Adam Samson in London",
	}

	testServer := mocks.NewDraftContentTestServer(true)
	defer testServer.Close()

	contentAPI, err := NewContentAPI(testServer.URL+"/drafts/content", testServer.URL+"/__gtg", http.DefaultClient)
	assert.NoError(t, err)

	content, err := contentAPI.FetchDraftContent(context.Background(), mocks.ValidMockContentUUID)

	assert.NoError(t, err)
	assert.True(t, content != nil)

	assert.Equal(t, mockDraftContent.UUID, content.UUID)
	assert.Equal(t, mockDraftContent.Byline, content.Byline)
	assert.Equal(t, mockDraftContent.Title, content.Title)
	assert.Equal(t, mockDraftContent.Body, content.Body)
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
