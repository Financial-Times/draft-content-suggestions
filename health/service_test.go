package health

import (
	"testing"
	"net/http/httptest"
	"net/http"
	"github.com/Financial-Times/draft-content-suggestions/draft"
	"github.com/Financial-Times/draft-content-suggestions/suggestions"
	"github.com/coreos/etcd/Godeps/_workspace/src/github.com/stretchr/testify/assert"
)

func TestHealthService_HealthSuccess(t *testing.T) {
	server := newHealthTestServer(t, true)

	contentAPI, _ := draft.NewContentAPI(server.URL+"/drafts/content", server.URL+"/__health", http.DefaultClient)
	umbrellaAPI, _ := suggestions.NewUmbrellaAPI(server.URL+"/content/suggest", http.DefaultClient)

	healthService := NewHealthService("", "", "", contentAPI, umbrellaAPI)

	gtg := healthService.GTG()

	assert.True(t, gtg.GoodToGo)
}

func TestHealthService_HealthFailure(t *testing.T) {
	server := newHealthTestServer(t, false)

	contentAPI, _ := draft.NewContentAPI(server.URL+"/drafts/content", server.URL+"/__health", http.DefaultClient)
	umbrellaAPI, _ := suggestions.NewUmbrellaAPI(server.URL+"/content/suggest", http.DefaultClient)

	healthService := NewHealthService("", "", "", contentAPI, umbrellaAPI)

	gtg := healthService.GTG()

	assert.False(t, gtg.GoodToGo)
}

func newHealthTestServer(t *testing.T, healthy bool) *httptest.Server {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/content/suggest" {
			if !healthy {
				w.WriteHeader(http.StatusServiceUnavailable)
				return
			}

			w.WriteHeader(http.StatusOK)
			return
		}

		if r.URL.Path == "/__health" {
			if !healthy {
				w.WriteHeader(http.StatusServiceUnavailable)
				return
			}

			w.WriteHeader(http.StatusOK)
		}
	}))

	t.Log("Generic external health test server started on:", server.URL)

	return server
}
