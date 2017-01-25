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
		fmt.Println(msg)
		os.Exit(1)
	}
}

func main() {

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
			Usage: "specify a path to the configuration `FILE`",
		},
		cli.StringFlag{
			Name:  "name, N",
			Usage: "specify name of host, uses os hostname by default",
		},
		cli.Uint64Flag{
			Name:  "cache, C",
			Usage: "specify maximum cache size in bytes, 4GB by default",
		},
		cli.StringFlag{
			Name:  "level, L",
			Usage: "specify minimum log level, INFO by default",
		},
		cli.BoolFlag{
			Name:  "readonly, R",
			Usage: "set the fs to read only mode, false by default",
		},
	}

	app.Action = runfs
	app.Run(os.Args)

}

func runfs(c *cli.Context) error {

	var err error
	var mountPath string
	var config *memfs.Config

	// Validate the arguments
	if c.NArg() != 1 {
		return cli.NewExitError("please supply the path to the mount point", 1)
	}

	// Get the mount path from the arguments
	mountPath = c.Args()[0]

	// Create the configuration from the passed in file or with defaults
	cpath := c.String("config")
	if config, err = makeConfig(cpath); err != nil {
		return cli.NewExitError(err.Error(), 1)
	}

	// Update the configuration with command line options
	if c.String("name") != "" {
		config.Name = c.String("name")
	}

	if c.Uint64("cache") != 0 {
		config.CacheSize = c.Uint64("cache")
	}

	if c.String("level") != "" {
		config.Level = c.String("level")
	}

	if c.Bool("readonly") {
		config.ReadOnly = c.Bool("readonly")
	}

	// Create the new file system
	fs = memfs.New(mountPath, config)

	// Handle interrupts
	go signalHandler()

	// Run the file system
	if err := fs.Run(); err != nil {
		return cli.NewExitError(err.Error(), 1)
	}

	return nil
}

// Helper function to make the configuration.
func makeConfig(cpath string) (*memfs.Config, error) {
	// Construct configuration from command line options or JSON file.
	config := new(memfs.Config)

	// Load the configuration if a path was passed in.

	if cpath != "" {
		if err := config.Load(cpath); err != nil {
			return nil, err
		}
	} else {
		name, err := os.Hostname()
		if err != nil {
			name = "terp"
		}

		// Add reasonable defaults to the configuration
		config.Name = name
		config.CacheSize = uint64(4295000000)
		config.Level = "info"
		config.ReadOnly = false
		config.Replicas = make([]*memfs.Replica, 0, 0)
	}

	return config, nil
}
