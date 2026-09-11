package build_info

import (
	_ "embed"
	"runtime/debug"
	"strings"
)

//go:generate sh -c "git describe --tags --always --match backend/* 2>/dev/null | sed -e s#^backend/## > version.txt; test -s version.txt || { test -n \"$HANKO_VERSION\" && printf '%s\n' \"$HANKO_VERSION\" > version.txt; }"
//go:embed version.txt
var version string
var realVersion *string
var isDirty *bool

func GetVersion() string {
	if realVersion == nil {
		tempVersion := strings.TrimSpace(version)
		if getIsDirty() {
			tempVersion += "-dirty"
		}
		realVersion = &tempVersion
	}
	return *realVersion
}

func getIsDirty() bool {
	if isDirty == nil {
		bi, ok := debug.ReadBuildInfo()
		if ok {
			modified := false
			for _, v := range bi.Settings {
				if v.Key == "vcs.modified" {
					if v.Value == "true" {
						modified = true
					}
				}
			}
			isDirty = &modified
		}
	}
	return *isDirty
}
