package skiptake

import (
	"encoding/binary"
)

// SkipTakeList is a storage implementation of a skip-take list that reads and
// writes from a series of variable-width unsigned integers packed into a byte
// array. This is similar to the scheme employed by Google's Protocol Buffers.
//
// As most values needing to be stored are small, this encoding saves
// considerable space.
//
// The details of the packing are isolated here, to allow for future schemes to
// deal with repetative patterns.

// The concrete type of the list storage
type SkipTakeList []byte

// A class to abstract reading pairs from the list. See SkipTakeIterator for a
// general-purpose iterator.
type SkipTakeDecoder struct {
	i        int
	Elements SkipTakeList
}

// A class to abstract appending items to the list
type SkipTakeEncoder struct {
	Elements *SkipTakeList
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
func (x *SkipTakeDecoder) Next() (skip, take uint64) {
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

func (x *SkipTakeDecoder) EOS() bool {
	return x.i >= len(x.Elements)
}

// Reset resets the location of the decoder to the beginning of the sequence.
func (x *SkipTakeDecoder) Reset() {
	x.i = 0
}

// Add adds a new skip-take pair to the sequence.
func (s SkipTakeEncoder) Add(skip, take uint64) {
	*s.Elements = appendVarint(*s.Elements, skip)
	*s.Elements = appendVarint(*s.Elements, take)
}

func (v SkipTakeList) Decode() SkipTakeDecoder {
	return SkipTakeDecoder{Elements: v}
}

func (v *SkipTakeList) Encode() SkipTakeEncoder {
	return SkipTakeEncoder{Elements: v}
}

func (v *SkipTakeList) Clear() {
	if v != nil {
		*v = (*v)[:0]
	}
}
