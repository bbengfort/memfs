# MemFS

[![Build Status](https://travis-ci.org/bbengfort/memfs.svg?branch=master)](https://travis-ci.org/bbengfort/memfs)
[![Coverage Status](https://coveralls.io/repos/github/bbengfort/memfs/badge.svg?branch=master)](https://coveralls.io/github/bbengfort/memfs?branch=master)
[![Go Report Card](https://goreportcard.com/badge/github.com/bbengfort/memfs)](https://goreportcard.com/report/github.com/bbengfort/memfs)
[![GoDoc](https://godoc.org/github.com/bbengfort/memfs?status.svg)](https://godoc.org/github.com/bbengfort/memfs)

**In memory file system that implements as many FUSE interfaces as possible.**

For more information see the [godoc documentation](https://godoc.org/github.com/bbengfort/memfs).

## Getting Started

Clone or download the repository into your Go path or use go get as follows:

```
$ go get github.com/bbengfort/memfs
```

Change your working directory to repository. Build and install the MemFS command line tool and add to your `$PATH`:

```
$ go install cmd/memfs.go
```

You should now be able to run the `memfs` command. To see the various command line options:

```
$ memfs --help
```

Finally to mount the MemFS in a directory called `data` in your home directory, with the reasonable defaults from the command line configuration:

```
$ memfs ~/data
```

The FileSystem should start running in the foreground with info log statements, and the mount point should appear in your OS.
