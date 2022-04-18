package main

// IsCompressed returns if the instruction is compressed 16-bit one.
func (cpu *CPU) IsCompressed(inst uint64) bool {
	last2bit := inst & 0b11
	// if the last 2-bit is one of 00/01/10, it is 16-bit instruction.
	return last2bit == 0b00 || last2bit == 0b01 || last2bit == 0b10
}

// Decompress extracts the given 16-bit instruction to 32-bit one.
func (cpu *CPU) DecodeCompressed(compressed uint64) (*Decoded, Exception) {
	bs := func(hi, lo int) uint64 {
		return bits(compressed, hi, lo)
	}

	op := bs(1, 0)

	switch op {
	case 0b00:
		rdRs2 := bs(4, 2)
		if rdRs2 == 0b0 {
			return nil, ExcpIllegalInstruction
		}

		funct3 := bs(15, 13)
		switch funct3 {
		case 0b000:
			// c.addi4spn
		case 0b001:
			// c.fld
		case 0b010:
			// c.lw
		case 0b011:
			// c.flw
		case 0b101:
			// c.fsd
		case 0b110:
			// c.sw
		case 0b111:
			// c.fsw
		}
	case 0b01:
		funct3 := bs(15, 13)
		switch funct3 {
		case 0b000:
			rdRs1 := bs(11, 7)
			switch rdRs1 {
			case 0b0:
				return nil, ExcpIllegalInstruction
			case 0b11:
				// c.addi16sp
			default:
				// c.lui
			}
		case 0b001:
			// c.jal
		case 0b010:
			// c.li
		case 0b011:
			rdRs1 := bs(11, 7)
			switch rdRs1 {
			case 0b0:
				// c.nop
			default:
				// c.addi
			}
		case 0b100:
			switch bs(11, 10) {
			case 0b00:
				// c.srli
			case 0b01:
				// c.srai
			case 0b10:
				// c.andi
			case 0b11:
				switch bs(6, 5) {
				case 0b00:
					// c.sub
					// -> sub rd, rd, rs2 while rd = 8 + rd', rs2 = 8 + rs2'
					rdRs1, rs2 := bs(6, 2), bs(11, 7)

					rdRs1, rs2 = rdRs1&0b111, rs2&0b111 // upper 2-bits are constant
					return &Decoded{
						Code: SUB,
						Param: &InstructionR{
							Opcode: 0b0110011,
							Rd:     rdRs1 + 8,
							Funct3: 0b000,
							Rs1:    rdRs1 + 8,
							Rs2:    rs2 + 8,
							Funct7: 0b0100000,
						},
					}, ExcpNone
				case 0b01:
					// c.xor
				case 0b10:
					// c.or
				case 0b11:
					// c.and
				}
			}
		case 0b101:
			// c.j
		case 0b110:
			// c.beqz
		case 0b111:
			// c.bnez
		}
	case 0b10:
		funct3 := bits(compressed, 15, 13)
		switch funct3 {
		case 0b000:
			switch bits(compressed, 6, 2) {
			case 0b0:
				// c.slli64
			default:
				// c.slli
			}
		case 0b001:
			// c.fldsp
		case 0b010:
			// c.lwsp
		case 0b011:
			// c.flwsp
		case 0b100:
			switch bit(compressed, 12) {
			case 0b0:
				switch bits(compressed, 6, 2) {
				case 0b0:
					// c.jr
				default:
					// c.mv
				}
			case 0b1:
				switch bits(compressed, 11, 7) {
				case 0b0:
					// c.ebreak
				default:
					switch bits(compressed, 6, 2) {
					case 0b0:
						// c.jalr
					default:
						// c.add
					}
				}
			}
		case 0b101:
			// c.fsdsp
		case 0b110:
			// c.swsp
		case 0b111:
			// c.fswsp
		}
	}

	//switch format {
	//case CompressedInstructionFormatCR:
	//	return cpu.DecompressCR(
	//		bits(compressed, 1, 0),
	//		bits(compressed, 6, 2),
	//		bits(compressed, 11, 7),
	//		bits(compressed, 15, 12),
	//	)
	//case CompressedInstructionFormatCI:
	//	return cpu.DecompressCI(
	//		bits(compressed, 1, 0),
	//		bits(compressed, 6, 2),
	//		bits(compressed, 11, 7),
	//		bits(compressed, 12, 12),
	//		bits(compressed, 15, 13),
	//	)
	//case CompressedInstructionFormatCSS:
	//case CompressedInstructionFormatCIW:
	//case CompressedInstructionFormatCL:
	//case CompressedInstructionFormatCS:
	//case CompressedInstructionFormatCB:
	//case CompressedInstructionFormatCJ:
	//}

	return nil, ExcpIllegalInstruction
}

