package cleaner

// Clean cleans up HTML page for further processing.
//
// Currently it just removes all the tags from the page.
func Clean(pageContent string) string {
	return StripTags(pageContent)
}
