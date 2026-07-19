package shared

import (
	"regexp"
	"strings"
	"unicode"
)

var slugRe = regexp.MustCompile(`[^a-z0-9]+`)

// Slugify converts a name to a URL-safe slug.
func Slugify(name string) string {
	s := strings.ToLower(name)
	s = strings.TrimFunc(s, func(r rune) bool {
		return !unicode.IsLetter(r) && !unicode.IsDigit(r)
	})
	s = slugRe.ReplaceAllString(s, "-")
	s = strings.Trim(s, "-")
	return s
}
