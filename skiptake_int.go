package skiptake

import (
	"math"
)

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
