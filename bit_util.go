package main

// bit returns val[i].
func bit(val uint64, pos int) uint64 {
	return bits(val, pos, pos)
}

// bits returns val[hi:lo].
func bits(val uint64, hi, lo int) uint64 {
	return (val >> lo) & ((1 << (hi - lo + 1)) - 1)
}

// setBit sets 1 to val[pos]
func setBit(val uint64, pos int) uint64 {
	return val | (1 << pos)
}

// clearBit sets 0 to val[pos]
func clearBit(val uint64, pos int) uint64 {
	return val & ^(1 << pos)
}
