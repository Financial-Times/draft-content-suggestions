package suggestions

import (
	"context"
	"encoding/json"
	"github.com/Financial-Times/draft-content-suggestions/draft"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
)

const sampleSuggestions = `{
    "suggestions": [
        {
            "predicate": "http://www.ft.com/ontology/annotation/mentions"
            "id": "http://www.ft.com/thing/6f14ea94-690f-3ed4-98c7-b926683c735a",
            "apiUrl": "http://api.ft.com/people/6f14ea94-690f-3ed4-98c7-b926683c735a",
            "prefLabel": "Donald Kaberuka",
            "type": "http://www.ft.com/ontology/person/Person",
            "isFTAuthor": false
        },
        {
            "predicate": "http://www.ft.com/ontology/annotation/mentions"
            "id": "http://www.ft.com/thing/9a5e3b4a-55da-498c-816f-9c534e1392bd",
            "apiUrl": "http://api.ft.com/people/9a5e3b4a-55da-498c-816f-9c534e1392bd",
            "prefLabel": "Lawrence Summers",
            "type": "http://www.ft.com/ontology/person/Person",
            "isFTAuthor": true
        }
     ]
}`

func TestUmbrellaAPI_IsHealthySuccess(t *testing.T) {

	testServer := newUmbrellaTestServer(t, true)
	defer testServer.Close()

	umbrellaAPI, err := NewUmbrellaAPI(testServer.URL+"/content/suggest", http.DefaultClient)

	assert.NoError(t, err)

	_, err = umbrellaAPI.IsHealthy(context.Background())
	assert.NoError(t, err)
}
func TestUmbrellaAPI_IsHealthyFailure(t *testing.T) {
	testServer := newUmbrellaTestServer(t, false)
	defer testServer.Close()

	umbrellaAPI, err := NewUmbrellaAPI(testServer.URL+"/content/suggest", http.DefaultClient)

	assert.NoError(t, err)

	_, err = umbrellaAPI.IsHealthy(context.Background())
	assert.Error(t, err)
}
func TestUmbrellaAPI_FetchSuggestions(t *testing.T) {

	mockDraftContent := mockDraftContent()

	testServer := newUmbrellaTestServer(t, true)
	defer testServer.Close()

	umbrellaAPI, err := NewUmbrellaAPI(testServer.URL+"/content/suggest", http.DefaultClient)
	assert.NoError(t, err)

	suggestions, err := umbrellaAPI.FetchSuggestions(context.Background(), mockDraftContent)

	assert.NoError(t, err)
	assert.True(t, suggestions != nil)
}
func TestUmbrellaAPI_FetchDraftContentFailure(t *testing.T) {

	testServer := newUmbrellaTestServer(t, true)
	testServer.Close()

	contentAPI, err := NewUmbrellaAPI(testServer.URL+"/content/suggest", http.DefaultClient)
	assert.NoError(t, err)

	suggestions, err := contentAPI.FetchSuggestions(context.Background(), mockDraftContent())

	assert.Error(t, err)
	assert.True(t, suggestions == nil)
}

func newUmbrellaTestServer(t *testing.T, healthy bool) *httptest.Server {

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/content/suggest" {
			if !healthy {
				w.WriteHeader(http.StatusServiceUnavailable)
				return
			}

			bytes, err := ioutil.ReadAll(r.Body)

			if err != nil || bytes == nil {
				w.WriteHeader(http.StatusBadRequest)
				return
			}

			var content draft.Content
			err = json.Unmarshal(bytes, &content)

			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				return
			}

			w.WriteHeader(200)
			w.Write([]byte(sampleSuggestions))
		}
	}))
	t.Log("Umbrella endpoint test server started on:", server.URL)

	return server
}

func mockDraftContent() *draft.Content {
	mockDraftContent := &draft.Content{
		UUID:   "9d5e441e-0b02-11e8-8eb7-42f857ea9f0",
		Body:   "<body><content data-embedded=\"true\" id=\"c0cc4ca2-0b43-11e8-24ad-bec2279df517\" type=\"http://www.ft.com/ontology/content/ImageSet\"></content><p>US stocks see-sawed in early trading on Tuesday, as volatility on global markets intensified, breaking an extended period of calm for investors.xxxx</body>",
		Title:  "Wall Street stocks xxx",
		Byline: "Eric Platt in New York, Michael Hunter and Adam Samson in London",
	}
	return mockDraftContent
}
