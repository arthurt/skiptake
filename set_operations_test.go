package skiptake

import (
	"testing"
)

func testUnion(t *testing.T, expected []uint64, lists ...List) {
	for i, l := range lists {
		t.Logf("Input Set %d: %v", i, l)
	}
	union := Union(lists...)
	t.Logf("Union: %v", union)
	result := union.Expand()
	if !equalUint64(expected, result) {
		t.Errorf("%v != %v", union, Create(expected...))
	}
}

func TestSetOperationsUnion(t *testing.T) {
	t.Run("NoList", func(t *testing.T) {
		testUnion(t, []uint64{})
	})

	t.Run("EmptyList", func(t *testing.T) {
		testUnion(t, []uint64{}, List{})
	})

	t.Run("SingleList", func(t *testing.T) {
		list := Create(3, 4, 5, 15, 16)
		// Union with one argument should return itself
		testUnion(t, list.Expand(), list)
	})

	t.Run("DifferentLength", func(t *testing.T) {
		testUnion(t,
			[]uint64{0, 1, 2, 4, 5, 6, 9, 10},
			Create(2),
			Create(1, 4),
			Create(0, 5, 6, 9, 10),
		)
	})

	t.Run("Common", func(t *testing.T) {
		testUnion(t,
			[]uint64{10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 31, 33, 34, 35, 36, 37, 38, 39, 40, 41, 42, 43, 44},
			Create(10, 11, 12, 13, 14),
			Create(15, 16, 17, 18, 19),
			Create(31, 33, 34, 36, 37, 39),
			Create(35, 36, 37, 38, 39, 40, 41, 42, 43, 44),
		)
	})

	t.Run("Combine", func(t *testing.T) {
		testUnion(t,
			[]uint64{1, 3, 5, 10, 15, 20, 25},
			Create(1, 5, 10, 15, 20, 25),
			Create(3),
		)
	})

	t.Run("Insert", func(t *testing.T) {
		testUnion(t,
			[]uint64{21, 31, 38, 53},
			Create(31),
			Create(38),
			Create(53),
			Create(21),
		)
	})

	t.Run("MaxRange", func(t *testing.T) {
		testUnion(t,
			[]uint64{10, 30, 0xfffffffffffffffe, 0xffffffffffffffff},
			Create(10),
			Create(30, 0xfffffffffffffffe),
			Create(0xffffffffffffffff),
		)
	})
}

func testIntersection(t *testing.T, expected []uint64, lists ...List) {
	for i, l := range lists {
		t.Logf("Input Set %d: %v", i, l)
	}
	intersection := Intersection(lists...)
	t.Logf("Intersection: %v", intersection)
	result := intersection.Expand()
	if !equalUint64(expected, result) {
		t.Errorf("%v != %v", result, expected)
	}
}

func TestSetOperationsIntersection(t *testing.T) {
	t.Run("NoList", func(t *testing.T) {
		testIntersection(t, []uint64{})
	})

	t.Run("EmptyList", func(t *testing.T) {
		testIntersection(t, []uint64{}, List{})
	})

	t.Run("SingleList", func(t *testing.T) {
		list := Create(3, 4, 5, 15, 16)
		// Interestion with one argument should return itself
		testIntersection(t, list.Expand(), list)
	})

	t.Run("Common", func(t *testing.T) {
		testIntersection(t,
			[]uint64{12, 13, 14, 16, 41},
			Create(10, 11, 12, 13, 14, 16, 19, 20, 21, 41),
			Create(5, 12, 13, 14, 15, 16, 40, 41, 50, 51, 53),
			makeRange([]intrv{intrv{10, 91}, intrv{100, 104}}),
			Create(1, 3, 5, 7, 9, 11, 12, 13, 14, 15, 16, 17, 19, 21, 23, 25, 40, 41),
		)
	})

	t.Run("NoCommon", func(t *testing.T) {
		testIntersection(t,
			[]uint64{},
			Create(10, 11, 19, 20, 21, 42),
			Create(5, 40, 41, 50, 51, 53),
			makeRange([]intrv{intrv{17, 91}, intrv{100, 104}}),
			Create(1, 3, 5, 7, 9, 11, 17, 19, 21, 23, 25, 40, 41),
		)
	})

	t.Run("MaxRange", func(t *testing.T) {
		testIntersection(t,
			[]uint64{0xfffffffffffffffe, 0xffffffffffffffff},
			Create(0xfffffffffffffffe, 0xffffffffffffffff),
			Create(0, 1, 2, 0xfffffffffffffffe, 0xffffffffffffffff),
			makeRange([]intrv{intrv{100, 110}, intrv{200, 210}, intrv{0xfffffffffffffff0, 0xffffffffffffffff}}),
			Create(40, 42, 44, 0xfffffffffffffffe, 0xffffffffffffffff),
		)
	})
}

func testComplement(t *testing.T, subject, expected List, max uint64) {

	t.Logf("Input List: %v", subject)
	complement := ComplementMax(subject, max)
	t.Logf("Complement Set (max %d): %v", max, complement)

	if !Equal(complement, expected) {
		t.Errorf("%v != %v", complement, expected)
		return
	}

	// Complement is a idempotent operation. Test that.
	reverse := ComplementMax(complement, max)
	t.Logf("Complement of Complement (max %d): %v", max, reverse)
	if !Equal(reverse, subject) {
		t.Errorf("Complement is not idempotent: %v != %v", reverse, subject)
		return
	}
}

func TestSetOperationsComplementMax(t *testing.T) {
	const max uint64 = 19

	t.Run("Empty", func(t *testing.T) {
		testComplement(
			t,
			List{},
			makeRange([]intrv{intrv{0, 19}}),
			max,
		)
	})

	t.Run("InRange", func(t *testing.T) {
		testComplement(
			t,
			makeRange([]intrv{intrv{2, 3}, intrv{8, 11}, intrv{17, 17}}),
			makeRange([]intrv{intrv{0, 1}, intrv{4, 7}, intrv{12, 16}, intrv{18, 19}}),
			max,
		)
	})

	t.Run("OverlapStart", func(t *testing.T) {
		testComplement(
			t,
			makeRange([]intrv{intrv{0, 1}, intrv{4, 5}}),
			makeRange([]intrv{intrv{2, 3}, intrv{6, 19}}),
			max,
		)
	})

	t.Run("OverlapEnd", func(t *testing.T) {
		testComplement(
			t,
			makeRange([]intrv{intrv{4, 5}, intrv{11, 19}}),
			makeRange([]intrv{intrv{0, 3}, intrv{6, 10}}),
			max,
		)
	})
}
