package skiptake

import (
	"fmt"
	"strings"
)

// SkipTakeList is a datatype for storing a sequence of strictly increasing
// integers, most efficient for sequences which have many contiguous
// sub-sequences with contiguous gaps.
//
// The basic idea is an interleaved list of skip and take instructions on how to
// build the sequence from preforming 'skip' and 'take' on the sequence of all
// integers.
//
// Eg: Skip 1, take 4, skip 2, take 3 would create the sequence: (1, 2, 3, 4,
// 7, 8, 9), and is effectively (Skip [0]), (Take [1-4]), (Skip [5-6]), (Take
// [8-9])
//
// As most skip and take values are not actual output values, but always
// positive differences between such values, the integer type used to store the
// differences can be of a smaller bitwidth to conserve memory.

// Create() creates a skip-take list from the passed slice of values. These
// values should be a strictly increasing sequence.
func Create(values []uint64) SkipTakeList {
	l := SkipTakeList{}
	b := Build(&l)
	for _, v := range values {
		if !b.Next(v) {
			return nil
		}
	}
	b.Flush()
	return l
}

// FromRaw() creates a skip-take list from a slice of []uint64 values
// representing a sequence of alternating skip and take values.
func FromRaw(v []uint64) SkipTakeList {
	l := SkipTakeList{}
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

// GetRaw() returns a []uint64 sequence of alternating skip and take values.
func (l SkipTakeList) GetRaw() []uint64 {
	result := []uint64{}
	for d := l.Decode(); !d.EOS(); {
		s, t := d.Next()
		result = append(result, s, t)
	}
	return result
}

// Len() returns how many values are in the expanded original sequence.
func (l SkipTakeList) Len() uint64 {
	var ret uint64
	for d := l.Decode(); !d.EOS(); {
		_, t := d.Next()
		ret += t
	}
	return ret
}

// Iterate() returns a SkipTakeIterator for the passed list.
func (l SkipTakeList) Iterate() SkipTakeIterator {
	d := l.Decode()
	return SkipTakeIterator{Decoder: &d}
}

// Expand() expands the sequence as a slice of uint64 values of length Len().
func (l SkipTakeList) Expand() []uint64 {
	var output []uint64

	iter := l.Iterate()
	for n := iter.Next(); !iter.EOS(); n = iter.Next() {
		output = append(output, n)
	}
	return output
}

// Implement Stringer interface
func (l SkipTakeList) String() string {
	return l.Format(120)
}

// Print the skip-take list as ranges. If maxlength < 0 will print all ranges.
// Otherwise,maxlength sets the upper bound for the returned string length, and
// the ranges will be truncated, with '...' appended.
func (l SkipTakeList) Format(maxLen int) string {
	n := uint64(0)
	b := strings.Builder{}
	dec := l.Decode()
	first := true
	for {
		skip, take := dec.Next()
		if take == 0 {
			break
		}
		n += skip
		var s string
		if take == 1 {
			s = fmt.Sprintf("%d", n)
		} else {
			s = fmt.Sprintf("[%d - %d]", n, n+take-1)
		}
		n += take
		needed_cap := len(s)
		if !first {
			needed_cap += 2
		}
		if !dec.EOS() {
			needed_cap += 3
		}

		if maxLen < 0 || maxLen >= needed_cap {
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
