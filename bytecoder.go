package skiptake

import (
	"encoding/binary"
)

// The details of the byte packing are isolated here, to allow for more complex
// schemes. The byte packing routines operate on a sequence of (skip, take)
// uint64 pairs.
//
// The byte packing routines are agnostic to the contents and should always
// decode the exact same sequence of pairs as were encoded. However, it does
// make assumptions about typical data to minimizing the packed size:
//
// - Take values of 0 are rare.
// - Skip values of 0 are rare, except as the first skip.
// - Both skip and take values are likely to be small.
// - Take values often repeat.
// - The most common take value in a large sequence is 1.
// - Shorter sequences have larger values, longer sequences have smaller values.

// -----------------------------------------------------------------------------

// This file employs the idea of 'varints' encoding, as used by Google's
// Protocol Buffers, and as implemented by the standard library encoding/binary
// package's "PutVarint()/GetVarint(), except augmented.
//
// The routines for putting and getting varints here actually encode up to (64 +
// split) bits. The extra bit(s) allow us to encode whether a value is a skip or
// a take. The extra bits are added to the least significant bits.
//
// Separate functions for putting/getting varints rather than bitwise shifts and
// operations are required so we can still encode and decode the full 64-bit
// values. This worst case scenario for a 64-bit int remains the same at 10
// bytes.

// Split is how many bits to add to the bottom of the varint. Split must be
// between 0 and 6.
const split = 1
const splitLowmask = (1 << split) - 1
const splitHighmask = 0x7f >> split

// readVarint2 Reads a varint from b starting at the offset pointed to by *i. *i
// is incremented as read. Returns the varint value as u, the extra split bits
// as e.
func readVarint2(b []byte, i *int) (u uint64, e int8) {
	var s uint
	if *i < len(b) {
		x := b[*i]
		*i++
		e = int8(x & splitLowmask)
		u = uint64((x & 0x7f) >> split)
		s += 7 - split
		if x >= 0x80 {
			for *i < len(b) {
				x = b[*i]
				*i++
				u |= uint64(x&0x7f) << s
				if x < 0x80 {
					return
				}
				s += 7
			}
		}
	}
	return
}

// Append a varint value u and split bits e to target. Behaves like append(),
// and returns the slice, if the slice was reallocated.
func appendVarint2(target []byte, u uint64, e int8) []byte {
	var ar [binary.MaxVarintLen64]byte
	i := 0
	x := byte(e&splitLowmask) | (byte(u&splitHighmask) << split)
	if u >= splitHighmask {
		ar[i] = x | 0x80
		u >>= (7 - split)
		i++
		for u > 0x80 {
			ar[i] = byte(u) | 0x80
			u >>= 7
			i++
		}
		ar[i] = byte(u)
	} else {
		ar[i] = x
	}
	return append(target, ar[:i+1]...)
}

const (
	skipFlag int8 = 0
	takeFlag      = 1
)

// Decoder abstracts reading pairs from the list.
//
// See skiptake.Iterator for a general-purpose iterator.
type Decoder struct {
	i        int
	Elements List
	lastTake uint64
}

// Next returns the next pair of skip, take values. Returns (0,0) in the case of
// end-of-sequence, although this sequence can occur within a list.
func (d *Decoder) Next() (skip, take uint64) {
	if !d.EOS() {
		n, e := readVarint2(d.Elements, &d.i)
		if e == skipFlag {
			skip = n
			if !d.EOS() {
				j := d.i
				n, e = readVarint2(d.Elements, &j)
				if e == takeFlag {
					d.lastTake = n
					d.i = j
				}
			}
			return skip + 1, d.lastTake + 1
		}
		if e == takeFlag {
			d.lastTake = n
			return 0, d.lastTake + 1
		}
	}
	return 0, 0
}

// PeekSkip returns the next skip values, without advancing the
// current decode location.
func (d *Decoder) PeekSkip() uint64 {
	j := d.i
	n, e := readVarint2(d.Elements, &j)
	if e != skipFlag {
		return 0
	}
	return n + 1
}

// EOS returns if the decoder is a that end of the sequence. Unlike
// Iterator.EOS(), this will return true before a call to Next().
func (d Decoder) EOS() bool {
	return d.i >= len(d.Elements)
}

// Reset resets the location of the decoder to the beginning of the sequence.
func (d *Decoder) Reset() {
	d.i = 0
}

// Encoder abstracts appending items to the list.
//
// See skiptake.Builder for a general-purpose list builder.
type Encoder struct {
	Elements *List
	lastTake uint64
}

// Add adds a new skip-take pair to the sequence.
func (e *Encoder) Add(skip, take uint64) {
	// Generally skips of zero should not occur in the middle of a list, so to
	// optimize our variable byte encoding, we instead encode skip-1. Skips of
	// size zero will still be encoded, but as the more expensive uint64(-1).
	//
	// However, skips of size zero are common at starting lists, indicating
	// that the first element should be part of the list. To optimise for this
	// common case, omit the skip and force-emit a take. When the decoder sees
	// a take when it expects a skip, it infers a skip of zero.
	//
	// This scheme of using a take when a skip was expected could be used to
	// encode all skips of zero. However, this requires that the previous take
	// was not omitted, to prevent ambiguity. Considering zero takes are the
	// exception, we don't do this, although the decoder would understand it.
	emitSkip := (skip != 0 || len(*e.Elements) > 0)
	skip--
	if emitSkip {
		*e.Elements = appendVarint2(*e.Elements, skip, skipFlag)
	}

	// Takes are only emitted if the new take value is different from the
	// previous take value. When the decoder sees a skip with no following take,
	// it infers to use the previous take value.
	//
	// Like skips, generally takes of zero should not ever occur, so again to
	// optimize  our variable byte encoding, we instead encode take - 1. Takes
	// of zero can still be encoded, but as the more expensive uint64(-1).
	//
	// We store the last take emitted as (take - 1). This allows for the
	// zero-state of both structures to correctly be a last take of one.
	take--
	if !emitSkip || take != e.lastTake {
		*e.Elements = appendVarint2(*e.Elements, take, takeFlag)
		e.lastTake = take
	}
}

// Flush instructs the encoder to write out any pending state.
func (e Encoder) Flush() {
	// No-op
}

// Decode returns a new skiptake.Decoder for the list.
// Note that the decoder should be passed by reference to maintain state.
func (l List) Decode() Decoder {
	return Decoder{Elements: l}
}

// Encode returns a new skiptake.Encoder for the list.
// Note that the encoder must be passed by reference to maintain state.
func (l *List) Encode() Encoder {
	return Encoder{Elements: l}
}

// Reset the list as a new empty list.
func (l *List) Reset() {
	if l != nil {
		*l = (*l)[:0]
	}
}
