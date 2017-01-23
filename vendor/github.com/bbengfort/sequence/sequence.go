package sequence

import (
	"encoding/json"
	"errors"
	"fmt"
)

const maxuint64 = ^uint64(0) - 1

// MinimumBound specifies the smallest integer value of a Sequence object.
// Note that the minimum bound is not zero because zero values denote
// unintialized numeric values, and because counters generally do not index
// from zero, but rather from 1.
const MinimumBound = 1

// MaximumBound specifies the largest integer value of a Sequence object.
// This bound is currently set to the largest possible unsigned 64-bit
// integer: 18,446,744,073,709,551,614.
const MaximumBound = maxuint64

//===========================================================================
// Sequence Structs and Interfaces
//===========================================================================

// Incrementer defines the interface for sequence-like objects. The primary
// diference between an Incrementer and other state-like iterables is that error
// handling is critical, and as a result many interaction methods like Next()
// and Restart() return an error. An Incrementer usually keep the state as
// private as possible to ensure that it can't be tampered with accidentally,
// and state is only revealed through the Current(), IsStarted(), and String()
// methods.
type Incrementer interface {
	Init(params ...uint64) error // Initialize the Incrementer with values
	Next() (uint64, error)       // Get the next value in the sequence and update
	Restart() error              // Restarts the sequence if possible
	Current() (uint64, error)    // Returns the current value of the Incrementer
	IsStarted() bool             // Returns the state of the Incrementer
	String() string              // Returns a string representation of the state
	Load(data []byte) error      // Load the sequence from a serialized representation
	Dump() ([]byte, error)       // Dump the sequence to a serialized representation
}

// Sequence implements an AutoIncrement counter class similar to the
// PostgreSQL sequence object. Sequence is the primary implementation of the
// Incrementer interface. Once a Sequence has been constructed either with the
// New() function that is part of this package or manually, it is expected
// that the state of the Sequence only changes via the two interaction methods
// - Next() and Restart(). Therefore, all internal members of Sequence are
// private to ensure outside libraries that the internal state of the sequence
// cannot be modified accidentally.
//
// Sequences can be created manually as follows:
//
//     seq := new(Sequence)
//     seq.Init()
//
// However it is recommended to create Sequences with the New() function:
//
//     import github.com/bbengfort/sequence
//     seq := sequence.New()
//
// Sequence objects are intended to act as monotonically increasing counters,
// maximizing the positive integer value range. Sequences can be constructed
// to be monotonically decreasing counters in the positive range or
// constrained by different bounds by passing different arguments to the New()
// function or to the Init() method.
//
// Sequences can also be serialized and deserialized to be passed between
// processes while still maintaining state. The final mechanism to create a
// Sequence is as follows:
//
//     data, err := seq.Dump()
//     if err == nil {
//         seq2 := &Sequence{}
//         seq2.Load(data)
//     }
//
// This will create a second Sequence (seq2) that is identical to the state of
// the first Sequence (seq) when it was dumped.
type Sequence struct {
	current     uint64 // The current value of the sequence
	increment   uint64 // The value to increment by (usually 1)
	minvalue    uint64 // The minimum value of the counter (usually 1)
	maxvalue    uint64 // The max value of the counter (usually bounded by type)
	initialized bool   // Flag that indicates if the sequence has been initialized.
}

// New constructs a Sequence object, and is the simplest way to create a new
// monotonically increasing counter. By default (passing in no arguments),
// the Sequence counts by 1 from 1 to the MaximumBound value without returning
// errors. Additional arguments can be supplied, as described by the Init()
// method. In brief, zero or more of the following arguments can be supplied:
// maximum value, minimum value, and step. However the number and order of the
// arguments matters, see Init() for details.
//
// The two most common Sequence constructors are:
//
//     seq := sequence.New() // monotonically increasing counter
//
// Which counts from 1 until the max bound and is equivalent to:
//
//     seq := sequence.New(1, sequence.MaximumBound, 1)
//
// Because the initialization is somewhat complex, New() can return an error,
// which is also defined by the Init() method.
func New(params ...uint64) (*Sequence, error) {
	seq := new(Sequence)
	err := seq.Init(params...)
	return seq, err
}

