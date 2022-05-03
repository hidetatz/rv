package main

func ParseIImm(inst uint64) uint64 {
	// sign-extend
	return uint64(int64(int32(inst)) >> 20)
}

func ParseSImm(inst uint64) uint64 {
	imm1 := bits(inst, 11, 7)
	imm2 := bits(inst, 31, 25)
	imm := imm1 | (imm2 << 5)
	mask := uint64(0b0)
	if imm>>11 == 0b1 {
		// sign-extend
		mask = ^uint64(0b1111_1111_1111)
	}
	imm |= mask
	return imm
}

func ParseBImm(inst uint64) uint64 {
	imm1 := bits(inst, 11, 7)  // imm1[4:0] -> imm[4:1|11]
	imm2 := bits(inst, 31, 25) // imm2[7:0] -> imm[12|10:5]

	imm := ((imm1 & 0b1) << 11) | // imm1[0] -> imm[11]
		(imm1 & 0b1_1110) | // imm1[4:1] -> imm[4:1]
		((imm2 & 0b11_1111) << 5) | // imm2[5:0] -> imm[10:5]
		((imm2 & 0b100_0000) << 6) // imm2[6] -> imm[12]

	mask := uint64(0b0)
	if imm>>12 == 0b1 {
		// sign-extend
		mask = ^uint64(0b1_1111_1111_1111)
	}
	imm |= mask
	return imm
}

func ParseJImm(inst uint64) uint64 {
	i := bits(inst, 31, 12)            // imm[20|10:1|11|19:12]
	imm := ((i & 0b1111_1111) << 12) | // i[7:0] -> imm[19:12]
		((i & 0b1_0000_0000) << 3) | // i[8] -> imm[11]
		((i & 0b111_1111_1110_0000_0000) >> 8) | // i[18:9] -> imm[10:1]
		(i & 0b1000_0000_0000_0000_0000 << 1) // i[19] -> imm[20]

	mask := uint64(0b0)
	if imm>>20 == 0b1 {
		mask = ^uint64(0b1_1111_1111_1111_1111_1111)
	}
	imm |= mask
	return imm
}
