package cache

import "fmt"

// Key prefixes for different cache types
const (
	PrefixFormSlug = "form:slug:"
	PrefixFormID   = "form:id:"
)

// FormSlugKey generates cache key for form by slug
func FormSlugKey(slug string) string {
	return fmt.Sprintf("%s%s", PrefixFormSlug, slug)
}

// FormIDKey generates cache key for form by ID
func FormIDKey(formID string) string {
	return fmt.Sprintf("%s%s", PrefixFormID, formID)
}
