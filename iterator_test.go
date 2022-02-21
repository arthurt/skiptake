package skiptake

import "testing"

func Test_SkipTake_Seek(t *testing.T) {

	subject := []uint64{10, 11, 12, 13, 14, 20, 21, 22, 30, 40, 41, 42, 43, 44, 50}
	list := Create(subject)
	t.Logf("%v -> %v", subject, list)

	iter := list.Iterate()
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

func Test_SkipTake_IterNextSkipTake(t *testing.T) {
	list := FromRaw([]uint64{1, 10, 2, 20, 3, 30, 4, 40})
	iter := list.Iterate()

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
	list := FromRaw([]uint64{1, 10, 2, 20, 3, 30, 4, 40})
	iter := list.Iterate()

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
