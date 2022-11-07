package skiptake

// Set algebra operations on skiptake.List instances

import (
	"container/heap"
	"math"
)

// firstHeap implements the container/heap.Interface for a set of Iterators.
//
// We choose to implement Less only on the interval first value, so it is as
// fast as possible.
type firstHeap []Iterator

func (m firstHeap) Len() int      { return len(m) }
func (m firstHeap) Swap(i, j int) { m[i], m[j] = m[j], m[i] }
func (m firstHeap) Less(i, j int) bool {
	a, b := m[i].n, m[j].n
	if a == b {
		return m[i].take < m[j].take
	}
	return a < b
}

func (m *firstHeap) Push(x interface{}) {
	*m = append(*m, x.(Iterator))
}

func (m *firstHeap) Pop() interface{} {
	old := *m
	n := len(old)
	x := old[n-1]
	*m = old[:n-1]
	return x
}

// Union returns a new List that is the computed set algebra union of the passed
// slice of lists.
func Union(lists ...List) List {
	b := Build(&List{})
	iter := make(firstHeap, len(lists))
	if len(lists) > 0 {
		for i := range lists {
			iter[i] = lists[i].Iterate()
			// Prime
			iter[i].NextSkipTake()
		}
		union(&b, iter)
	}
	return b.Finish()
}

func union(result *Builder, iter firstHeap) {
	var n uint64 // Current candidate intersection interval first value
	var r uint64 // Current candidate intersection interval last value
	var l uint64 // Proceededing non-intersection interval first value

	heap.Init(&iter)
	// Starting is a special case because of zero skips.
	n, r = iter[0].Interval()
	if n > r { // EOS
		return
	}
	iter[0].NextInterval()
	heap.Fix(&iter, 0)
	result.Skip(n)
	result.Take(r - n)

	for {
		first, last := iter[0].Interval()
		if first > last { // EOS
			return
		}
		if first > r {
			// Next earliest interval starts after our current candiate.
			// Have a discontinuity.
			l = r + 1
			n, r = first, last
			result.Skip(n - l)
			result.Take(r - n)
		} else if last > r {
			// Extend the candidate interval.
			result.Take(last - r)
			r = last
		}
		iter[0].NextInterval()
		heap.Fix(&iter, 0)
	}
}

// Intersection returns a new List that is the computed set algebra intersection
// of the passed slice of lists.
func Intersection(lists ...List) List {
	b := Build(&List{})
	iter := make([]Iterator, len(lists))
	for i := range lists {
		iter[i] = lists[i].Iterate()
	}
	intersection(&b, iter)
	return b.Finish()
}

func intersection(result *Builder, iter []Iterator) {

	// Handle a degenerate case out of hand
	if len(iter) == 0 {
		return
	}

	var n uint64 // Current candidate intersection interval first value
	var r uint64 // Current candidate intersection interval last value
	var l uint64 // Proceededing non-intersection interval first value
outer:
	for r != math.MaxUint64 {
		for i := range iter {
			// Scan intervals while they are before our candidate area.
			it := &iter[i]
			for first, last := it.Interval(); ; first, last = it.NextInterval() {
				if first > last { // EOS
					return
				}
				if last >= n {
					if first > n {
						// Increased the lower bound of the candidate interval.
						n = first
						// Rescan all sequences.
						continue outer
					}
					break
				}
			}
		}

		// All sequences are at an interval that includes our candidate.
		// Find the longest range of all intervals.
		r = math.MaxUint64
		for _, it := range iter {
			if _, last := it.Interval(); last < r {
				r = last
			}
		}
		result.Skip(n - l)
		result.Take(r - n)
		l = r + 1
		n = l
	}
}

// Complement returns a new List that is the set algebra complement of the
// passed List set. The range of the returned list is [0, math.MaxUint64].
func Complement(list List) List {
	return ComplementMax(list, math.MaxUint64)
}

// ComplementMax returns a new List that is the set algebra complement of the
// passed List set, bounded to the range [0, max].
func ComplementMax(list List, max uint64) List {
	b := Build(&List{})
	complement(&b, list.Iterate(), max)
	return b.Finish()
}

func complement(result *Builder, set Iterator, max uint64) {
	var n uint64
	for {
		skip, take := set.NextSkipTake()
		if skip > 0 {
			if max-n <= skip {
				// End boundary case. Next source skip starts outside our max
				// boundary. Add a take for the rest of the output range and be
				// done.
				result.Take(max - n + 1)
				n += skip
				return
			}
			if n == 0 {
				// Beginning boundary case. Need to add a zero-skip and take if
				// the source starts with a skip.
				result.Skip(0)
			}
			// Non-boundary case. Take the skip, minus the implied source take.
			result.Take(skip - 1)

		}
		n += skip
		if take == 0 {
			// Input source is done.
			break
		}
		if max-n <= take {
			// End boundary case. Next source take finished outside our max
			// boundary.
			n += take
			return
		}
		// Non-boundary case. Add source take as a skip.
		result.Skip(take)

		n += take
	}

	// End-boundary case. Source list finished before our max boundary. Add a
	// take for the rest of the interval between the end of the source to the
	// max boundary.
	if n <= max {
		if n == 0 {
			result.Skip(0)
		}
		result.Take(max - n)
	}
}
