package main

type InstructionR struct {
	Raw, Opcode, Rd, Funct3, Rs1, Rs2, Funct7 uint64
}

func ParseR(inst uint64) *InstructionR {
	return &InstructionR{
		Raw:    inst,
		Opcode: bits(inst, 6, 0),
		Rd:     bits(inst, 11, 7),
		Funct3: bits(inst, 14, 12),
		Rs1:    bits(inst, 19, 15),
		Rs2:    bits(inst, 24, 20),
		Funct7: bits(inst, 31, 25),
	}
}

type InstructionI struct {
	Raw, Opcode, Rd, Funct3, Rs1, Imm uint64
}

func ParseI(inst uint64) *InstructionI {
	imm := bits(inst, 31, 20)
	mask := uint64(0b0)
	if imm>>11 == 1 {
		// sign-extend. inst[31] should be extended.
		mask = ^uint64(0b1111_1111_1111)
	}
	imm |= mask
	return &InstructionI{
		Raw:    inst,
		Opcode: bits(inst, 6, 0),
		Rd:     bits(inst, 11, 7),
		Funct3: bits(inst, 14, 12),
		Rs1:    bits(inst, 19, 15),
		Imm:    imm,
	}
}

type InstructionS struct {
	Raw, Opcode, Funct3, Rs1, Rs2, Imm uint64
}

func ParseS(inst uint64) *InstructionS {
	imm1 := bits(inst, 11, 7)
	imm2 := bits(inst, 31, 25)
	imm := imm1 | (imm2 << 5)
	mask := uint64(0b0)
	if imm>>11 == 0b1 {
		// sign-extend
		mask = ^uint64(0b1111_1111_1111)
	}
	imm |= mask
	return &InstructionS{
		Raw:    inst,
		Opcode: bits(inst, 6, 0),
		Funct3: bits(inst, 14, 12),
		Rs1:    bits(inst, 19, 15),
		Rs2:    bits(inst, 24, 20),
		Imm:    imm,
	}
}

type InstructionB struct {
	Raw, Opcode, Funct3, Rs1, Rs2, Imm uint64
}

func ParseB(inst uint64) *InstructionB {
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
	return &InstructionB{
		Raw:    inst,
		Opcode: bits(inst, 6, 0),
		Funct3: bits(inst, 14, 12),
		Rs1:    bits(inst, 19, 15),
		Rs2:    bits(inst, 24, 20),
		Imm:    imm,
	}
}

type InstructionU struct {
	Raw, Opcode, Rd, Imm uint64
}

func ParseU(inst uint64) *InstructionU {
	imm := bits(inst, 31, 12)
	imm = (imm << 12) | 0b0000_0000_0000
	return &InstructionU{
		Raw:    inst,
		Opcode: bits(inst, 6, 0),
		Rd:     bits(inst, 11, 7),
		Imm:    imm,
	}
}

type InstructionJ struct {
	Raw, Opcode, Rd, Imm uint64
}

func ParseJ(inst uint64) *InstructionJ {
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
	return &InstructionJ{
		Raw:    inst,
		Opcode: bits(inst, 6, 0),
		Rd:     bits(inst, 11, 7),
		Imm:    imm,
	}
}
