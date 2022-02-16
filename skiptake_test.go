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

func equalSkipTakeList(a, b SkipTakeList) bool {
	ad := a.Decode()
	bd := b.Decode()

	for !ad.EOS() {
		as, at := ad.Next()
		bs, bt := bd.Next()
		if as != bs || at != bt {
			return false
		}
	}

	if !bd.EOS() {
		return false
	}
	return true
}

type intrv [2]uint

func makeRange(args []intrv) SkipTakeIntList {
	ret := SkipTakeIntList{}

	b := Build(&ret)

	for _, p := range args {
		b.Next(uint64(p[0]))
		b.AddTake(uint64(p[1] - p[0]))
	}
	b.Flush()
	return ret
}

func Test_SkipTake_Expand(t *testing.T) {

	subject := SkipTakeIntList{5, 3, 10, 1, 1, 2}
	expected := []uint64{5, 6, 7, 18, 20, 21}
	result := subject.Expand()

	t.Logf("%v -> %v", []uint32(subject), result)

	if !equalUint64(expected, result) {
		t.Fatalf("%v != %v", result, expected)
	}
}

func Test_SkipTake_Compress(t *testing.T) {

	subject := []uint64{2, 3, 4, 5, 9, 11, 13, 15, 16}
	expected := SkipTakeIntList{2, 4, 3, 1, 1, 1, 1, 1, 1, 2}
	result := Create(&SkipTakeIntList{}, subject)
	t.Logf("%v -> %v", subject, result)
	if !equalSkipTakeList(expected, result) {
		t.Fatalf("Encode %v != %v", result, expected)
	}
}

func Test_SkipTake_CompressExpand(t *testing.T) {
	subject := []uint64{2, 3, 4, 5, 9, 22, 23, 24, 100, 200, 201}
	list := Create(&SkipTakeIntList{}, subject)
	t.Logf("%v -> %v", subject, list)

	result := Expand(list)

	t.Logf("%v -> %v", list, result)
	if !equalUint64(subject, result) {
		t.Fatalf("%v != %v", subject, result)
	}
}

func expectUint64(t *testing.T, result, expected uint64) {
	if result != expected {
		t.Fatalf("%d != %d", result, expected)
	}
}

func Test_SkipTake_Seek(t *testing.T) {

	subject := []uint64{10, 11, 12, 13, 14, 20, 21, 22, 30, 40, 41, 42, 43, 44, 50}
	list := Create(&SkipTakeListVarInt{}, subject)
	t.Logf("%v -> %v", subject, list)

	iter := Iterate(list)
	skip, take := iter.NextSkipTake()

	expectUint64(t, skip, 10)
	expectUint64(t, take, 5)

	n := iter.Next()
	expectUint64(t, n, 10)

	skip, take = iter.Seek(5)
	expectUint64(t, skip, 20)
	expectUint64(t, take, 3)

	// Check that next works after a seek
	n = iter.Next()
	expectUint64(t, n, 20)

	// Seek for the current position
	_, _ = iter.Seek(6)
	n = iter.Next()
	expectUint64(t, n, 21)

	// Seek again
	skip, take = iter.Seek(10)
	expectUint64(t, skip, 41)
	expectUint64(t, take, 4)

	// Check NextSkipTake works after a seek
	skip, take = iter.NextSkipTake()
	expectUint64(t, skip, 5)
	expectUint64(t, take, 1)

	//Seek backwards
	skip, take = iter.Seek(1)
	expectUint64(t, skip, 11)
	expectUint64(t, take, 4)

	// Seek beyond the end
	skip, take = iter.Seek(16)
	expectUint64(t, take, 0)
}

func Test_SkipTake_LargeSkip(t *testing.T) {
	subject := []uint64{0x200000000, 0x200000001, 0xaaaabbbbccccddd0, 0xaaaabbbbccccddd1, 0xaaaabbbbccccddd2}
	list := Create(&SkipTakeListVarInt{}, subject)
	t.Logf("%v -> %v", subject, list)

	if Len(list) != 5 {
		t.Fatal("Len != 5")
	}
	result := Expand(list)

	if !equalUint64(subject, result) {
		t.Fatalf("%v != %v", result, subject)
	}
}

func Test_SkipTake_LargeTake(t *testing.T) {
	list := SkipTakeIntList{0x10, 0xffffffff, 0, 1}

	if Len(list) != 0x100000000 {
		t.Fatal("Bad big skip")
	}

	iter := Iterate(list)
	expectUint64(t, iter.Next(), 0x10)

	skip, take := iter.Seek(0xffffffff)
	expectUint64(t, skip, 0x10000000f)
	expectUint64(t, take, 1)
}

func Test_SkipTake_IterNextSkipTake(t *testing.T) {
	list := SkipTakeIntList{1, 10, 2, 20, 3, 30, 4, 40}
	iter := Iterate(list)

	skip, take := iter.NextSkipTake()
	expectUint64(t, skip, 1)
	expectUint64(t, take, 10)

	skip, take = iter.NextSkipTake()
	expectUint64(t, skip, 2)
	expectUint64(t, take, 20)

	skip, take = iter.NextSkipTake()
	expectUint64(t, skip, 3)
	expectUint64(t, take, 30)

	skip, take = iter.NextSkipTake()
	expectUint64(t, skip, 4)
	expectUint64(t, take, 40)

	skip, take = iter.NextSkipTake()
	expectUint64(t, skip, 0)
	expectUint64(t, take, 0)
}

func Test_SkipTake_IterNextSkipTakeSeek(t *testing.T) {
	list := SkipTakeIntList{1, 10, 2, 20, 3, 30, 4, 40}
	iter := Iterate(list)

	t.Logf("%v", list)

	skip, take := iter.NextSkipTake()
	expectUint64(t, skip, 1)
	expectUint64(t, take, 10)

	skip, take = iter.Seek(12)
	expectUint64(t, skip, 15)
	expectUint64(t, take, 18)
	expectUint64(t, iter.Next(), 15)

	skip, take = iter.NextSkipTake()
	expectUint64(t, skip, 3)
	expectUint64(t, take, 30)

	skip, take = iter.Seek(0)
	expectUint64(t, skip, 1)
	expectUint64(t, take, 10)
}
