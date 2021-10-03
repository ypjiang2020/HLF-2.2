/*
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package main

import (
	"os"

	"github.com/Yunpeng-J/HLF-2.2/bccsp/factory"
	"github.com/Yunpeng-J/HLF-2.2/cmd/common"
	discovery "github.com/Yunpeng-J/HLF-2.2/discovery/cmd"
)

func main() {
	factory.InitFactories(nil)
	cli := common.NewCLI("discover", "Command line client for fabric discovery service")
	discovery.AddCommands(cli)
	cli.Run(os.Args[1:])
}
