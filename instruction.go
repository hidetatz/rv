package rv

// Instruction is a general-purpose 32-bit instruction representation.
// This is to hold the decoded raw instruction binary string.
// Some fields such as Rd, Funct3, Rs1... are called other name like imm or shamt in some instructions,
// but the most common and widely used names are chosen.
type Instruction struct {
	Raw    uint32 // raw instruction
	Op     uint32 // operation
	Opcode uint8  // 7-bit (from 0 to 6).
	Rd     uint8  // 5-bit (from 7 to 11).
	Funct3 uint8  // 3-bit (from 12 to 14).
	Rs1    uint8  // 5-bit (from 15 to 19).
	Rs2    uint8  // 5-bit (from 20 to 24).
	Funct7 uint8  // 7-bit (from 25 to 31).
}

const (
	// R
	InsAdd = iota + 1
	InsSub
	InsSLL
	InsSLT
	InsSLTU
	InsXOR
	InsSRL
	InsSRA
	InsOr
	InsAnd
)
