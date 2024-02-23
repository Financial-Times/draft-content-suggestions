package endpointessentials

import (
	"context"
	"errors"
	"fmt"
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

	// IsGTG
	// Checks if this endpoint is actually reachable and performing as expected
	IsGTG(ctx context.Context) (string, error)
}

// ValidateEndpoints provides url/uri level validation, it does not make any actual http(s) requests
func ValidateEndpoint(endpoint string) error {
	if !strings.HasPrefix(endpoint, "http://") && !strings.HasPrefix(endpoint, "https://") {
		return fmt.Errorf("missing scheme in endpoint: %s", endpoint)
	}
	_, err := url.ParseRequestURI(endpoint)

	if err != nil {
		return errors.New(fmt.Sprintln("Invalid endpoint configuration:", err, " for:", endpoint))
	}

	return nil
}
