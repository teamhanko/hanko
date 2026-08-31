package build_info

import "testing"

const emptyVersionError = "GetVersion() returned an empty version"

func TestGetVersionIsNotEmpty(t *testing.T) {
	if GetVersion() == "" {
		t.Fatal(emptyVersionError)
	}
}
