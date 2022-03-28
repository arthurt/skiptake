package skiptake

import (
	"testing"
)

func equalUint64(a, b []uint64) bool {
	if len(a) != len(b) {
		return false
	}

	for i := 0; i < len(a); i++ {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

type intrv [2]uint

func makeRange(args []intrv) List {
	b := Build(&List{})
	for _, p := range args {
		b.Next(uint64(p[0]))
		b.Take(uint64(p[1] - p[0]))
	}
	return b.Finish()
}

// Test the testers
func Test_makeRange(t *testing.T) {
	subject := makeRange([]intrv{intrv{0, 2}, intrv{4, 5}})
	expected := []uint64{0, 1, 2, 4, 5}
	result := subject.Expand()

	if !equalUint64(expected, result) {
		t.Fatalf("Test function makeRange() Does not work: %v != %v", subject, Create(expected))
	}
}

func Test_SkipTake_Expand(t *testing.T) {

	subject := FromRaw([]uint64{5, 3, 10, 1, 1, 2})
	expected := []uint64{5, 6, 7, 18, 20, 21}
	result := subject.Expand()

	t.Logf("%v -> %v", subject.GetRaw(), result)

	if !equalUint64(expected, result) {
		t.Errorf("%v != %v", result, expected)
	}
}

func Test_SkipTake_Compress(t *testing.T) {

	subject := []uint64{2, 3, 4, 5, 9, 11, 13, 15, 16}
	expected := FromRaw([]uint64{2, 4, 3, 1, 1, 1, 1, 1, 1, 2})
	result := Create(subject)
	t.Logf("%v -> %v", subject, result)
	if !Equal(expected, result) {
		t.Errorf("Encode %v != %v", result, expected)
	}
}

func Test_SkipTake_CompressExpand(t *testing.T) {
	subject := []uint64{2, 3, 4, 5, 9, 22, 23, 24, 100, 200, 201}
	list := Create(subject)
	t.Logf("%v -> %v", subject, list)

	result := list.Expand()

	t.Logf("%v -> %v", list, result)
	if !equalUint64(subject, result) {
		t.Errorf("%v != %v", subject, result)
	}
}

func expectUint64(t *testing.T, result, expected uint64) {
	if result != expected {
		t.Errorf("%d != %d", result, expected)
	}
}

func Test_SkipTake_LargeSkip(t *testing.T) {
	subject := []uint64{0x200000000, 0x200000001, 0xaaaabbbbccccddd0, 0xaaaabbbbccccddd1, 0xaaaabbbbccccddd2}
	list := Create(subject)
	t.Logf("%v -> %v", subject, list)

	if list.Len() != 5 {
		t.Fatal("Len != 5")
	}
	result := list.Expand()

	if !equalUint64(subject, result) {
		t.Errorf("%v != %v", result, subject)
	}
}

func Test_SkipTake_LargeTake(t *testing.T) {
	list := FromRaw([]uint64{0x10, 0xffffffff, 0, 1})

	if list.Len() != 0x100000000 {
		t.Fatal("Bad big skip")
	}

	iter := list.Iterate()
	expectUint64(t, iter.Next(), 0x10)

	skip, take := iter.Seek(0xffffffff)
	expectUint64(t, skip, 0x10000000f)
	expectUint64(t, take, 1)
}
