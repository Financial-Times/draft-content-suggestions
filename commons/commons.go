package commons

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/Financial-Times/service-status-go/buildinfo"
	tidutils "github.com/Financial-Times/transactionid-utils-go"
	"github.com/satori/go.uuid"
	"io"
	"net/http"
	"net/url"
	"strings"
)

// Common type/behaviour definition for an endpoint
type Endpoint interface {
	// Endpoint
	// Returns the endpoint
	Endpoint() string

	// IsValid
	// Validates the structure of the url/uri(s)
	IsValid() error

	// IsHealthy
	// Checks if this endpoint is actually reachable and performing as expected
	IsHealthy(ctx context.Context) (string, error)
}

// NewHttpRequest provides a bare minimum request with mandatory headers included
func NewHttpRequest(ctx context.Context, method string, urlStr string, body io.Reader) (*http.Request, error) {
	req, err := http.NewRequest(method, urlStr, body)
	if err != nil {
		return nil, err
	}

	tid, err := tidutils.GetTransactionIDFromContext(ctx)

	if err != nil {
		tid = tidutils.NewTransactionID()
	}

	req.Header.Set(tidutils.TransactionIDHeader, tid)

	req.Header.Set("User-Agent", "PAC-draft-content-suggestions/"+strings.Replace(buildinfo.GetBuildInfo().Version, " ", "-", -1))
	return req, nil
}

type message struct {
	Message string `json:"message"`
}

// WriteJSONMessage writes the msg provided as encoded json with the proper content type header added.
func WriteJSONMessage(w http.ResponseWriter, status int, msg string) error {
	w.WriteHeader(status)
	w.Header().Add("Content-Type", "application/json")

	enc := json.NewEncoder(w)
	return enc.Encode(&message{Message: msg})
}

// NewContextFromRequest provides a new context including a trxId
// from the request or if missing, a brand new trxId.
func NewContextFromRequest(r *http.Request) context.Context {
	return tidutils.TransactionAwareContext(context.Background(), tidutils.GetTransactionIDFromRequest(r))
}

// ValidateEndpoints provides url/uri level validation, it does not make any actual http(s) requests
func ValidateEndpoint(endpoint string) error {

	if !strings.HasPrefix(endpoint, "http://") && !strings.HasPrefix(endpoint, "https://") {
		return errors.New(fmt.Sprintf("Missing scheme in endpoint: %v", endpoint))
	}
	_, err := url.ParseRequestURI(endpoint)

	if err != nil {
		return errors.New(fmt.Sprintln("Invalid endpoint configuration:", err, " for:", endpoint))
	}

	return nil
}

// ValidateUUID checks the uuid string for supported formats
func ValidateUUID(u string) error {
	_, err := uuid.FromString(u)
	return err
}
