package main

import (
	"os"

	"github.com/bbengfort/memfs"
	"github.com/urfave/cli"
)

func main() {

	app := cli.NewApp()
	app.Name = "memfs"
	app.Usage = "In memory file system with anti-entropy replication"
	app.Version = memfs.PackageVersion()
	app.Author = "Benjamin Bengfort"
	app.Email = "bengfort@cs.umd.edu"

	app.Run(os.Args)

}
