package main

// bit returns val[i].
func bit(val uint64, pos int) uint64 {
	return bits(val, pos, pos)
}

// bits returns val[hi:lo].
func bits(val uint64, hi, lo int) uint64 {
	return (val >> lo) & ((1 << (hi - lo + 1)) - 1)
}

// setBit sets 1 to val[pos] and return it.
func setBit(val uint64, pos int) uint64 {
	return val | (1 << pos)
}

// clearBit sets 0 to val[pos] and return it.
func clearBit(val uint64, pos int) uint64 {
	return val & ^(1 << pos)
}

// signExtend extends the sign of the given unsigned value.
// size must be the length of the value.
// copied from: https://gist.github.com/Code-Hex/b113083b9631f63de9b9ddc72e8c703e
func signExtend(v uint64, size int) uint64 {
	tmp := 64 - size
	return uint64((int64(v) << tmp) >> tmp)
}
