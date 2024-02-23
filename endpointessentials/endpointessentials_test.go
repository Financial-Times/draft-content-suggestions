package endpointessentials

import (
	"testing"

	"github.com/stretchr/testify/assert"
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
	assert.Error(t, isValid)
}
