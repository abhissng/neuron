package helpers

import (
	"github.com/abhissng/neuron/utils/types"
	"github.com/nicksnyder/go-i18n/v2/i18n"
	"golang.org/x/text/language"
)

// GetDefaultLanguageTag returns the default language tag
func GetDefaultLanguageTag() types.LanguageTag {
	return types.LanguageTag(language.English)
}

// ParseLanguageTag parses a string into a language.Tag and returns a LanguageTag
func ParseLanguageTag(tagString string) types.LanguageTag {
	if tagString == "" {
		return GetDefaultLanguageTag()
	}
	// Parse the string into a language.Tag
	parsedTag, err := language.Parse(tagString)
	if err != nil {
		return GetDefaultLanguageTag()
	}
	return types.LanguageTag(parsedTag)
}

// NewBundle creates a new i18n.Bundle
func NewBundle(language types.LanguageTag) *i18n.Bundle {
	if IsEmpty(language) {
		language = GetDefaultLanguageTag()
	}
	return i18n.NewBundle(types.ToLanguageTag(language))
}
