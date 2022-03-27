package main

// CompressedInstructionFormat is a format for RV32C/RV64C.
type CompressedInstructionFormat uint8

const (
	CompressedInstructionFormatInvalid CompressedInstructionFormat = iota
	CompressedInstructionFormatCR
	CompressedInstructionFormatCI
	CompressedInstructionFormatCSS
	CompressedInstructionFormatCIW
	CompressedInstructionFormatCL
	CompressedInstructionFormatCS
	CompressedInstructionFormatCB
	CompressedInstructionFormatCJ
)

// IsCompressed returns if the instruction is compressed 16-bit one.
func (cpu *CPU) IsCompressed(inst uint64) bool {
	last2bit := inst & 0b11
	// if the last 2-bit is one of 00/01/10, it is 16-bit instruction.
	return last2bit == 0b00 || last2bit == 0b01 || last2bit == 0b10
}

// Decompress extracts the given 16-bit instruction to 32-bit one.
func (cpu *CPU) Decompress(compressed uint64) (uint64, Exception) {
	compressedInstructionFormat := cpu.DecodeCompressedInstructionFormat(compressed)
	if compressedInstructionFormat == CompressedInstructionFormatInvalid {
		return 0, ExcpIllegalInstruction
	}

	switch compressedInstructionFormat {
	case CompressedInstructionFormatCR:
		return cpu.DecompressCR(
			bits(compressed, 1, 0),
			bits(compressed, 6, 2),
			bits(compressed, 11, 7),
			bits(compressed, 15, 12),
		)
	case CompressedInstructionFormatCI:
		return cpu.DecompressCI(
			bits(compressed, 1, 0),
			bits(compressed, 6, 2),
			bits(compressed, 11, 7),
			bits(compressed, 12, 12),
			bits(compressed, 15, 13),
		)
	case CompressedInstructionFormatCSS:
	case CompressedInstructionFormatCIW:
	case CompressedInstructionFormatCL:
	case CompressedInstructionFormatCS:
	case CompressedInstructionFormatCB:
	case CompressedInstructionFormatCJ:
	}

	return 0, ExcpIllegalInstruction
}

// DecompressCR decompresses the CR format compressed instruction using the given parts.
func (cpu *CPU) DecompressCR(op, rs2, rdOrRs1, funct4 uint64) (uint64, Exception) {
	switch op {
	case 0b00:
		return 0, ExcpIllegalInstruction
	case 0b01:
		switch rs2 >> 3 {
		case 0b00:
			// c.sub
			// -> sub rd, rd, rs2 while rd = 8 + rd', rs2 = 8 + rs2'
			rdOrRs1, rs2 = rdOrRs1&0b111, rs2&0b111
			rdOrRs1 += 8
			rs2 += 8
			sub := uint64(0b0100000_00000_00000_000_00000_0110011)
			return sub | (rs2 << 20) | (rdOrRs1 << 15) | (rdOrRs1 << 7), ExcpNone
		case 0b01:
			// c.xor
			// -> xor rd, rd, rs2 while rd = 8 + rd', rs2 = 8 + rs2'
			rdOrRs1, rs2 = rdOrRs1&0b111, rs2&0b111
			rdOrRs1 += 8
			rs2 += 8
			xor := uint64(0b0000000_00000_00000_100_00000_0110011)
			return xor | (rs2 << 20) | (rdOrRs1 << 15) | (rdOrRs1 << 7), ExcpNone
		case 0b10:
			// c.or
			// -> or rd, rd, rs2 while rd = 8 + rd', rs2 = 8 + rs2'
			rdOrRs1, rs2 = rdOrRs1&0b111, rs2&0b111
			rdOrRs1 += 8
			rs2 += 8
			or := uint64(0b0000000_00000_00000_110_00000_0110011)
			return or | (rs2 << 20) | (rdOrRs1 << 15) | (rdOrRs1 << 7), ExcpNone
		case 0b11:
			// c.and
			// -> and rd, rd, rs2 while rd = 8 + rd', rs2 = 8 + rs2'
			rdOrRs1, rs2 = rdOrRs1&0b111, rs2&0b111
			rdOrRs1 += 8
			rs2 += 8
			and := uint64(0b0000000_00000_00000_111_00000_0110011)
			return and | (rs2 << 20) | (rdOrRs1 << 15) | (rdOrRs1 << 7), ExcpNone
		default:
			return 0, ExcpIllegalInstruction
		}
	case 0b10:
		switch funct4 & 0b1 {
		case 0b0:
			// c.mv
			// -> add rd, x0, rs2. Illegal if rs2 = x0.
			if rs2 == 0 {
				return 0, ExcpIllegalInstruction
			}
			add := uint64(0b0000000_00000_00000_000_00000_0110011)
			return add | (rs2 << 20) | (rdOrRs1 << 7), ExcpNone
		case 0b1:
			// c.add
			// -> add rd, rd, rs2. Illegal if rd = x0 || rs2 = x0.
			if rdOrRs1 == 0 || rs2 == 0 {
				return 0, ExcpIllegalInstruction
			}
			add := uint64(0b0000000_00000_00000_000_00000_0110011)
			return add | (rs2 << 20) | (rdOrRs1 << 15) | (rdOrRs1 << 7), ExcpNone
		default:
			return 0, ExcpIllegalInstruction

		}
	default:
		return 0, ExcpIllegalInstruction
	}

	return 0, ExcpIllegalInstruction
}

