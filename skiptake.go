package skiptake

import (
	"fmt"
	"math"
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

// SkipTakeDecoder is an interface that can produce SkipTake Values,
// maintaining a current location state.
type SkipTakeDecoder interface {
	// Next returns the next pair of skip, take values. Returns (0,0) as a
	// special case of End-of-Sequence.
	Next() (skip, take uint64)

	// EOS returns true when the end of the sequence has been reached, false
	// otherwise.
	EOS() bool

	// Reset resets the location of the decoder to the beginning of the
	// sequence.
	Reset()
}

// SkipTakeEncoder is an interface that can store a skip take sequence.
type SkipTakeEncoder interface {
	// Add adds a new skip, take pair to the sequence.
	Add(skip, take uint64)

	// Finish completes the sequence and returns the corresponding readable
	// result.
	Finish() SkipTakeList
}

// SkipTakeList is an interface for reading a sequence of skip take pairs.
type SkipTakeList interface {
	Decode() SkipTakeDecoder
}

// SkipTakeWriter is an interface for writing a sequence of skip take pairs.
type SkipTakeWriter interface {
	Encode() SkipTakeEncoder
	Clear()
}

func Iterate(l SkipTakeList) SkipTakeIterator {
	return SkipTakeIterator{Decoder: l.Decode()}
}

func Build(l SkipTakeWriter) SkipTakeBuilder {
	l.Clear()
	return SkipTakeBuilder{Encoder: l.Encode()}
}

func Create(out SkipTakeWriter, values []uint64) SkipTakeList {
	b := Build(out)
	for _, v := range values {
		if !b.Next(v) {
			return nil
		}
	}
	b.Flush()
	return b.Encoder.Finish()
}

// Encode a skip take list from a list of (skip, take) []uint64 pairs
func FromRaw(out SkipTakeWriter, v []uint64) {
	e := out.Encode()
	for i := 0; i < len(v); i++ {
		skip := v[i]
		i++
		if !(i < len(v)) {
			break
		}
		take := v[i]
		e.Add(skip, take)
	}
}

func Len(l SkipTakeList) uint64 {
	var ret uint64
	for d := l.Decode(); !d.EOS(); {
		_, t := d.Next()
		ret += t
	}
	return ret
}

// SkipTakeIterator holds the state for iterating and seeking through the original input values
type SkipTakeIterator struct {
	Decoder SkipTakeDecoder
	skipSum uint64
	take    uint64
	n       uint64
}

// Builds a skip take sequence
type SkipTakeBuilder struct {
	Encoder SkipTakeEncoder
	n       uint64
	skip    uint64
	take    uint64
}

func (b *SkipTakeBuilder) Skip(skip uint64) {
	b.n += skip + 1
	if skip == 0 {
		b.take++
	} else {
		b.Flush()
		b.skip = skip
		b.take = 1
	}
}

func (b *SkipTakeBuilder) IncTake() {
	b.take++
	b.n++
}

func (b *SkipTakeBuilder) AddTake(take uint64) {
	b.take += take
	b.n++
}

func (b *SkipTakeBuilder) Next(n uint64) bool {
	if n < b.n {
		return false
	}
	b.Skip(n - b.n)
	return true
}

func (b *SkipTakeBuilder) Flush() {
	if b.take > 0 {
		b.Encoder.Add(b.skip, b.take)
	}
}

func (t *SkipTakeIterator) Reset() {
	t.Decoder.Reset()
	t.skipSum = 0
	t.n = 0
}

func (t SkipTakeIterator) EOS() bool {
	return t.n == math.MaxUint64
}

//NextSkipTake advances the iterator to the beginning of the next Skip-Take
//range, returning the skip and take values. A following call to Next() returns
//the sequence value of the beginning of the interval.
func (t *SkipTakeIterator) NextSkipTake() (skip, take uint64) {
	skip, take = t.Decoder.Next()
	if take == 0 {
		t.n = math.MaxUint64
		return
	}
	t.skipSum += skip
	t.n += t.take + skip
	t.take = take
	return
}

//Next returns the next value in the sequence. Returns max uint64 as
//end-of-sequence. If following a call to NextSkipTake() or Seek(), returns
//to sequence value of the new take interval.
func (t *SkipTakeIterator) Next() uint64 {
	if t.take == 0 {
		skip, take := t.Decoder.Next()
		if take == 0 {
			t.n = math.MaxUint64
			return math.MaxUint64
		}
		t.skipSum += skip
		t.n += skip
		t.take = take
	}
	result := t.n
	t.take--
	t.n++
	return result
}

// Seek seeks to the i'th value in the expanded sequence. Returns the i'th
// seqence value as skip, and how many sequential elements until the next skip
// as take. A following call to Next() returns the same i'th sequence value.
func (t *SkipTakeIterator) Seek(pos uint64) (uint64, uint64) {
	takeSum := t.n + t.take - t.skipSum
	if pos < takeSum {
		t.Reset()
		takeSum = 0
	}
	for takeSum <= pos {
		_, take := t.NextSkipTake()
		if take == 0 {
			return 0, 0
		}
		takeSum += take
	}
	t.take = takeSum - pos
	t.n = t.skipSum + pos
	return t.n, t.take
}

// Expand expands the sequence as a slice of uint64 values
func Expand(s SkipTakeList) []uint64 {
	var output []uint64
	iter := Iterate(s)
	for n := iter.Next(); !iter.EOS(); n = iter.Next() {
		output = append(output, n)
	}
	return output
}

func ToString(s SkipTakeList) string {
	n := uint64(0)
	b := strings.Builder{}
	dec := s.Decode()

	for {
		skip, take := dec.Next()
		if take == 0 {
			break
		}
		if n > 0 {
			b.WriteString(", ")
		}
		n += skip
		if take == 1 {
			fmt.Fprintf(&b, "%d", n)
		} else {
			fmt.Fprintf(&b, "[%d - %d]", n, n+take-1)
		}
		n += take
	}
	return b.String()
}
