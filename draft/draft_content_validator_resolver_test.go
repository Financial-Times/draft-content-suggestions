package draft

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDraftContentValidatorResolver_ValidatorForContentType(t *testing.T) {
	ucv := NewDraftContentValidatorService("upp-article-endpoint", http.DefaultClient)
	resolver := NewContentValidatorResolver(cctOnlyResolverConfig(ucv))

	uppContentValidator, err := resolver.ValidatorForContentType("application/vnd.ft-upp-article+json; version=1.0; charset=utf-8")

	assert.NoError(t, err, "UPP Validator relies on content-type and originId. Both are present")
	assert.Equal(t, ucv, uppContentValidator, "Should return the same instance impl of DraftContentValidator")
}

func TestDraftContentValidatorResolver_MissingValidation(t *testing.T) {
	resolver := NewContentValidatorResolver(map[string]ContentValidator{})

	validator, err := resolver.ValidatorForContentType("application/vnd.ft-upp-article+json; version=1.0; charset=utf-8")

	assert.Error(t, err)
	assert.Nil(t, validator)
}

func cctOnlyResolverConfig(ucv ContentValidator) (contentTypeToValidator map[string]ContentValidator) {
	return map[string]ContentValidator{
		contentTypeArticle: ucv,
	}
}
