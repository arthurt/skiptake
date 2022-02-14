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

type SkipTakeList []uint32

// SkipTakeDecoder holds the state for expanding an escaped 32-bit skip-take
// list as 64-bit values skip-take values
type SkipTakeDecoder struct {
	i        int
	Elements SkipTakeList
}

// SkipTakeIterator holds the state for iterating and seeking through the original input values
type SkipTakeIterator struct {
	Decoder SkipTakeDecoder
	skipSum uint64
	take    uint64
	n       uint64
}

// SkipTakeEncoder holds the state for turning an incrementing sequence into a skip-take sequence.
type SkipTakeEncoder struct {
	Elements *SkipTakeList
	n        uint64
	skip     uint32
	take     uint32
}

func Iterate(elements SkipTakeList) SkipTakeIterator {
	return SkipTakeIterator{
		Decoder: SkipTakeDecoder{
			Elements: elements,
		},
	}
}

func (x *SkipTakeDecoder) Next() (skip, take uint64) {
	if x.i+1 >= len(x.Elements) {
		if x.i < len(x.Elements) {
			x.i++
			return uint64(x.Elements[0]), 1
		}
		return 0, 0
	}

	skip = uint64(x.Elements[x.i])
	x.i++
	take = uint64(x.Elements[x.i])
	x.i++
	if take == 0 {
		if !(x.i+1 < len(x.Elements)) {
			// ERROR!
			return 0, 0
		}
		skip = skip | (uint64(x.Elements[x.i]) << 32)
		x.i++
		take = uint64(x.Elements[x.i])
		x.i++
	}

	// Try to keep eating so long as skip == 0
	for x.i+1 < len(x.Elements) && x.Elements[x.i] == 0 {
		x.i++
		take += uint64(x.Elements[x.i])
		x.i++
	}
	return
}

func (x *SkipTakeDecoder) EOS() bool {
	return x.i+1 >= len(x.Elements)
}

func (x *SkipTakeDecoder) Reset() {
	x.i = 0
}

func (t *SkipTakeIterator) Reset() {
	t.Decoder.Reset()
	t.skipSum = 0
	t.n = 0
}

func (x SkipTakeDecoder) Len() int {
	return x.Elements.Len()
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

// Skip add a skip value to the encoding sequence, and adds an implied take of 1
func (s *SkipTakeEncoder) Skip(skip uint64) {
	// Flush last (skip, take) pair if non-zero
	s.n += skip + 1
	if s.take > 0 {
		*s.Elements = append(*s.Elements, s.skip, s.take)
	}
	if skip > math.MaxUint32 {
		*s.Elements = append(*s.Elements, uint32(skip&math.MaxUint32), 0)
		skip = (skip >> 32) & math.MaxUint32
	}
	s.skip = uint32(skip)
	s.take = 1
}

// Take increments the latest take of the encoding sequence.
func (s *SkipTakeEncoder) IncTake() {
	s.take++
	s.n++
	if s.take == math.MaxUint32 {
		// Unlikely case, next use would overflow take. Write out take, restart
		// (skip, take) pair with a skip of 0.
		*s.Elements = append(*s.Elements, s.skip, s.take)
		s.skip = 0
		s.take = 0
	}
}

// AddTake incremenets the latest take of the encoding sequence by the amount
// passed.
func (s *SkipTakeEncoder) AddTake(take uint64) {
	take += uint64(s.take)
	s.n += take
	for take > math.MaxUint32 {
		*s.Elements = append(*s.Elements, s.skip, math.MaxUint32)
		s.skip = 0
		take -= math.MaxUint32
	}
	s.take = uint32(take)
}

// Next accepts the next sequence value. Add either a skip or take as required
// depending on if the value immediately follows the previous, or if there is
// a gap. Calls to n must be strictly-increasing. If not, Next returns false.
func (s *SkipTakeEncoder) Next(n uint64) bool {
	i := n - s.n
	if i < 0 {
		return false
	}
	if i == 0 {
		s.IncTake()
	}
	if i > 0 {
		s.Skip(i)
	}
	return true
}

//Flush finishes the pending skip-take pair to the list.
func (s *SkipTakeEncoder) Flush() {
	if s.take > 0 {
		// Corner case: One index, only write one skip number
		if s.take == 1 && len(*s.Elements) == 0 {
			*s.Elements = append(*s.Elements, s.skip)
		} else {
			*s.Elements = append(*s.Elements, s.skip, s.take)
		}
	}
}

// Len() returns how many items are in the expanded sequence
func (s SkipTakeList) Len() int {
	if s == nil {
		return 0
	}
	if len(s) == 1 {
		return 1
	}
	capacity := 0
	for i := 1; i < len(s); i += 2 {
		capacity += int(s[i])
	}
	return capacity
}

// Expand expands the sequence as a slice of uint64 values
func (s SkipTakeList) Expand() []uint64 {
	output := make([]uint64, s.Len())
	iter := Iterate(s)

	for i := 0; i < len(output); i++ {
		output[i] = iter.Next()
	}
	return output
}

// Create takes a sequence of uint64s and creates a skip-take list, or nil if
// the sequence is non in sorted order.
func Create(sequence []uint64) SkipTakeList {
	list := SkipTakeList{}
	ste := SkipTakeEncoder{Elements: &list}
	for _, n := range sequence {
		if !ste.Next(n) {
			return nil
		}
	}
	ste.Flush()
	return list
}

// IsZeroOrSingle returns if the encoded sequence is zero or 1 in length.
func (s SkipTakeList) IsZeroOrSingle() bool {
	// Length == 0 - zero item sequence
	// Length == 1 - one item sequence
	if len(s) < 2 {
		return true
	}
	// Length == 2, take == 1 - one item sequence
	if len(s) == 2 && s[1] == 1 {
		return true
	}

	// Length == 4, take1 == 0, take2 == 1 - one item sequence (escaped large skip)
	if len(s) == 4 && s[1] == 0 && s[3] == 1 {
		return true
	}
	return false
}

func (s SkipTakeList) String() string {
	n := uint64(0)
	b := strings.Builder{}
	dec := SkipTakeDecoder{Elements: s}

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
