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

// Next returns the next pair of skip, take values. Returns (0,0) in the case of
// end-of-sequence, although this sequence can occur within a list.
func (x *Decoder) Next() (skip, take uint64) {
	skip = readVarint(x.Elements, &x.i)
	take = readVarint(x.Elements, &x.i)
	return
}

// PeekSkip returns the next skip values, without advancing the
// current decode location.
func (x *Decoder) PeekSkip() (skip uint64) {
	j := x.i
	return readVarint(x.Elements, &j)
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
