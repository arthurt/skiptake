package skiptake

import "math"

// Iterator holds the state for iterating and seeking through the
// original input values.
type Iterator struct {
	Decoder *Decoder
	skipSum uint64
	take    uint64
	n       uint64
}

// Reset resets the iterator to be beginning of the list.
func (t *Iterator) Reset() {
	t.Decoder.Reset()
	t.skipSum = 0
	t.n = 0
}

// EOS returns true if the stream is at end-of-stream. Because Iterator can
// iterate by either individual sequence values or intervals, EOS will not
// return true until AFTER a call of either Next() or NextSkipTake() causes EOS
// to be reached. Use of EOS should therefor occur after calls to Next() or
// NextSkipTake(). Eg:
//
//		for skip, take := iter.Next(); !iter.EOS(); skip, take = iter.Next() {
//			...
//		}
//
func (t Iterator) EOS() bool {
	return t.n == math.MaxUint64
}

// NextSkipTake advances the iterator to the next skip-take interval, returning
// the skip and take values. A following call to Next() returns the expanded
// sequence value at the beginning of the interval. Returns (0, 0)
// if-and-only-if the iterator is at the end-of-sequence.
//
// NextSkipTake will coalesce any zero skip or zero take values that occur
// within the underlying list. Callers are safe to assume that while not at the
// end-of-sequence, all returned intervals will have a discontinuity (non-zero
// skip), and have a greater than zero size (non-zero take.)
//
// This means that Iterator.NextSkipTake() may not always return the same values
// as would be returned by Decoder.Next(), which can return zero skip and take
// values if they are present in the source list.
func (t *Iterator) NextSkipTake() (skip, take uint64) {
	// Coalesce zero takes
	for take == 0 {
		if t.Decoder.EOS() {
			t.n = math.MaxUint64
			return 0, 0
		}
		nskip, ntake := t.Decoder.Next()
		skip += nskip
		take = ntake
	}
	// Coalesce zero skips
	for !t.Decoder.EOS() {
		nskip := t.Decoder.PeekSkip()
		if nskip != 0 {
			break
		}
		_, ntake := t.Decoder.Next()
		take += ntake
	}
	t.skipSum += skip
	t.n += t.take + skip
	t.take = take
	return
}

// Next returns the next value in the sequence. Returns max uint64 as
// end-of-sequence. If following a call to NextSkipTake() or Seek(), returns
// the first sequence value of the new take interval.
func (t *Iterator) Next() uint64 {
	if t.take == 0 {
		t.NextSkipTake()
		if t.EOS() {
			return math.MaxUint64
		}
	}
	result := t.n
	t.take--
	t.n++
	return result
}

// Seek seeks to the i'th value in the expanded sequence. Returns the i'th
// expanded sequence value as skip, and how many sequential elements until the
// next skip as take. These values are identical to what would be the first
// skip-take pair of an expanded sequence that was truncated to start at the
// i'th position.
//
// A following call to Next() start returns the same i'th sequence value, and
// subsequent calls to Next() continue as normal within the expanded sequence.
//
// A following call to NextSkipTake() returns next skip take pair, not the skip
// take pair that was seeked to.
func (t *Iterator) Seek(pos uint64) (uint64, uint64) {
	takeSum := t.n + t.take - t.skipSum
	if pos < takeSum {
		t.Reset()
		takeSum = 0
	}
	for takeSum <= pos {
		if t.Decoder.EOS() {
			t.n = math.MaxUint64
			return 0, 0
		}
		_, take := t.NextSkipTake()
		takeSum += take
	}
	t.take = takeSum - pos
	t.n = t.skipSum + pos
	return t.n, t.take
}
