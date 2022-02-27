package rv

// Memory is a DRAM emulator.
type Memory struct {
	// 1MiB
	Mem [1024 * 1024]uint8
}

func NewMemory() *Memory {
	return &Memory{Mem: [1024 * 1024]uint8{}}
}

func (mem *Memory) Read(addr uint64, size uint8) uint64 {
	switch size {
	case 8:
		// Read and return 1 bit as uint64
		return uint64(mem.Mem[addr])
	case 16:
		// Read and return 2 bits. Byte order is Little Endian.
		// Read every byte and combine them into 1 uint64 by Or operator.
		return uint64(mem.Mem[addr]) | uint64(mem.Mem[addr+1])<<8
	case 32:
		// Read and return 4 bits. The same as the case above.
		return uint64(mem.Mem[addr]) |
			uint64(mem.Mem[addr+1])<<8 |
			uint64(mem.Mem[addr+2])<<16 |
			uint64(mem.Mem[addr+3])<<24
	case 64:
		// Read and return 8 bits. The same as the case above.
		return uint64(mem.Mem[addr]) |
			uint64(mem.Mem[addr+1])<<8 |
			uint64(mem.Mem[addr+2])<<16 |
			uint64(mem.Mem[addr+3])<<24 |
			uint64(mem.Mem[addr+4])<<32 |
			uint64(mem.Mem[addr+5])<<40 |
			uint64(mem.Mem[addr+6])<<48 |
			uint64(mem.Mem[addr+7])<<56
	}

	// TODO: Should throw LoadAccessFault exception
	return 0
}
