package skiptake

// SkipTakeBuilder constructs a skip-take list from callers either appending
// absolute values in a strictly increasing sequence, or appending raw skip and
// take values.
type SkipTakeBuilder struct {
	Encoder SkipTakeEncoder
	n       uint64
	skip    uint64
	take    uint64
}

// Build() returns a skip take builder that stores the list it creates in the
// passed argument.
func Build(l *SkipTakeList) SkipTakeBuilder {
	l.Clear()
	return SkipTakeBuilder{Encoder: l.Encode()}
}

// Skip adds a skip value to the list being built. Every call to skip implies a
// take of one. Repeat calls to Skip() will NOT sum together due to this
// implied take.
//
// A skip of 0 has the same effect as incrementing the current take count. A
// non-zero skip always flushes the current take count.
func (b *SkipTakeBuilder) Skip(skip uint64) {
	b.n += skip + 1
	if skip == 0 {
		b.take++
	} else {
		b.Flush()
		b.skip = skip
		b.take = 1
	}
}

// IncTake Increases the current take count by one.
func (b *SkipTakeBuilder) IncTake() {
	b.take++
	b.n++
}

// AddTake increases the current take by the passed take amount.
func (b *SkipTakeBuilder) AddTake(take uint64) {
	b.take += take
	b.n++
}

// Next() feeds the next value of the strictly increasing sequence to encode to
// the builder.
func (b *SkipTakeBuilder) Next(n uint64) bool {
	if n < b.n {
		return false
	}
	b.Skip(n - b.n)
	return true
}

// Flush() instructs the builder that it has completed, flushing out the final
// take count.
func (b *SkipTakeBuilder) Flush() {
	if b.take > 0 {
		b.Encoder.Add(b.skip, b.take)
	}
}
