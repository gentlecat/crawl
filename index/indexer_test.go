package index

import "testing"

func TestPrepKeyword(t *testing.T) {
	dirtyWord := "Test!123"
	cleanKeyword := Keyword("test123")
	if prepKeyword(dirtyWord) != cleanKeyword {
		t.Error("Didn't clean the keyword properly")
	}
}
