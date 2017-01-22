package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/bbengfort/memfs"
	"github.com/urfave/cli"
)

var fs *memfs.FileSystem
var logger *memfs.Logger

//===========================================================================
// OS Signal Handlers
//===========================================================================

func signalHandler() {
	// Make signal channel and register notifiers for Interupt and Terminate
	sigchan := make(chan os.Signal, 1)
	signal.Notify(sigchan, os.Interrupt)
	signal.Notify(sigchan, syscall.SIGTERM)

	// Block until we receive a signal on the channel
	<-sigchan

	// Defer the clean exit until the end of the function
	defer os.Exit(0)

	// Shutdown now that we've received the signal
	err := fs.Shutdown()
	if err != nil {
		msg := fmt.Sprintf("shutdown error: %s", err.Error())
		logger.Fatal(msg)
		os.Exit(1)
	}
}

func main() {
	// Handle interrupts
	go signalHandler()

	var err error
	logger, err = memfs.InitLogger("", "INFO")
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	app := cli.NewApp()
	app.Name = "memfs"
	app.Usage = "In memory file system with anti-entropy replication"
	app.ArgsUsage = "mount point"
	app.Version = memfs.PackageVersion()
	app.Author = "Benjamin Bengfort"
	app.Email = "bengfort@cs.umd.edu"
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:  "config, c",
			Value: "conf/memfs.json",
			Usage: "specify a path to the configuration `FILE`",
		},
	}

	app.Action = runfs
	app.Run(os.Args)

}

func runfs(c *cli.Context) error {

	var mountPath string
	var config *memfs.Config

	// Get the mount path from the arguments
	if c.NArg() != 1 {
		return cli.NewExitError("please supply the path to the mount point", 1)
	}

	mountPath = c.Args()[0]

	// Load the configuration from the flag
	cpath := c.String("config")
	if cpath == "" {
		return cli.NewExitError("please supply the path to a configuration file", 1)
	}

	config = new(memfs.Config)
	if err := config.Load(cpath); err != nil {
		return cli.NewExitError(err.Error(), 1)
	}

	logger.Info("loaded configuration from %s", cpath)

	fs = memfs.New(mountPath, config)
	if err := fs.Run(); err != nil {
		return cli.NewExitError(err.Error(), 1)
	}

	return nil
}
