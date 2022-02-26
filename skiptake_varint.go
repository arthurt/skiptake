package skiptake

import (
	"encoding/binary"
)

// The details of the byte packing are isolated here, to allow for future
// schemes to deal with repetative patterns.

// Decoder abstracts reading pairs from the list.
//
// See skiptake.Iterator for a general-purpose iterator.
type Decoder struct {
	i        int
	Elements List
}

// Encoder abstracts appending items to the list.
//
// See skiptake.Builder for a general-purpose list builder.
type Encoder struct {
	Elements *List
}

// Read a varint from the byte slice
func readVarint(s []byte, i *int) uint64 {
	v, n := binary.Uvarint(s[*i:])
	if n <= 0 {
		return 0
	}
	*i += n
	return v
}

// Append a VarInt to the byte slice.
func appendVarint(target []byte, v uint64) []byte {
	var ar [10]byte
	i := binary.PutUvarint(ar[:], v)
	return append(target, ar[:i]...)
}

// Next returns the next pair of skip, take values. Returns (0,0) as a special
// case of End-of-Sequence.
func (x *Decoder) Next() (skip, take uint64) {
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

// EOS returns if the decoder is a that end of the sequence.
func (x *Decoder) EOS() bool {
	return x.i >= len(x.Elements)
}

// Reset resets the location of the decoder to the beginning of the sequence.
func (x *Decoder) Reset() {
	x.i = 0
}

// Add adds a new skip-take pair to the sequence.
func (s Encoder) Add(skip, take uint64) {
	*s.Elements = appendVarint(*s.Elements, skip)
	*s.Elements = appendVarint(*s.Elements, take)
}

// Decode returns a new skiptake.Decoder for the list.
func (v List) Decode() Decoder {
	return Decoder{Elements: v}
}

// Encode returns a new skiptake.Encoder for the list
func (v *List) Encode() Encoder {
	return Encoder{Elements: v}
}

// Clear resets the list as a new empty list.
func (v *List) Clear() {
	if v != nil {
		*v = (*v)[:0]
	}
}
