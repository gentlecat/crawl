package classifier

import (
	"strings"
)

// IsTopical function determines if a page matches a topic or not. Topic is
// represented by a set of keywords.
//
// It is recommended to clean up the page before processing. For example, if
// the page is in HTML format (which they would be in most cases), then it
// would be a good idea to remove all the tags and do some other kind of
// processing, if necessary.
func IsTopical(pageContent string, keywords []string) bool {
	pageContent = strings.ToUpper(pageContent)
	// Also converting both page content strings and keywords to
	// make the search case insensitive.
	for _, keyword := range keywords {
		if !strings.Contains(pageContent, strings.ToUpper(keyword)) {
			return false
		}
	}
	return true
}
