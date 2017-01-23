# Shell to use with Make
SHELL := /bin/bash

# Export targets not associated with files.
.PHONY: fmt test clean

# Format the Go source code
fmt:
	@echo "Formatting the source"
	-gofmt -w .

# Target for simple testing on the command line
test:
	go test -v -coverprofile=sequence.coverprofile github.com/bbengfort/sequence

# Clean build files
clean:
	@echo "Cleaning up the project source."
	-go clean
	-find . -name "*.coverprofile" -print0 | xargs -0 rm -rf
	-rm -rf site
	-rm -rf _bin
	-rm -rf _build
