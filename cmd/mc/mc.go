package main

import (
	"fmt"
	"os"

	"github.com/codegangsta/cli"
	"github.com/materials-commons/mcstore/cmd/mc/login"
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
		{
			Name:  "send",
			Usage: "Send data over the air",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "name, n",
					Usage: "name of receiving service",
				},
				cli.StringFlag{
					Name:  "project, p",
					Usage: "project to send from",
				},
				cli.StringFlag{
					Name:  "directory, d",
					Usage: "directory to send files from",
				},
				cli.StringFlag{
					Name:  "file, f",
					Usage: "file to send",
				},
			},
			Action: func(c *cli.Context) {
				fmt.Println("send:", c.Args())
			},
		},
		{
			Name:  "get",
			Usage: "Get data over the air",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "name, n",
					Usage: "name of receiving service",
				},
			},
			Action: func(c *cli.Context) {
				fmt.Println("get:", c.Args())
			},
		},
		{
			Name:  "upload",
			Usage: "Upload data to MaterialsCommons",
			Flags: []cli.Flag{},
			Action: func(c *cli.Context) {
				fmt.Println("upload:", c.Args())
			},
		},
		{
			Name:  "download",
			Usage: "Download data from MaterialsCommons",
			Flags: []cli.Flag{},
			Action: func(c *cli.Context) {
				fmt.Println("download:", c.Args())
			},
		},
		{
			Name:   "login",
			Usage:  "Login to MaterialsCommons",
			Flags:  []cli.Flag{},
			Action: login.Cmd,
		},
	}
	app.Run(os.Args)
}
