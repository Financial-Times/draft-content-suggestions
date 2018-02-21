package commons

import "testing"
import (
	"context"
	"fmt"
	"github.com/Financial-Times/transactionid-utils-go"
	"github.com/satori/go.uuid"
	"github.com/stretchr/testify/assert"
	"net/http"
	"strings"
)

const (
	validEndpoint   = "http://some.valid.url/with/sub/resources:8080"
	invalidEndpoint = "/missing.com/scheme/or/uri/type"
	invalidUUID     = "random:invalid-string"
)

func TestValidateEndpointSuccess(t *testing.T) {
	assert.NoError(t, ValidateEndpoint(validEndpoint))
}

func TestValidateEndpointFailure(t *testing.T) {
	isValid := ValidateEndpoint(invalidEndpoint)

	if isValid == nil {
		assert.Fail(t, fmt.Sprintf("Endpoint validation should've failed for: %v", invalidEndpoint))
	}
}

func TestValidateUUIDSuccess(t *testing.T) {
	v4 := uuid.NewV4()
	assert.NoError(t, ValidateUUID(v4.String()))
}

func TestValidateUUIDFailure(t *testing.T) {
	isValid := ValidateUUID(invalidUUID)

	if isValid == nil {
		assert.Fail(t, fmt.Sprintf("UUID validation should've failed for: %v", invalidUUID))
	}
}

func TestNewContextFromRequest(t *testing.T) {
	request, _ := http.NewRequest(http.MethodGet, validEndpoint, nil)
	contextFromRequest := NewContextFromRequest(request)

	trxId, ok := contextFromRequest.Value(transactionidutils.TransactionIDKey).(string)

	assert.True(t, ok)
	assert.True(t, trxId != "")
}

func TestNewHttpRequestShouldContainTrxIdHeader(t *testing.T) {
	request, err := NewHttpRequest(context.Background(), http.MethodGet, validEndpoint, nil)

	assert.NoError(t, err)
	assert.True(t, request.Header.Get(transactionidutils.TransactionIDHeader) != "")
}
func TestNewHttpRequestShouldContainUserAgentHeader(t *testing.T) {
	request, err := NewHttpRequest(context.Background(), http.MethodGet, validEndpoint, nil)

	assert.NoError(t, err)
	assert.True(t, strings.HasPrefix(request.Header.Get("User-Agent"), "PAC-draft-content-suggestions/"))
}
