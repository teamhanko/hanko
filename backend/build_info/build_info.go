package build_info

import (
	_ "embed"
	"log"
	"runtime/debug"
	"strings"
)

//go:generate sh -c "git describe --tags --always --match backend/* > version.txt"
//go:embed version.txt
var Version string

func PrintBuildInfo() {
	bi, ok := debug.ReadBuildInfo()
	ActualVersion := strings.TrimSpace(Version)
	if ok {
		modified := false
		for _, v := range bi.Settings {
			if v.Key == "vcs.modified" {
				if v.Value == "true" {
					modified = true
				}
			}
		}
		if modified {
			ActualVersion += "-dirty"
		}
	}
	log.Println(ActualVersion)
}
