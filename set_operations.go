package skiptake

import (
	"math"
)

func Union(lset ...SkipTakeList) SkipTakeList {
	iter := make([]SkipTakeIterator, len(lset))
	for i, l := range lset {
		iter[i] = Iterate(l)
		iter[i].NextSkipTake()
	}
	e := SkipTakeEncoder{Elements: &SkipTakeList{}}

	var n, r, l uint64
	for {
		n = math.MaxUint64
		for _, it := range iter {
			if !it.EOS() {
				//fmt.Printf("%d: %d - %d\n", i, it.n, it.n+it.take-1)
				if it.n < n {
					n = it.n
					r = it.take + it.n - 1
				}
			}
		}
		if n == math.MaxUint64 {
			break
		}

		//fmt.Printf("Found a skip of %d (pos %d)\n", n-l, n)
		e.Skip(n - l)

		for {
			found := false
			for i := range iter {
				it := &iter[i]
				for !it.EOS() && it.n <= r {
					if it.take+it.n-1 > r {
						found = true
						r = it.take + it.n - 1
						//fmt.Printf("Extending take to end at %d\n", r)
					}
					it.NextSkipTake()
				}
			}
			if !found {
				break
			}
		}
		//fmt.Printf("Found a take of %d (end pos %d)\n", (r - n + 1), r)
		e.AddTake(r - n)
		l = r + 1
	}
	e.Flush()
	return *e.Elements
}

func Intersection(lset ...SkipTakeList) SkipTakeList {
	iter := make([]SkipTakeIterator, len(lset))
	for i, l := range lset {
		iter[i] = Iterate(l)
		iter[i].NextSkipTake()
	}
	e := SkipTakeEncoder{Elements: &SkipTakeList{}}

	var n, r, l uint64

outer:
	for n != math.MaxUint64 {
		//fmt.Printf("Candidate range starts at %d\n", n)
		for i := range iter {
			it := &iter[i]
			//fmt.Printf("%d: %d - %d\n", i, it.n, it.n+it.take-1)
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
				//fmt.Printf("Shrinking end to %d\n", r)
			}
		}
		e.Skip(n - l)
		e.AddTake(r - n)
		l = r + 1
		n = l
	}

	e.Flush()
	return *e.Elements
}

func Complement(slist SkipTakeList, max uint64) SkipTakeList {
	dec := SkipTakeDecoder{Elements: slist}
	e := SkipTakeEncoder{Elements: &SkipTakeList{}}
	var n uint64

	for {
		skip, take := dec.Next()
		if skip > 0 {
			if n+skip > max {
				e.AddTake(max - n)
				n += skip
				break
			} else {
				if n == 0 {
					e.Skip(0)
				}
				e.AddTake(skip - 1)
			}
		}
		n += skip
		if take == 0 {
			break
		}
		if n+take > max {
			n += take
			break
		} else {
			e.Skip(take)
		}
		n += take
	}

	if n < max {
		if n == 0 {
			e.AddTake(max)
		} else {
			e.AddTake(max - n - 1)
		}
	}

	e.Flush()
	return *e.Elements
}
