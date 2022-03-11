package skiptake

import (
	"testing"
)

func testEncodeDecode(t *testing.T, values [][2]uint64) {
	var l List

	enc := l.Encode()
	for _, pair := range values {
		enc.Add(pair[0], pair[1])
		t.Logf("Add skip: %d, take: %d", pair[0], pair[1])
	}
	enc.Flush()
	t.Logf("Encoded as %d bytes: %v", len([]byte(l)), []byte(l))

	dec := l.Decode()
	for _, pair := range values {
		skip, take := dec.Next()
		if skip != pair[0] || take != pair[1] {
			t.Fatalf("Decoded list differs from encoded values. (%d != %d) || (%d != %d)", skip, pair[0], take, pair[1])
		}
	}
	if !dec.EOS() {
		skip, take := dec.Next()
		t.Fatalf("Decoder has more symbols than encoded. Read %d, %d", skip, take)
	}
}

func Test_EncodeDecode(t *testing.T) {

	// The empty list
	testEncodeDecode(t, [][2]uint64{})

	// List which matches default encoder state
	testEncodeDecode(t, [][2]uint64{[2]uint64{0, 1}})

	// Zero-skip start
	testEncodeDecode(t, [][2]uint64{[2]uint64{0, 5000000000}})

	// Non-zero skip start
	testEncodeDecode(t, [][2]uint64{[2]uint64{5, 100}})

	// Common-case: Single offset list
	testEncodeDecode(t, [][2]uint64{[2]uint64{30, 1}})

	// Last take same as previous and not one.
	testEncodeDecode(t, [][2]uint64{[2]uint64{0, 2}, [2]uint64{2, 2}})

	// Average-ish case
	testEncodeDecode(t, [][2]uint64{
		[2]uint64{0, 1},
		[2]uint64{1, 1},
		[2]uint64{1, 1},
		[2]uint64{1, 1},
		[2]uint64{83, 1},
		[2]uint64{3, 4},
		[2]uint64{100, 1},
		[2]uint64{32, 2},
	})

	// Large and 64-bit values
	testEncodeDecode(t, [][2]uint64{
		[2]uint64{0x100000000, 0x200000000},
		[2]uint64{0x400000000000, 0x2000000000000},
		[2]uint64{0x8000000000000000, 0x8000000000000000},
	})

	// Bad data - mid-stream 0 skip
	testEncodeDecode(t, [][2]uint64{
		[2]uint64{9, 1},
		[2]uint64{0, 1},
		[2]uint64{1, 1},
	})

	// Bad data - mid-stream 0 take
	testEncodeDecode(t, [][2]uint64{
		[2]uint64{9, 1},
		[2]uint64{3, 0},
		[2]uint64{1, 1},
	})

	// Bad data - begining of stream 0 take
	testEncodeDecode(t, [][2]uint64{
		[2]uint64{0, 0},
		[2]uint64{3, 1},
		[2]uint64{1, 1},
	})

	// Bad data - mid stream 0 skip and 0 take
	testEncodeDecode(t, [][2]uint64{
		[2]uint64{0, 4},
		[2]uint64{0, 0},
		[2]uint64{50, 50},
	})
}
