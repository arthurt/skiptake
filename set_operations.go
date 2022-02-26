package skiptake

// Set algebra operations on skiptake.List instances

import (
	"math"
)

// Union returns a new List that is the computed set algebra union of the passed
// slice of lists.
func Union(lists ...List) List {
	b := Build(&List{})
	iter := make([]Iterator, len(lists))
	for i := range lists {
		iter[i] = lists[i].Iterate()
		iter[i].NextSkipTake()
	}
	union(&b, iter)
	return b.Finish()
}

func union(result *Builder, iter []Iterator) {
	var n, r, l uint64
	for {
		// Find the lowest start of a new range
		n = math.MaxUint64
		for _, it := range iter {
			if !it.EOS() {
				if it.n < n {
					n = it.n
					r = it.take + it.n - 1
				}
			}
		}
		// All ranges are done
		if n == math.MaxUint64 {
			break
		}

		// Have the start of a new output range.
		result.Skip(n - l)

		// Run through the sequences, eating ranges within our found range,
		// while also extending the output range should a found range start
		// within but extend beyond it.
		for {
			found := false
			for i := range iter {
				it := &iter[i]
				for !it.EOS() && it.n <= r {
					if it.take+it.n-1 > r {
						found = true
						r = it.take + it.n - 1
					}
					it.NextSkipTake()
				}
			}
			// Hit a discontinuity
			if !found {
				break
			}
		}
		result.Take(r - n)
		l = r + 1
	}
}

// Intersection returns a new List that is the computed set algebra intersection
// of the passed slice of lists.
func Intersection(lists ...List) List {
	b := Build(&List{})
	iter := make([]Iterator, len(lists))
	for i := range lists {
		iter[i] = lists[i].Iterate()
		iter[i].NextSkipTake()
	}
	intersection(&b, iter)
	return b.Finish()
}

func intersection(result *Builder, iter []Iterator) {
	var n, r, l uint64
outer:
	for n != math.MaxUint64 {
		for i := range iter {
			it := &iter[i]
			// Scan intervals while they are before our candidate area.
			for !it.EOS() && it.n+it.take-1 < n {
				it.NextSkipTake()
			}

			if it.n > n {
				// Increased the lower bound of the candidate interval.
				n = it.n
				// Rescan all sequences.
				continue outer
			}
		}

		// All sequences are at an interval that includes our candidate.
		// Find the longest range of all intervals.
		r = math.MaxUint64
		for _, it := range iter {
			if it.take+it.n-1 < r {
				r = it.take + it.n - 1
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
	complement(&b, list.Decode(), max)
	return b.Finish()
}

func complement(result *Builder, set Decoder, max uint64) {
	var n uint64
	for {
		skip, take := set.Next()
		if skip > 0 {
			if n+skip >= max {
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
		if n+take >= max {
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
