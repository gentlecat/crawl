package html_cleaner

import "testing"

func TestClean(t *testing.T) {
	htmlString := "<html><body><s>Test</s>!</body></html>"
	if Clean(htmlString) != "Test!" {
		t.Error("Didn't clean the HTML properly")
	}
}