// DecompressCI decompresses the CR format compressed instruction using the given parts.
func (cpu *CPU) DecompressCI(op, imm1, rdOrRs1, imm2, funct3 uint64) (uint64, Exception) {
	switch op {
	case 0b00:
		return 0, ExcpIllegalInstruction
	case 0b01:
		switch funct3 {
		case 0b000:
			switch rdOrRs1 {
			case 0b0:
				// c.nop
				// -> addi x0, x0, 0 (do nothing)
				return 0b000000000000_00000_000_00000_0010011, ExcpNone
			default:
				// c.addi
				// -> addi rd, rd, nzimm[5:0] while imm1 = nzimm[4:0], imm2 = nzimm[5]
				addi := uint64(0b000000000000_00000_000_00000_0010011)
				var mask uint32 = 0b0
				if imm2 == 0b1 {
					// Sign-extend. if imm2 is 1, imm[31:6] should be 1.
					// Also, set imm[5](=imm2) 1 here.
					mask = 0b1111_1111_1111_1111_1111_1111_1110_0000
				}
				nzimm := imm1 | uint64(int64(int32(mask)))
				return addi | (rdOrRs1 << 15) | (rdOrRs1 << 7) | (nzimm << 20), ExcpNone
			}
		case 0b010:
			// c.li
			// -> addi rd, x0, imm
			addi := uint64(0b000000000000_00000_000_00000_0010011)
			var mask uint32 = 0b0
			if imm2 == 0b1 {
				// Sign-extend. if imm2 is 1, imm[31:6] should be 1.
				// Also, set imm[5](=imm2) 1 here.
				mask = 0b1111_1111_1111_1111_1111_1111_1110_0000
			}
			nzimm := imm1 | uint64(int64(int32(mask)))
			return addi | (rdOrRs1 << 7) | (nzimm << 20), ExcpNone
		case 0b011:
			switch rdOrRs1 {
			case 0b10:
				// c.addi16sp
				// -> addi x2, x2, imm. Illegal if imm is 0.
				addi := uint64(0b000000000000_00010_000_00010_0010011)
				var mask uint32 = 0b0
				if imm2 == 0b1 {
					// Sign-extend. if imm2 is 1, imm[31:10] should be 1.
					// Also, set imm[9](=imm2) 1 here.
					mask = 0b1111_1111_1111_1111_1111_1110_0000_0000
				}
				imm := uint64(int64(int32(mask))) |
					(imm1 & 0b1_0000) | // imm1[4] -> imm[4]
					((imm1 << 3) & 0b100_0000) | // imm1[3] -> imm[6]
					((imm1 << 6) & 0b1_0000_0000) | // imm1[2] -> imm[8]
					((imm1 << 6) & 0b1000_0000) | // imm1[1] -> imm[7]
					((imm1 << 5) & 0b10_0000) // imm1[0] -> imm[5]
				return addi | (imm << 20), ExcpNone
			case 0b00:
				// must not be 0
				return 0, ExcpIllegalInstruction
			default:
				// c.lui
				// -> lui rd, imm. Illegal if rd = x2 || imm = 0
				if rdOrRs1 == 0b10 || rdOrRs1 == 0b0 {
					return 0, ExcpIllegalInstruction
				}

				if imm1 == 0b0 && imm2 == 0b0 {
					return 0, ExcpIllegalInstruction
				}

				var mask uint64 = 0b0
				if imm2 == 0b1 {
					// Sign-extend. if imm2 is 1, imm[31:18] should be 1.
					// Also, set imm[17](=imm2) 1 here.
					mask = 0b1111_1111_1111_1110_0000_0000_0000_0000
				}
				imm := uint64(int64(int32(mask))) | (imm1 << 12) // imm1 -> imm[16:12]
				lui := uint64(0b00000000000000000000_00000_0110111)
				return lui | (rdOrRs1 << 7) | (imm << 12), ExcpNone
			}
		case 0b100:
			switch rdOrRs1 >> 3 {
			case 0b00:
				// c.srli
				// -> srli rd, rd, uimm while rd = 8 + rd'
				rdOrRs1 = rdOrRs1 & 0b111 // left 2-bit of rd is opcode
				rdOrRs1 += 8
				uimm := imm1 |
					(imm2 << 5) // imm2 -> uimm[5]
				srli := uint64(0b000000_000000_00000_101_00000_0010011)
				return srli | (rdOrRs1 << 7) | (rdOrRs1 << 15) | (uimm << 20), ExcpNone
			case 0b01:
				// c.srai
				// -> srai rd, rd, uimm while rd = 8 + rd'
				rdOrRs1 = rdOrRs1 & 0b111 // left 2-bit of rd is opcode
				rdOrRs1 += 8
				uimm := imm1 |
					(imm2 << 5) // imm2 -> uimm[5]
				srai := uint64(0b010000_000000_00000_101_00000_0010011)
				return srai | (rdOrRs1 << 7) | (rdOrRs1 << 15) | (uimm << 20), ExcpNone
			case 0b10:
				// c.andi
			default:
				return 0, ExcpIllegalInstruction
			}
		default:
			return 0, ExcpIllegalInstruction
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
	default:
		return 0, ExcpIllegalInstruction
	}

	return 0, ExcpIllegalInstruction
}

// DecodeCompressedInstructionFormat decodes the given compressed instruction and returns its format.
func (cpu *CPU) DecodeCompressedInstructionFormat(compressed uint64) CompressedInstructionFormat {
	opcode := compressed & 0b11
	funct3 := (compressed >> 13) & 0b111
	switch opcode {
	case 0b00:
		switch funct3 {
		case 0b000:
			return CompressedInstructionFormatCIW
		case 0b001, 0b010, 0b011, 0b101, 0b110, 0b111:
			return CompressedInstructionFormatCL
		default:
			return CompressedInstructionFormatInvalid
		}
	case 0b01:
		switch funct3 {
		case 0b000, 0b010, 0b011:
			return CompressedInstructionFormatCI
		case 0b100:
			f := (compressed >> 10) & 0b11
			switch f {
			case 0b00, 0b01, 0b10:
				return CompressedInstructionFormatCI
			case 0b11:
				return CompressedInstructionFormatCR
			default:
				return CompressedInstructionFormatInvalid
			}
		case 0b001, 0b101:
			return CompressedInstructionFormatCJ
		case 0b110, 0b111:
			return CompressedInstructionFormatCB
		default:
			return CompressedInstructionFormatInvalid
		}
	case 0b10:
		switch funct3 {
		case 0b000:
			return CompressedInstructionFormatCI
		case 0b100:
			f1 := (compressed >> 12) & 0b1
			f2 := (compressed >> 2) & 0b1_1111
			f3 := (compressed >> 7) & 0b1_1111
			switch f1 {
			case 0b0:
				if f2 == 0 {
					return CompressedInstructionFormatCJ
				} else {
					return CompressedInstructionFormatCR
				}
			case 0b1:
				if f2 == 0 && f3 == 0 {
					return CompressedInstructionFormatCI
				} else if f2 == 0 {
					return CompressedInstructionFormatCJ
				} else {
					return CompressedInstructionFormatCR
				}
			}
		case 0b001, 0b010, 0b011, 0b101, 0b110, 0b111:
			return CompressedInstructionFormatCSS
		default:
			return CompressedInstructionFormatInvalid
		}
	default:
		return CompressedInstructionFormatInvalid
	}

	return CompressedInstructionFormatInvalid
}
