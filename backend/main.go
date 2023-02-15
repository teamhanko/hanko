/*
Copyright © 2022 Hanko GmbH <developers@hanko.io>
*/
package main

import (
	"github.com/teamhanko/hanko/backend/build_info"
	"github.com/teamhanko/hanko/backend/cmd"
)

func main() {
	build_info.PrintBuildInfo()
	cmd.Execute()
}
