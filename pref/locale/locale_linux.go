//+build !android

package locale

import (
	"os"
	"strings"
)

func getLanguage() string {
	lang := os.Getenv("LANG")
	if lang == "" {
		return ""
	}

	// Strip the ".UTF-8" (or equivalent) from the language.
	langs := strings.Split(lang, ".")
	if len(langs) < 1 {
		return ""
	}

	return langs[0]
}