// DecompressCR decompresses the CR format compressed instruction using the given parts.
func (cpu *CPU) DecompressCR(op, rs2, rdOrRs1, funct4 uint64) (*Decoded, Exception) {
	switch op {
	case 0b00:
		return nil, ExcpIllegalInstruction
	case 0b01:
		switch rs2 >> 3 {
		case 0b00:
		case 0b01:
			// c.xor
			// -> xor rd, rd, rs2 while rd = 8 + rd', rs2 = 8 + rs2'
			rdOrRs1, rs2 = rdOrRs1&0b111, rs2&0b111
			return &Decoded{
				Code: XOR,
				Param: &InstructionR{
					Opcode: 0b0110011,
					Rd:     rdOrRs1 + 8,
					Funct3: 0b100,
					Rs1:    rdOrRs1 + 8,
					Rs2:    rs2 + 8,
					Funct7: 0b0000000,
				},
			}, ExcpNone
		case 0b10:
			// c.or
			// -> or rd, rd, rs2 while rd = 8 + rd', rs2 = 8 + rs2'
			rdOrRs1, rs2 = rdOrRs1&0b111, rs2&0b111
			return &Decoded{
				Code: OR,
				Param: &InstructionR{
					Opcode: 0b0110011,
					Rd:     rdOrRs1 + 8,
					Funct3: 0b110,
					Rs1:    rdOrRs1 + 8,
					Rs2:    rs2 + 8,
					Funct7: 0b0000000,
				},
			}, ExcpNone
		case 0b11:
			// c.and
			// -> and rd, rd, rs2 while rd = 8 + rd', rs2 = 8 + rs2'
			rdOrRs1, rs2 = rdOrRs1&0b111, rs2&0b111
			return &Decoded{
				Code: AND,
				Param: &InstructionR{
					Opcode: 0b0110011,
					Rd:     rdOrRs1 + 8,
					Funct3: 0b111,
					Rs1:    rdOrRs1 + 8,
					Rs2:    rs2 + 8,
					Funct7: 0b0000000,
				},
			}, ExcpNone
		}
	case 0b10:
		switch funct4 & 0b1 {
		case 0b0:
			// c.mv
			// -> add rd, x0, rs2. Illegal if rs2 = x0.
			if rs2 == 0 {
				return nil, ExcpIllegalInstruction
			}

			return &Decoded{
				Code: ADD,
				Param: &InstructionR{
					Opcode: 0b0110011,
					Rd:     rdOrRs1,
					Funct3: 0b000,
					Rs1:    0, // x0
					Rs2:    rs2,
					Funct7: 0b0000000,
				},
			}, ExcpNone
		case 0b1:
			// c.add
			// -> add rd, rd, rs2. Illegal if rd = x0 || rs2 = x0.
			if rdOrRs1 == 0 || rs2 == 0 {
				return nil, ExcpIllegalInstruction
			}

			return &Decoded{
				Code: ADD,
				Param: &InstructionR{
					Opcode: 0b0110011,
					Rd:     rdOrRs1,
					Funct3: 0b000,
					Rs1:    rdOrRs1,
					Rs2:    rs2,
					Funct7: 0b0000000,
				},
			}, ExcpNone
		}
	}

	return nil, ExcpIllegalInstruction
}

