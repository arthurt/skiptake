package skiptake

import (
	"math"
)

// This is an implementation of a skip-take list using a []uint32 array.
//
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

type SkipTakeIntList []uint32

// SkipTakeIntDecoder holds the state for expanding an escaped 32-bit skip-take
// list as 64-bit values skip-take values
type SkipTakeIntDecoder struct {
	i        int
	Elements SkipTakeIntList
}

// SkipTakeIntEncoder holds the state for turning an incrementing sequence into a skip-take sequence.
type SkipTakeIntEncoder struct {
	Elements *SkipTakeIntList
}

var _ SkipTakeList = &SkipTakeIntList{}
var _ SkipTakeWriter = &SkipTakeIntList{}

func (x *SkipTakeIntDecoder) EOS() bool {
	return x.i+1 >= len(x.Elements)
}

func (x *SkipTakeIntDecoder) Reset() {
	x.i = 0
}

func (x *SkipTakeIntDecoder) Next() (skip, take uint64) {
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

func (s SkipTakeIntEncoder) Add(skip, take uint64) {
	if skip > math.MaxUint32 {
		*s.Elements = append(*s.Elements, uint32(skip&math.MaxUint32), 0)
		skip = (skip >> 32) & math.MaxUint32
	}
	for take > math.MaxUint32 {
		*s.Elements = append(*s.Elements, uint32(skip), math.MaxUint32)
		skip = 0
		take -= math.MaxUint32
	}
	*s.Elements = append(*s.Elements, uint32(skip), uint32(take))
}

func (s SkipTakeIntEncoder) Finish() SkipTakeList {
	return s.Elements
}

func (l SkipTakeIntList) Decode() SkipTakeDecoder {
	return &SkipTakeIntDecoder{Elements: l}
}

func (l *SkipTakeIntList) Encode() SkipTakeEncoder {
	return SkipTakeIntEncoder{Elements: l}
}

func (l *SkipTakeIntList) Clear() {
	if l != nil {
		*l = (*l)[:0]
	}
}

func (l SkipTakeIntList) String() string {
	return ToString(l)
}

func (l SkipTakeIntList) Expand() []uint64 {
	return Expand(l)
}

// Len() returns how many items are in the expanded sequence
func (s SkipTakeIntList) Len() int {
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
