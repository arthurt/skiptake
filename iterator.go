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

//Reset resets the iterator to be beginning of the list.
func (t *Iterator) Reset() {
	t.Decoder.Reset()
	t.skipSum = 0
	t.n = 0
}

//EOS returns true if the stream is at end-of-stream. Because Iterator
//can iterate by either individual numbers or whole ranges, EOS will not return
//true until AFTER a call to either Next() or NextSkipTake() causes EOS to be
//reached.
func (t Iterator) EOS() bool {
	return t.n == math.MaxUint64
}

//NextSkipTake advances the iterator to the beginning of the next Skip-Take
//range, returning the skip and take values. A following call to Next() returns
//the sequence value of the beginning of the interval.
func (t *Iterator) NextSkipTake() (skip, take uint64) {
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
func (t *Iterator) Next() uint64 {
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
func (t *Iterator) Seek(pos uint64) (uint64, uint64) {
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
