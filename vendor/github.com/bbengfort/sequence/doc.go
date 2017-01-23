// Package sequence provides monotonically increasing counters similar to the
// autoincrement sequence objects that are available in databases like
// PostgreSQL. While this type of object is simple to implement, stateful
// representations are used in a wide variety of projects and this package is
// about having a standard API and interface to reuse these objects without
// fear. In particular, this project does not implement simple counters or
// iterables, but is designed to maximize the sequence space for constant
// growth and to raise errors when sequences become too large.
//
// The primary interface is the Sequence object, which essentially wraps a
// 64-bit unsigned integer, providing the single largest positive integer
// range. The Sequence object implements the Incrementer interface, which
// specifies how to interact with the Sequence object.
//
// Simple usage is as follows:
//
//     import github.com/bbengfort/sequence
//
//     seq := sequence.New()
//     idx, err := seq.Next()
//
// By default, this will provide an integer sequence in the positive range
// 1 - 18,446,744,073,709,551,614. Note that sequences start at one as they
// are typically used to compute autoincrement positive ids.
//
// The Sequence.Next() method will return an error if it reaches the maximum
// bound, which by default is the maximal uint64 value, such that incrementing
// will not start to repeat values. If the increment is negative, then the
// sequence will return an error if it reaches a minimum bound, which by
// default is 0 since the Sequence will always return positive values.
//
// The Sequence object provides several helper methods to interact with it
// during long running processes, including Current(), IsStarted(), and
// String() methods which return information about the state of the Sequence.
//
// Interacting with sequences across processes is provided through the
// Sequence.Dump and Sequence.Load methods, which serialize the Sequence to a
// []byte JSON representation. Note that this does not use the standard
// json.Marshal and json.Unmarshal interface in order to keep members of the
// sequence inaccessible outside the library, ensuring that a sequence cannot
// be modified except to be restarted.
package sequence
