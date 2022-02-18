package skiptake

// SkipTakeListVarInt is a implementation of SkipTakeList/SkipTakeWriter that
// reads and writes from a series of variable-width unsigned integers packed
// into a byte array. This is similar to the scheme employed by Google's
// Protocol Buffers.
//
// As most values needing to be stored are small, this encoding saves
// considerable space.

type SkipTakeListVarInt []byte

type SkipTakeVarIntDecoder struct {
	i        int
	Elements SkipTakeListVarInt
}

type SkipTakeVarIntEncoder struct {
	Elements *SkipTakeListVarInt
}

// Assert that we implement the interfaces.
var _ SkipTakeList = SkipTakeListVarInt{}
var _ SkipTakeWriter = &SkipTakeListVarInt{}
var _ SkipTakeDecoder = &SkipTakeVarIntDecoder{}
var _ SkipTakeEncoder = SkipTakeVarIntEncoder{}

// Read a VarInt from the byte slice. Values are packed as 7-bits per byte,
// unused high bit used as a flag that more bytes follow.
func readVarint(s []byte, i *int) uint64 {
	var ret uint64
	for n := 0; *i < len(s); n += 7 {
		b := s[*i]
		*i++
		ret = ret | (uint64(b&0x7f) << n)
		if b&0x80 == 0 {
			break
		}
	}
	return ret
}

// Append a VarInt to the byte slice.
func appendVarint(target []byte, v uint64) []byte {
	var ar [10]byte
	i := 0
	for {
		b := byte(v & 0x7f)
		v = v >> 7
		if v != 0 {
			ar[i] = b | 0x80
			i++
		} else {
			ar[i] = b
			i++
			break
		}
	}
	return append(target, ar[:i]...)
}

func (x *SkipTakeVarIntDecoder) Next() (skip, take uint64) {
	skip = readVarint(x.Elements, &x.i)
	take = readVarint(x.Elements, &x.i)

	// Try to keep eating so long as skip == 0
	j := x.i
	for j < len(x.Elements) {
		nskip := readVarint(x.Elements, &j)
		if nskip != 0 {
			break
		}
		take += readVarint(x.Elements, &j)
		x.i = j
	}
	return
}

func (x *SkipTakeVarIntDecoder) EOS() bool {
	return x.i+1 > len(x.Elements)
}

func (x *SkipTakeVarIntDecoder) Reset() {
	x.i = 0
}

func (s SkipTakeVarIntEncoder) Add(skip, take uint64) {
	*s.Elements = appendVarint(*s.Elements, skip)
	*s.Elements = appendVarint(*s.Elements, take)
}

func (s SkipTakeVarIntEncoder) Finish() SkipTakeList {
	return *s.Elements
}

func (v SkipTakeListVarInt) Decode() SkipTakeDecoder {
	return &SkipTakeVarIntDecoder{Elements: v}
}

func (v *SkipTakeListVarInt) Encode() SkipTakeEncoder {
	return &SkipTakeVarIntEncoder{Elements: v}
}

func (v *SkipTakeListVarInt) Clear() {
	if v != nil {
		*v = (*v)[:0]
	}
}

// Implement Stringer interface
func (l SkipTakeListVarInt) String() string {
	return ToString(l)
}

func (l SkipTakeListVarInt) Expand() []uint64 {
	return Expand(l)
}
