package upload

import (
	"fmt"
	"os"

	"github.com/codegangsta/cli"
	"github.com/gronpipmaster/pb"
	"github.com/materials-commons/mcstore/pkg"
)

var Command = cli.Command{
	Name:    "upload",
	Aliases: []string{"up", "u"},
	Usage:   "Upload data to MaterialsCommons",
	Flags:   []cli.Flag{},
	Action:  Cmd,
}

var pbPool = &pb.Pool{}

func processFiles(done <-chan struct{}, entries <-chan pkg.FileEntry, result chan<- string) {
	for entry := range entries {
		fmt.Println("Processing", entry.Path)
		select {
		case result <- entry.Path:
		case <-done:
			fmt.Println("Received done stopping...")
		}
	}
}

func Cmd(c *cli.Context) {
	fmt.Println("upload: ", c.Args())
	if len(c.Args()) != 1 {
		fmt.Println("You must give the directory to walk")
		os.Exit(1)
	}
	dir := c.Args()[0]

	_, errc := pkg.PWalk(dir, 5, processFiles)
	if err := <-errc; err != nil {
		fmt.Println("Got error: ", err)
	}
}
