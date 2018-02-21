package draft

import (
	"context"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testing"
)

const validContentUUID = "9d5e441e-0b02-11e8-8eb7-42f857ea9f0"
const missingContentUUID = "9d5e441e-0b02-11e8-1234-42f857ea9f1"
const draftContent = `
{  
   "uuid":"9d5e441e-0b02-11e8-8eb7-42f857ea9f0",
   "title":"Wall Street stocks xxx",
   "alternativeTitles":{  
      "promotionalTitle":"Wall Street volatile amid global equities rout",
      "contentPackageTitle":null
   },
   "type":"Article",
   "byline":"Eric Platt in New York, Michael Hunter and Adam Samson in London",
   "brands":[  
      {  
         "id":"http://api.ft.com/things/dbb0bdae-1f0c-11e4-b0cb-b2227cce2b54"
      }
   ],
   "identifiers":[  
      {  
         "authority":"http://api.ft.com/system/FTCOM-METHODE",
         "identifierValue":"9d5e441e-0b02-11e8-8eb7-42f857ea9f09"
      }
   ],
   "publishedDate":"2018-02-06T16:17:08.000Z",
   "standfirst":"Gauge of US market turbulence hits 50 for first time since 2015 before retreating",
   "body":"<body><content data-embedded=\"true\" id=\"c0cc4ca2-0b43-11e8-24ad-bec2279df517\" type=\"http://www.ft.com/ontology/content/ImageSet\"></content><p>US stocks see-sawed in early trading on Tuesday, as volatility on global markets intensified, breaking an extended period of calm for investors.xxxx</body>",
   "description":null,
   "mediaType":null,
   "pixelWidth":null,
   "pixelHeight":null,
   "internalBinaryUrl":null,
   "externalBinaryUrl":null,
   "members":null,
   "mainImage":"c0cc4ca2-0b43-11e8-24ad-bec2279df517",
   "standout":{  
      "editorsChoice":false,
      "exclusive":false,
      "scoop":false
   },
   "comments":{  
      "enabled":true
   },
   "copyright":null,
   "webUrl":null,
   "publishReference":"tid_7pzh8cvkue",
   "lastModified":"2018-02-06T16:17:08.781Z",
   "canBeSyndicated":"yes",
   "firstPublishedDate":"2018-02-06T07:31:29.000Z",
   "accessLevel":"subscribed",
   "canBeDistributed":"yes",
   "rightsGroup":null,
   "masterSource":null,
   "alternativeStandfirsts":{  
      "promotionalStandfirst":null
   }
}`

func TestDraftContentAPI_IsHealthySuccess(t *testing.T) {

	testServer := newDraftContentTestServer(t, true)
	defer testServer.Close()

	contentAPI, err := NewContentAPI(testServer.URL+"/drafts/content", testServer.URL+"/__health", http.DefaultClient)

	assert.NoError(t, err)

	_, err = contentAPI.IsHealthy(context.Background())
	assert.NoError(t, err)
}
func TestDraftContentAPI_IsHealthyFailure(t *testing.T) {
	testServer := newDraftContentTestServer(t, false)
	defer testServer.Close()

	contentAPI, err := NewContentAPI(testServer.URL+"/drafts/content", testServer.URL+"/__health", http.DefaultClient)

	assert.NoError(t, err)

	_, err = contentAPI.IsHealthy(context.Background())
	assert.Error(t, err)
}
func TestDraftContentAPI_FetchDraftContentSuccess(t *testing.T) {

	mockDraftContent := &Content{
		UUID:   "9d5e441e-0b02-11e8-8eb7-42f857ea9f0",
		Body:   "<body><content data-embedded=\"true\" id=\"c0cc4ca2-0b43-11e8-24ad-bec2279df517\" type=\"http://www.ft.com/ontology/content/ImageSet\"></content><p>US stocks see-sawed in early trading on Tuesday, as volatility on global markets intensified, breaking an extended period of calm for investors.xxxx</body>",
		Title:  "Wall Street stocks xxx",
		Byline: "Eric Platt in New York, Michael Hunter and Adam Samson in London",
	}

	testServer := newDraftContentTestServer(t, true)
	defer testServer.Close()

	contentAPI, err := NewContentAPI(testServer.URL+"/drafts/content", testServer.URL+"/__health", http.DefaultClient)
	assert.NoError(t, err)

	content, err := contentAPI.FetchDraftContent(context.Background(), validContentUUID)

	assert.NoError(t, err)
	assert.True(t, content != nil)

	assert.True(t, mockDraftContent.UUID == content.UUID)
	assert.True(t, mockDraftContent.Byline == content.Byline)
	assert.True(t, mockDraftContent.Title == content.Title)
	assert.True(t, mockDraftContent.Body == content.Body)
}
func TestDraftContentAPI_FetchDraftContentMissing(t *testing.T) {

	testServer := newDraftContentTestServer(t, true)
	defer testServer.Close()

	contentAPI, err := NewContentAPI(testServer.URL+"/drafts/content", testServer.URL+"/__health", http.DefaultClient)
	assert.NoError(t, err)

	content, err := contentAPI.FetchDraftContent(context.Background(), missingContentUUID)

	assert.NoError(t, err)
	assert.True(t, content == nil)
}
func TestDraftContentAPI_FetchDraftContentFailure(t *testing.T) {

	testServer := newDraftContentTestServer(t, true)
	testServer.Close()

	contentAPI, err := NewContentAPI(testServer.URL+"/drafts/content", testServer.URL+"/__health", http.DefaultClient)
	assert.NoError(t, err)

	content, err := contentAPI.FetchDraftContent(context.Background(), validContentUUID)

	assert.Error(t, err)
	assert.True(t, content == nil)
}

func newDraftContentTestServer(t *testing.T, healthy bool) *httptest.Server {

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/drafts/content/" + validContentUUID:
			w.WriteHeader(200)
			w.Write([]byte(draftContent))
		case "/drafts/content/" + missingContentUUID:
			w.WriteHeader(404)
		case "/__health":
			fallthrough
		case "__health":
			if healthy {
				w.WriteHeader(200)
			} else {
				w.WriteHeader(503)
			}
		}
	}))

	t.Log("Draft content test server started on: ", server.URL)

	return server
}
