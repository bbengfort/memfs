# Go-Sequence

[![Build Status](https://travis-ci.org/bbengfort/sequence.svg?branch=master)](https://travis-ci.org/bbengfort/sequence)
[![Coverage Status](https://coveralls.io/repos/github/bbengfort/sequence/badge.svg?branch=master)](https://coveralls.io/github/bbengfort/sequence?branch=master)
[![Go Report Card](https://goreportcard.com/badge/github.com/bbengfort/sequence)](https://goreportcard.com/report/github.com/bbengfort/sequence)
[![GoDoc](https://godoc.org/github.com/bbengfort/sequence?status.svg)](https://godoc.org/github.com/bbengfort/sequence)

**Implements an AutoIncrement counter class similar to PostgreSQL's sequence.**

This library provides monotonically increasing and decreasing sequences similar to the autoincrement sequence objects that are available in databases like PostgreSQL. Unlike simple counter objects, `Sequence` objects are bred for safety, that is they expose the largest possible range of positive integers using the `uint64` data type and raise exceptions when that number overflows or when the increment function does something unexpected. Moreover, the internal state of a `Sequence` is not accessible by external libraries and therefore cannot be modified (except for a straight-up reset), giving developers confidence to use these objects in sequence-critical usage such as automatically incrementing IDs.

## Getting Started

To install the sequence library, simply `go get` it from GitHub:

```
$ go get github.com/bbengfort/sequence
```

For more specifics, please read the [API documentation](https://godoc.org/github.com/bbengfort/sequence).

### Basic Usage

The basic usage is to create a default, monotonically incrementing by 1 sequence that starts at 1 and goes until 18,446,744,073,709,551,614 (the largest possible `uint64` value).

```go
import github.com/bbengfort/sequence

seq := sequence.New()
idx, err := seq.Next()
```

A range can be specified using different arguments to New. For example, to specify a different maximal value:

```go
seq := sequence.New(1000)
```

Will produce a sequence from 1 to 1000 (inclusive) and will return errors when the state goes beyond 1000. A specifically bounded range:

```go
seq := sequence.New(10, 100)
```

Will provide a sequence on the integers from 10 to 100 (inclusive). Finally a step can be provided to determine how the sequence is incremented:

```go
seq := sequence.New(2, 500, 2)
```

Which will return all the even numbers from 2 to 500 (inclusive). Sequences can be reset, returning them to their original state as follows:

```go
err := seq.Reset()
```

### Sequence State

To get the state of a sequence, you can use the following methods:

```go
seq.IsStarted() // Returns a boolean value if the Sequence is started.

// Get the current value of the sequence
idx, err := seq.Current()

// Print a string representation of the sequence state.
fmt.Println(seq.String())
```

You can also serialize and deserialize the Sequence to pass it across processes as follows:

```go
data, err := seq.Dump()
seq2 := &sequence.Sequence{}
err := seq2.Load(data)
```

This snippet of code will result in `seq2` having an identical state to `seq` at the moment that it was dumped.

## Development

Pull requests are more than welcome to help develop this project!

### Testing

To execute tests for this library:

```
$ make test
```
