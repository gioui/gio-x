// Package locale can be used to get the end-user preferred language, useful for multi-language apps.
package locale

import (
	"errors"

	"golang.org/x/text/language"
)

var (
	// ErrNotAvailableAPI indicates that the current device/OS doesn't support such function.
	ErrNotAvailableAPI = errors.New("pref: not available api")

	// ErrUnknownLanguage indicates that the current language is not supported by x/text/language.
	ErrUnknownLanguage = errors.New("pref: unknown language")
)

// Language is the preferred language of the end-user or the language of the operating system.
func Language() (language.Tag, error) {
	l := getLanguage()
	if l == "" {
		return language.Tag{}, ErrNotAvailableAPI
	}

	tag, err := language.Parse(l)
	if err != nil {
		return language.Tag{}, ErrUnknownLanguage
	}

	return tag, nil
}
