/*
Copyright © 2022 Hanko GmbH <developers@hanko.io>
*/
package main

import (
	_ "embed"
	"github.com/teamhanko/hanko/backend/cmd"
	"log"
	"runtime/debug"
)

//go:generate bash generate_version.sh
//go:embed version.txt
var Version string

func main() {
	bi, ok := debug.ReadBuildInfo()
	if ok {
		log.Println(bi.Settings)
		log.Println(Version)
	}

	cmd.Execute()
}
