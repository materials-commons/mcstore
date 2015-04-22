package main

import (
	"os"

	"github.com/codegangsta/cli"
	"github.com/materials-commons/mcstore/cmd/mc/download"
	"github.com/materials-commons/mcstore/cmd/mc/get"
	"github.com/materials-commons/mcstore/cmd/mc/login"
	"github.com/materials-commons/mcstore/cmd/mc/monitor"
	"github.com/materials-commons/mcstore/cmd/mc/receive"
	"github.com/materials-commons/mcstore/cmd/mc/send"
	"github.com/materials-commons/mcstore/cmd/mc/set"
	"github.com/materials-commons/mcstore/cmd/mc/upload"
)

func main() {
	app := cli.NewApp()
	app.Version = "0.0.1"
	app.Authors = []cli.Author{
		{
			Name:  "V. Glenn Tarcea",
			Email: "gtarcea@umich.edu",
		},
	}
	app.Commands = []cli.Command{
		send.Command,
		receive.Command,
		get.Command,
		set.Command,
		upload.Command,
		download.Command,
		monitor.Command,
		login.Command,
	}

	app.Run(os.Args)
}