// DecompressCI decompresses the CR format compressed instruction using the given parts.
func (cpu *CPU) DecompressCI(op, imm1, rdOrRs1, imm2, funct3 uint64) (*Decoded, Exception) {
	switch op {
	case 0b00:
		return nil, ExcpIllegalInstruction
	case 0b01:
		switch funct3 {
		case 0b000:
			switch rdOrRs1 {
			case 0b0:
				// c.nop
				// -> addi x0, x0, 0 (do nothing)
				return &Decoded{
					Code: ADDI,
					Param: &InstructionI{
						Opcode: 0b0010011,
						Rd:     0b0,
						Funct3: 0b000,
						Rs1:    0b0,
						Imm:    0b0,
					},
				}, ExcpNone
			default:
				// c.addi
				// -> addi rd, rd, nzimm[5:0] while imm1 = nzimm[4:0], imm2 = nzimm[5]
				var mask uint32 = 0b0
				if imm2 == 0b1 {
					// Sign-extend. if imm2 is 1, imm[31:6] should be 1.
					// Also, set imm[5](=imm2) 1 here.
					mask = ^uint32(0b1_1111)
				}
				nzimm := imm1 | uint64(int64(int32(mask)))

				return &Decoded{
					Code: ADDI,
					Param: &InstructionI{
						Opcode: 0b0010011,
						Rd:     rdOrRs1,
						Funct3: 0b000,
						Rs1:    rdOrRs1,
						Imm:    nzimm,
					},
				}, ExcpNone
			}
		case 0b010:
			// c.li
			// -> addi rd, x0, imm
			var mask uint32 = 0b0
			if imm2 == 0b1 {
				// Sign-extend. if imm2 is 1, imm[31:6] should be 1.
				// Also, set imm[5](=imm2) 1 here.
				mask = ^uint32(0b1_1111)
			}
			nzimm := imm1 | uint64(int64(int32(mask)))

			return &Decoded{
				Code: ADDI,
				Param: &InstructionI{
					Opcode: 0b0010011,
					Rd:     rdOrRs1,
					Funct3: 0b000,
					Rs1:    0b0, // x0
					Imm:    nzimm,
				},
			}, ExcpNone
		case 0b011:
			switch rdOrRs1 {
			case 0b10:
				// c.addi16sp
				// -> addi x2, x2, imm. Illegal if imm is 0.
				var mask uint32 = 0b0
				if imm2 == 0b1 {
					// Sign-extend. if imm2 is 1, imm[31:10] should be 1.
					// Also, set imm[9](=imm2) 1 here.
					mask = ^uint32(0b1_1111_1111)
				}
				imm := uint64(int64(int32(mask))) |
					(imm1 & 0b1_0000) | // imm1[4] -> imm[4]
					((imm1 << 3) & 0b100_0000) | // imm1[3] -> imm[6]
					((imm1 << 6) & 0b1_0000_0000) | // imm1[2] -> imm[8]
					((imm1 << 6) & 0b1000_0000) | // imm1[1] -> imm[7]
					((imm1 << 5) & 0b10_0000) // imm1[0] -> imm[5]

				return &Decoded{
					Code: ADDI,
					Param: &InstructionI{
						Opcode: 0b0010011,
						Rd:     0b10, // x2
						Funct3: 0b000,
						Rs1:    0b10, // x2
						Imm:    imm,
					},
				}, ExcpNone
			case 0b00:
				// must not be 0
				return nil, ExcpIllegalInstruction
			default:
				// c.lui
				// -> lui rd, imm. Illegal if rd = x2 || imm = 0
				if rdOrRs1 == 0b10 || rdOrRs1 == 0b0 {
					return nil, ExcpIllegalInstruction
				}

				if imm1 == 0b0 && imm2 == 0b0 {
					return nil, ExcpIllegalInstruction
				}

				var mask uint64 = 0b0
				if imm2 == 0b1 {
					// Sign-extend. if imm2 is 1, imm[31:18] should be 1.
					// Also, set imm[17](=imm2) 1 here.
					mask = 0b1111_1111_1111_1110_0000_0000_0000_0000
				}
				imm := uint64(int64(int32(mask))) | (imm1 << 12) // imm1 -> imm[16:12]

				return &Decoded{
					Code: LUI,
					Param: &InstructionU{
						Opcode: 0b0110111,
						Rd:     rdOrRs1,
						Imm:    imm,
					},
				}, ExcpNone
			}
		case 0b100:
			switch rdOrRs1 >> 3 {
			case 0b00:
				// c.srli
				// -> srli rd, rd, uimm while rd = 8 + rd'
				rdOrRs1 = rdOrRs1 & 0b111 // left 2-bit of rd is opcode
				rdOrRs1 += 8
				uimm := imm1 | (imm2 << 5) // imm2 -> uimm[5]

				return &Decoded{
					Code: SRLI,
					Param: &InstructionI{
						Opcode: 0b0010011,
						Rd:     rdOrRs1,
						Funct3: 0b101,
						Rs1:    rdOrRs1,
						Imm:    uimm,
					},
				}, ExcpNone
			case 0b01:
				// c.srai
				// -> srai rd, rd, uimm while rd = 8 + rd'
				rdOrRs1 = rdOrRs1 & 0b111 // left 2-bit of rd is opcode
				rdOrRs1 += 8
				uimm := imm1 | (imm2 << 5) // imm2 -> uimm[5]

				return &Decoded{
					Code: SRAI,
					Param: &InstructionI{
						Opcode: 0b0010011,
						Rd:     rdOrRs1,
						Funct3: 0b101,
						Rs1:    rdOrRs1,
						Imm:    uimm,
					},
				}, ExcpNone
			case 0b10:
				// c.andi
				// -> andi rd, rd, imm while rd = 8 + rd'
				rdOrRs1 = rdOrRs1 & 0b111 // left 2-bit of rd is opcode
				rdOrRs1 += 8
				uimm := imm1 | (imm2 << 5) // imm2 -> uimm[5]

				return &Decoded{
					Code: ANDI,
					Param: &InstructionI{
						Opcode: 0b0010011,
						Rd:     rdOrRs1,
						Funct3: 0b101,
						Rs1:    rdOrRs1,
						Imm:    uimm,
					},
				}, ExcpNone
			}
		}
	case 0b10:
		switch funct3 {
		case 0b000:
			switch imm1 {
			case 0b0:
				// c.slli64
			default:
				// c.slli
			}
		case 0b100:
			// c.ebreak
		}
	}

	return nil, ExcpIllegalInstruction
}
