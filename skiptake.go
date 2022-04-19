//License: foo

/*
Package skiptake is an implementation of a low-complexity integer sequence
compression.

A Skip-Take List is a datatype for storing a sequence of strictly increasing
integers, most efficient for sequences which have many contiguous sub-sequences
with contiguous gaps.

The basic idea is an interleaved list of skip and take instructions on how to
build the sequence from preforming 'skip' and 'take' on the sequence of all
integers.

Eg:

	Skip 1, take 4, skip 2, take 3

would create the sequence:

	(1, 2, 3, 4, 7, 8, 9),

and is effectively

	Skip: 0         5 6
	Take:   1 2 3 4     7 8 9

As most skip and take values are not actual output values, but always
positive differences between such values, the integer type used to store the
differences can be of a smaller bitwidth to conserve memory.

In this package, the operation 'skip' always implies a take of at lest 1
following the skip, but the operation take does not imply a non-zero take before
it.
*/
package skiptake

import (
	"fmt"
	"strings"
)

// List is the type that holds a Skip-Take list.
//
// A Skip-Take list is encoded as a series of variable width integers, with
// additional run length encoding.
type List []byte

// Create creates a skip-take list from the passed slice of values. These
// values should be a strictly increasing sequence. Returns nil if not.
func Create(values ...uint64) List {
	b := Build(&List{})
	for _, v := range values {
		if !b.Next(v) {
			return nil
		}
	}
	return b.Finish()
}

// Equal returns true if two lists are the same. That is, they contain the same
// subsequence.
func Equal(a, b List) bool {
	ad := a.Decode()
	bd := b.Decode()

	for !ad.EOS() {
		as, at := ad.Next()
		bs, bt := bd.Next()
		if as != bs || at != bt {
			return false
		}
	}

	if !bd.EOS() {
		return false
	}
	return true
}

// FromRaw creates a skip-take list from a slice of []uint64 values
// representing a sequence of alternating skip and take values.
func FromRaw(v []uint64) List {
	l := List{}
	e := l.Encode()
	for i := 0; i < len(v); i++ {
		skip := v[i]
		i++
		if !(i < len(v)) {
			break
		}
		take := v[i]
		e.Add(skip, take)
	}
	return l
}

// GetRaw returns a []uint64 sequence of alternating skip and take values.
func (l List) GetRaw() []uint64 {
	result := []uint64{}
	for d := l.Decode(); !d.EOS(); {
		s, t := d.Next()
		result = append(result, s, t)
	}
	return result
}

// Len returns how many values are in the expanded sequence.
func (l List) Len() uint64 {
	var ret uint64
	for d := l.Decode(); !d.EOS(); {
		_, t := d.Next()
		ret += t
	}
	return ret
}

// Iterate returns a new skiptake.Iterator for the passed list.
func (l List) Iterate() Iterator {
	d := l.Decode()
	return Iterator{Decoder: &d}
}

// Expand expands the sequence as a slice of uint64 values of length Len().
//
// Note: Caution is to be exercised. A naive use of Expand() on a result of
// skiptake.Complement() can create an array of a multiple of 2^64, which is
// ~147 EB (147,000,000 GB) in size.
func (l List) Expand() []uint64 {
	// Fail-fast
	output := make([]uint64, 0, l.Len())

	iter := l.Iterate()
	for n := iter.Next(); !iter.EOS(); n = iter.Next() {
		output = append(output, n)
	}
	return output
}

// Implement the fmt.Stringer interface. Returns
//		l.Format(120).
func (l List) String() string {
	return l.Format(120)
}

// Format returns a human-friendly representation of the list of values as
// a sequence of ranges, or individual members for ranges of length 1.
//
// If the argument maxLen > 0, it specifies the maximum length of the string to
// return. List which exceed this limit will be truncated with `...`. Otherwise
// all ranges will be included.
func (l List) Format(maxLen int) string {
	b := strings.Builder{}

	first := true
	iter := l.Iterate()
	for iter.NextSkipTake(); !iter.EOS(); iter.NextSkipTake() {
		var s string
		begin, end := iter.Interval()
		if end <= begin {
			s = fmt.Sprintf("%d", begin)
		} else {
			s = fmt.Sprintf("[%d - %d]", begin, end)
		}
		neededCap := len(s)
		if !first {
			neededCap += 2
		}
		if !iter.EOS() {
			neededCap += 3
		}

		if maxLen < 0 || maxLen >= neededCap {
			if !first {
				b.WriteString(", ")
			}
			b.WriteString(s)
			if maxLen >= 0 {
				maxLen -= len(s) + 2
			}
		} else {
			if maxLen >= 3 {
				b.WriteString("...")
			}
			break
		}
		first = false
	}
	return b.String()
}
