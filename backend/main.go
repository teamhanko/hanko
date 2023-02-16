/*
Copyright © 2022 Hanko GmbH <developers@hanko.io>
*/
package main

import (
	"github.com/teamhanko/hanko/backend/build_info"
	"github.com/teamhanko/hanko/backend/cmd"
	"log"
)

func main() {
	log.Println(build_info.GetVersion())
	log.Println(build_info.GetVersion())
	cmd.Execute()
}