//===========================================================================
// Sequence Interaction Methods
//===========================================================================

// Init a sequence with reasonable defaults based on the number and order of
// the numeric parameters passed into this method. By default, if no arguments
// are passed into Init, then the Sequence will be initialized as a
// monotonically increasing counter in the positive space as follows:
//
//     seq.Init() // count by 1 from 1 to MaximumBound
//
// If only a single argument is passed in, then it is interpreted as the
// maximum bound as follows:
//
//     seq.Init(100) // count by 1 from 1 until 100.
//
// If two arguments are passed in, then it is interpreted as a discrete range.
//
//     seq.Init(10, 100) // count by 1 from 10 until 100.
//
// If three arguments are passed in, then the third is the step.
//
//     seq.Init(2, 100, 2) // even numbers from 2 until 100.
//
// Both endpoints of these ranges are inclusive.
//
// Init can return a variety of errors. The most common error is if Init is
// called twice - that is that an already initialized sequence is attempted to
// be modified in a way that doesn't reset it. This is part of the safety
// features that Sequence provides. Other errors include mismatched or
// non-sensical arguments that won't initialize the Sequence properly.
func (s *Sequence) Init(params ...uint64) error {

	// Ensure that the sequence is zeroed out.
	if s.initialized {
		return errors.New("cannot re-initialize a sequence object")
	}

	// If no parameters, create the default sequence.
	if len(params) == 0 {
		s.increment = 1
		s.minvalue = MinimumBound
		s.maxvalue = MaximumBound
	}

	// If a single parameter create a maximal bounding.
	if len(params) == 1 {

		// Ensure that the parameter is greater than the minimum value.
		if params[0] < MinimumBound {
			return errors.New("must specify a maximal value greater than 0")
		}

		s.increment = 1
		s.minvalue = MinimumBound
		s.maxvalue = params[0]
	}

	// If two parameters create a positive range.
	if len(params) == 2 {
		if params[1] < params[0] {
			return errors.New("for a positive increment, the maximum value must be greater than or equal to the minimum value")
		}

		if params[0] < MinimumBound || params[1] > MaximumBound {
			return errors.New("part of the range is out of bounds for positive increment")
		}

		s.increment = 1
		s.minvalue = params[0]
		s.maxvalue = params[1]
	}

	// If three parameters create a range with a new step.
	if len(params) == 3 {
		// The step cannot be zero
		if params[2] == 0 {
			return errors.New("must have a non-zero step to increment by")
		}

		if params[2] < 0 {
			// If the step is negative
			// TODO: This is not yet implemented since uints have to be positive.
			if params[0] < params[1] {
				return errors.New("for a negative increment, the first value must be greater than or equal to the second value")
			}

			if params[1] < MinimumBound || params[0] > MaximumBound {
				return errors.New("part of the range is out of bounds for negative increment")
			}
		} else {
			// If the step is positive
			if params[1] < params[0] {
				return errors.New("for a positive increment, the second value must be greater than or equal to the first value")
			}

			if params[0] < MinimumBound || params[1] > MaximumBound {
				return errors.New("part of the range is out of bounds for positive increment")
			}
		}

		s.increment = params[2]
		s.minvalue = params[0]
		s.maxvalue = params[1]

	}

	// If more than three parameters then return an error.
	if len(params) > 3 {
		return errors.New("too many arguments specified")
	}

	// Ensure unsigned subtraction won't lead to a problem.
	if int(s.minvalue)-int(s.increment) < 0 {
		return errors.New("the minimum value must be less than or equal to the step")
	}

	// Set current based on the minvalue and the increment.
	s.current = s.minvalue - s.increment

	// Set initialized to true and return
	s.initialized = true
	return nil
}

// Next updates the state of the Sequence and return the next item in the
// sequence. It will return an error if either the minimum or the maximal
// value has been reached.
func (s *Sequence) Next() (uint64, error) {
	s.current += s.increment

	// Check for missed minimum condition
	if s.current < s.minvalue {
		return 0, errors.New("reached minimum bound of the sequence")
	}

	// Check for reached maximum condition
	if s.current > s.maxvalue {
		return 0, errors.New("reached maximum bound of sequence")
	}

	return s.current, nil
}

