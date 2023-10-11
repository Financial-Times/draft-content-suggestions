package mocks

import (
	"encoding/base64"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
)

const ValidMockContentUUID = "6f14ea94-690f-3ed4-98c7-b926683c735a"

const FailsRetrivalContentUuid = "7597bf73-4420-4194-98ef-4e5b1e8267e9"

const MissingMockContentUUID = "711e5bc1-3470-4297-ae26-154f145a6287"

const UnprocessableContentUUID = "910b60e8-13d8-4b51-871a-d29cf21eb583"

const AuthorizationHeader = "Authorization"

const MockSuggestions = `{
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

const MockDraftContent = `
{  
   "uuid":"6f14ea94-690f-3ed4-98c7-b926683c735a",
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

func NewDraftContentTestServer(healthy bool) *httptest.Server {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		switch r.URL.Path {
		case "/drafts/content/" + ValidMockContentUUID:
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(MockDraftContent))
		case "/drafts/content/" + MissingMockContentUUID:
			w.WriteHeader(http.StatusNotFound)
		case "/drafts/content/" + UnprocessableContentUUID:
			w.WriteHeader(http.StatusUnprocessableEntity)
		case "/drafts/content/" + FailsRetrivalContentUuid:
			w.WriteHeader(http.StatusInternalServerError)
		case "/__gtg", "__gtg":
			if healthy {
				w.WriteHeader(http.StatusOK)
			} else {
				w.WriteHeader(http.StatusServiceUnavailable)
			}
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))

	return server
}

func NewUmbrellaTestServer(healthy bool) *httptest.Server {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if basicAuth := r.Header.Get(AuthorizationHeader); basicAuth != createBasicAuth("username", "password") {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		switch r.URL.Path {
		case "/content/suggest":
			if r.Header.Get("X-Origin") != "PAC" {
				w.WriteHeader(http.StatusBadRequest)
				w.Write([]byte(`{ message: "Request has invalid 'Origin' header" }`))
				return
			}
			if !healthy {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			bytes, err := io.ReadAll(r.Body)

			if err != nil || bytes == nil {
				w.WriteHeader(http.StatusBadRequest)
				return
			}

			content := make(map[string]interface{})
			err = json.Unmarshal(bytes, &content)

			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				return
			}

			w.WriteHeader(200)
			w.Write([]byte(MockSuggestions))
		case "/content/suggest/__gtg":
			if !healthy {
				w.WriteHeader(http.StatusServiceUnavailable)
				return
			}
			w.WriteHeader(http.StatusOK)
		}
	}))

	return server
}

func createBasicAuth(testUsername string, testPassword string) string {
	return "Basic " + base64.StdEncoding.EncodeToString([]byte(strings.Join([]string{testUsername, testPassword}, ":")))
}
