package skiptake

import (
	"testing"
)

func Test_SkipTake_Union(t *testing.T) {

	list1 := Create([]uint64{10, 11, 12, 13, 14})
	list2 := Create([]uint64{15, 16, 17, 18, 19})
	list3 := Create([]uint64{31, 33, 34, 36, 37, 39})
	list4 := Create([]uint64{35, 36, 37, 38, 39, 40, 41, 42, 43, 44})

	t.Logf("Input Set 1: %v", list1)
	t.Logf("Input Set 2: %v", list2)
	t.Logf("Input Set 3: %v", list3)
	t.Logf("Input Set 4: %v", list4)

	union := Union(list1, list2, list3, list4)
	t.Logf("Union: %v", union)

	expected := []uint64{10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 31, 33, 34, 35, 36, 37, 38, 39, 40, 41, 42, 43, 44}
	result := union.Expand()
	if !equalUint64(expected, result) {
		t.Fatalf("%v != %v", result, expected)
	}
}

func Test_SkipTake_Intersection(t *testing.T) {

	list1 := Create([]uint64{10, 11, 12, 13, 14, 16, 19, 20, 21, 41})
	list2 := Create([]uint64{5, 12, 13, 14, 15, 16, 40, 41})
	list3 := makeRange([]intrv{intrv{10, 91}, intrv{100, 104}})
	list4 := Create([]uint64{1, 3, 5, 7, 9, 11, 12, 13, 14, 15, 16, 17, 19, 21, 23, 25, 40, 41})

	t.Logf("Input Set 1: %v", list1)
	t.Logf("Input Set 2: %v", list2)
	t.Logf("Input Set 3: %v", list3)
	t.Logf("Input Set 4: %v", list4)

	intersection := Intersection(list1, list2, list3, list4)
	t.Logf("Intersection: %v", intersection)

	expected := []uint64{12, 13, 14, 16, 41}
	result := intersection.Expand()
	if !equalUint64(expected, result) {
		t.Fatalf("%v != %v", result, expected)
	}
}

func testComplement(t *testing.T, subject, expected List, max uint64) {

	t.Logf("Input List: %v", subject)
	complement := ComplementMax(subject, max)
	t.Logf("Complement Set (max %d): %v", max, complement)

	if !equalList(complement, expected) {
		t.Fatalf("%v != %v", complement, expected)
		return
	}

	// Complement is a idempotent operation. Test that.
	reverse := ComplementMax(complement, max)
	t.Logf("Complement of Complement (max %d): %v", max, reverse)
	if !equalList(reverse, subject) {
		t.Fatalf("Complement is not idempotent: %v != %v", reverse, subject)
		return
	}
}

func Test_SkipTake_ComplementMax(t *testing.T) {
	const max uint64 = 19

	// Test input within range
	testComplement(
		t,
		makeRange([]intrv{intrv{2, 3}, intrv{8, 11}, intrv{17, 17}}),
		makeRange([]intrv{intrv{0, 1}, intrv{4, 7}, intrv{12, 16}, intrv{18, 19}}),
		max,
	)

	// Test input overlaps start
	testComplement(
		t,
		makeRange([]intrv{intrv{0, 1}, intrv{4, 5}}),
		makeRange([]intrv{intrv{2, 3}, intrv{6, 19}}),
		max,
	)

	// Test input overlaps end
	testComplement(
		t,
		makeRange([]intrv{intrv{4, 5}, intrv{11, 19}}),
		makeRange([]intrv{intrv{0, 3}, intrv{6, 10}}),
		max,
	)

	// Test empty input
	testComplement(
		t,
		List{},
		makeRange([]intrv{intrv{0, 19}}),
		max,
	)
}
