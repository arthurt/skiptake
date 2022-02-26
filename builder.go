package skiptake

// Builder incrementally constructs a skip-take list from a sequence of
// increasing integers or raw skip and take values.
//
// The resulting list is built by maintaining a current take count which is
// flushed whenever a non-zero skip occurs.
type Builder struct {
	Encoder Encoder
	n       uint64
	skip    uint64
	take    uint64
}

// Build returns a skip take builder that stores the list it creates in the
// passed argument. The passed list will also be returned by Finish().
//
// Note that Build returns a Builder, not a pointer to a Builder. In order for
// changes to propage, this builder should be passed by reference.
func Build(l *List) Builder {
	l.Clear()
	return Builder{Encoder: l.Encode()}
}

// Skip adds a skip value to the list being built. Every call to skip implies a
// take of one. Repeat calls to Skip() will NOT sum together due to this
// implied take.
//
// A skip of 0 has the same effect as incrementing the current take count. A
// non-zero skip always flushes the current take count.
func (b *Builder) Skip(skip uint64) {
	b.n += skip + 1
	if skip == 0 {
		b.take++
	} else {
		b.flush()
		b.skip = skip
		b.take = 1
	}
}

// Take increases the current take count by the passed amount.
func (b *Builder) Take(take uint64) {
	b.take += take
	b.n += take
}

// Next feeds the value 'n' of the strictly increasing sequence to encode to
// the builder. Next automatically detects skip and non-skips in the sequence, encoding
// appropriately. Returns true if 'n' is greater than all previous values add,
// meaning that the sequence is strictly increasing. Otherwise the value 'n' is
// ignored and false is returned.
func (b *Builder) Next(n uint64) bool {
	if n < b.n {
		return false
	}
	b.Skip(n - b.n)
	return true
}

// flush instructs the builder that it has completed, flushing out the final
// take count.
func (b *Builder) flush() {
	if b.take > 0 {
		b.Encoder.Add(b.skip, b.take)
	}
}

// Finish flushes any pending data to the built list and returns it.
func (b *Builder) Finish() List {
	b.flush()
	return *b.Encoder.Elements
}
