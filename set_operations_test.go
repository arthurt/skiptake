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

func Test_SkipTake_Complement(t *testing.T) {
	const max uint64 = 20

	// Test input within range
	list := Create([]uint64{2, 3, 8, 9, 10, 11, 17})
	t.Logf("Input Set: %v", list)
	complement := ComplementMax(list, max)
	t.Logf("Complement Set (range %d): %v", max, complement)

	expected := []uint64{0, 1, 4, 5, 6, 7, 12, 13, 14, 15, 16, 18, 19}
	result := complement.Expand()
	if !equalUint64(expected, result) {
		t.Fatalf("%v != %v", result, expected)
	}

	// Test input overlaps start
	list = Create([]uint64{0, 1, 2, 4, 5})
	t.Logf("Input Set: %v", list)
	complement = ComplementMax(list, max)
	t.Logf("Complement Set (range %d): %v", max, complement)

	expected = []uint64{3, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19}
	result = complement.Expand()
	if !equalUint64(expected, result) {
		t.Fatalf("%v != %v", result, expected)
	}

	// Test input overlaps end
	list = Create([]uint64{4, 5, 11, 12, 13, 14, 15, 16, 17, 18, 19})
	t.Logf("Input Set: %v", list)
	complement = ComplementMax(list, max)
	t.Logf("Complement Set (range %d): %v", max, complement)

	expected = []uint64{0, 1, 2, 3, 6, 7, 8, 9, 10}
	result = complement.Expand()
	if !equalUint64(expected, result) {
		t.Fatalf("%v != %v", result, expected)
	}
}
