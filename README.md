# SkipTakeList

SkipTakeList is a datatype for storing a sequence of strictly increasing
integers, most efficient for sequences which have many contiguous
sub-sequences with contiguous gaps.

The basic idea is an interleaved list of skip and take instructions on how to
build the sequence from preforming 'skip' and 'take' on the sequence of all
integers.

Eg: Skip 1, take 4, skip 2, take 3 would create the sequence: (1, 2, 3, 4,
7, 8, 9), and is effectively (Skip [0]), (Take [1-4]), (Skip [5-6]), (Take
[8-9])

As most skip and take values are not actual output values, but always
positive differences between such values, the integer type used to store the
differences can be of a smaller width to conserve memory.

It is a form of low complexity bitmap compression well suited to very long
sequences of unknown length. It was written with the storage of table indexes in
mind.

The sequence stored must always be strictly increasing and non-negative, such as
a sorted sequence of array indicies.

The skip take list is stored as a sequence of variable width integers in a byte
sequence.
