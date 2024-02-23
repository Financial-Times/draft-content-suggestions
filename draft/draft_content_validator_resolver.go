package draft

import (
	"fmt"
	"strings"
)

// ContentValidatorResolver manages the validators available for a given originId/content-type pair.
type ContentValidatorResolver interface {
	// ValidatorForContentType Resolves and returns a ContentValidator implementation if present.
	ValidatorForContentType(contentType string) (ContentValidator, error)
}

// NewContentValidatorResolver returns a ContentValidatorResolver implementation
func NewContentValidatorResolver(contentTypeToValidator map[string]ContentValidator) ContentValidatorResolver {
	return &contentValidatorResolver{contentTypeToValidator}
}

type contentValidatorResolver struct {
	contentTypeToValidator map[string]ContentValidator
}

// ValidatorForContentType implementation checks the content-type validation for a validator resolution.
func (resolver *contentValidatorResolver) ValidatorForContentType(contentType string) (ContentValidator, error) {
	contentType = stripMediaTypeParameters(contentType)
	validator, found := resolver.contentTypeToValidator[contentType]

	if !found {
		return nil, fmt.Errorf(
			"no validator configured for contentType: %s\ncontentTypeMap: %v",
			contentType,
			resolver.contentTypeToValidator,
		)
	}

	return validator, nil
}

func stripMediaTypeParameters(contentType string) string {
	if strings.Contains(contentType, ";") {
		contentType = strings.Split(contentType, ";")[0]
	}
	return contentType
}
