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
//
// In this implementation, 32-bit unsigned integers are used as a compromise
// to store both the skip and take values.
//
// However, occasionally a value which overflows a 32-bit integer is needed.
// Weighing the idea that large skips are more likely than large takes, the
// following escaping scheme is used.
//
// For a take larger than 2^32, the scheme exploits the fact that a skip of size
// 0 is (and must be) a legitimate value. If the take value of a skiptake
// element were to overflow, rather than incrementing it, a new skiptake element
// is appended with a skip of zero. This scheme does however require
// ceil(n/2^32) or O(n) skiptake elements for a take of (n).
//
// For a skip larger than 2^32, an escape is used to encode a full 64-bit skip
// value. The escape is signaled by setting the take value of the first
// skiptake element to 0. The 64-bit skip value is then divided into two 32-bit
// parts, the low 32 bits stored in the first skip-take element, the high 32-
// bits stored in the skip of the second skip-take element.
//
//  Skiptake pair 1 | int 1:	skip: [skip bits 0-31]
//					| int 2:	take: 0 (flag value)
//  Skiptake pair 2	| int 3:	skip: [skip bits 32-63]
//					| int 4:	take: <normal take value>
//
// The take value for the second element of the escape is the take value that
// would have occurred in the skiptake element, had the skip value not
// overflowed the skip value. This scheme is O(1) for 64-bit skip values.
//
// A skip take list should always have an even length, consisting of skip then
// take values. However, to optimize for the case of a sequence containing only
// one number, as an exception a list consisting of one take at a skip value
// that fits within a uin32 may have a length of one consiting of a single skip
// value, with an implied take of one.

type SkipTakeDecoder interface {
	Next() (skip, take uint64)
	EOS() bool
	Reset()
}

type SkipTakeEncoder interface {
	Add(skip, take uint64)
	Finish() SkipTakeList
}

type SkipTakeList interface {
	Decode() SkipTakeDecoder
}

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

// =========================================

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