// Restart the sequence by resetting the current value. This is the only
// method that allows direct manipulation of the sequence state which violates
// the monotonically increasing or decreasing rule. Use with care and as a
// fail safe if required.
func (s *Sequence) Restart() error {
	// Ensure that the sequence has been initialized.
	if !s.initialized {
		return errors.New("sequence has not been initialized")
	}

	// Ensure unsigned subtraction won't lead to a problem.
	if int(s.minvalue)-int(s.increment) < 0 {
		return errors.New("the minimum value must be less than or equal to the step")
	}

	// Set current based on the minvalue and the increment.
	s.current = s.minvalue - s.increment
	return nil
}

//===========================================================================
// Sequence State Methods
//===========================================================================

// Current value of the sequence. Normally the Next() method should be used to
// ensure only a single value is retrieved from the sequence. This method is
// used to compare the sequence state to another or to verify some external
// rule. Current will return an error if the sequence has not been started or
// initialized.
func (s *Sequence) Current() (uint64, error) {
	if !s.initialized {
		return 0, errors.New("sequence has not been initialized")
	}

	if !s.IsStarted() {
		return 0, errors.New("sequence has not been started")
	}

	return s.current, nil
}

// IsStarted returns the state of the Sequence (started or stopped). This
// method returns true if the current value is greater than or equal to the
// minimum value and if it is less than the maximal value. This method will
// also return false if the Sequence is not yet initialized.
func (s *Sequence) IsStarted() bool {
	if !s.initialized {
		return false
	}
	return !(s.current < s.minvalue) && s.current < s.maxvalue
}

// String returns a human readable representation of the sequence.
func (s *Sequence) String() string {
	d := fmt.Sprintf("incremented by %d between %d and %d", s.increment, s.minvalue, s.maxvalue)
	if !s.IsStarted() {
		return fmt.Sprintf("Unstarted Sequence %s", d)
	}
	return fmt.Sprintf("Sequence at %d, %s", s.current, d)
}

//===========================================================================
// Sequence Serialization Methods
//===========================================================================

// Dump the sequence into a JSON binary representation for the current state.
// The data that is dumped from this method can be loaded by an uninitialized
// Sequence to bring it as up to date as the sequence state when it was
// dumped. This method is intended to allow cross process communication of the
// sequence state.
//
// Note, however, that the autoincrement invariant is not satisfied during
// concurrent access. Therefore Dump and Load should be used with locks to
// ensure that the system does not end up diverging the state of the Sequence.
// It is up to the calling library to implement these locks.
func (s *Sequence) Dump() ([]byte, error) {
	if !s.IsStarted() {
		return nil, errors.New("cannot dump an uninitialized or unstarted sequence")
	}

	data := make(map[string]uint64)
	data["current"] = s.current
	data["increment"] = s.increment
	data["minvalue"] = s.minvalue
	data["maxvalue"] = s.maxvalue

	return json.Marshal(data)
}

// Load an uninitialized sequence from a JSON binary representation of the
// state of another sequence. The data should be exported from the sequence
// Dump method. If the data does not match the Sequence specification this
// method will return an error. Note that different versions of the sequence
// library could lead to errors.
func (s *Sequence) Load(data []byte) error {
	if s.initialized {
		return errors.New("cannot load into an initialized sequence")
	}

	vals := make(map[string]uint64)
	if err := json.Unmarshal(data, &vals); err != nil {
		return err
	}

	var ok bool

	if s.current, ok = vals["current"]; !ok {
		return errors.New("improperly formatted data or sequence version")
	}

	if s.increment, ok = vals["increment"]; !ok {
		return errors.New("improperly formatted data or sequence version")
	}

	if s.minvalue, ok = vals["minvalue"]; !ok {
		return errors.New("improperly formatted data or sequence version")
	}

	if s.maxvalue, ok = vals["maxvalue"]; !ok {
		return errors.New("improperly formatted data or sequence version")
	}

	s.initialized = true
	return nil
}
