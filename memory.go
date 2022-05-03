package main

// DramBase is a base address of actual DRAM space in the whole Memory.
// Before DramBase, there should be some memory-mapped IO devices, etc.
//const DramBase = 0x8000_0000
const DramBase = 0x0

// Memory is a DRAM emulator.
type Memory struct {
	// 4GiB
	Mem [4 * 1024 * 1024 * 1024]uint8
}

func NewMemory() *Memory {
	return &Memory{Mem: [4 * 1024 * 1024 * 1024]uint8{}}
}

func (mem *Memory) Set(data []uint8) {
	copy(mem.Mem[:], data)
}

func (mem *Memory) Read(addr uint64, size Size) uint64 {
	index := addr - DramBase
	switch size {
	case Byte:
		// Read and return 1 bit as uint64
		return uint64(mem.Mem[index])
	case HalfWord:
		// Read and return 2 bits. Byte order is Little Endian.
		// Read every byte and combine them into 1 uint64 by Or operator.
		return uint64(mem.Mem[index]) | uint64(mem.Mem[index+1])<<8
	case Word:
		// Read and return 4 bits. The same as the case above.
		return uint64(mem.Mem[index]) |
			uint64(mem.Mem[index+1])<<8 |
			uint64(mem.Mem[index+2])<<16 |
			uint64(mem.Mem[index+3])<<24
	case DoubleWord:
		// Read and return 8 bits. The same as the case above.
		return uint64(mem.Mem[index]) |
			uint64(mem.Mem[index+1])<<8 |
			uint64(mem.Mem[index+2])<<16 |
			uint64(mem.Mem[index+3])<<24 |
			uint64(mem.Mem[index+4])<<32 |
			uint64(mem.Mem[index+5])<<40 |
			uint64(mem.Mem[index+6])<<48 |
			uint64(mem.Mem[index+7])<<56
	}

	// TODO: Should throw LoadAccessFault exception
	return 0
}

func (mem *Memory) Write(addr, val uint64, size Size) {
	index := addr - DramBase
	switch size {
	case Byte:
		mem.Mem[index] = uint8(val)
	case HalfWord:
		mem.Mem[index] = uint8(val & 0x1111_1111)
		mem.Mem[index+1] = uint8((val >> 8) & 0x1111_1111)
	case Word:
		mem.Mem[index] = uint8(val & 0x1111_1111)
		mem.Mem[index+1] = uint8((val >> 8) & 0x1111_1111)
		mem.Mem[index+2] = uint8((val >> 16) & 0x1111_1111)
		mem.Mem[index+3] = uint8((val >> 24) & 0x1111_1111)
	case DoubleWord:
		mem.Mem[index] = uint8(val & 0x1111_1111)
		mem.Mem[index+1] = uint8((val >> 8) & 0x1111_1111)
		mem.Mem[index+2] = uint8((val >> 16) & 0x1111_1111)
		mem.Mem[index+3] = uint8((val >> 24) & 0x1111_1111)
		mem.Mem[index+4] = uint8((val >> 32) & 0x1111_1111)
		mem.Mem[index+5] = uint8((val >> 40) & 0x1111_1111)
		mem.Mem[index+6] = uint8((val >> 48) & 0x1111_1111)
		mem.Mem[index+7] = uint8((val >> 56) & 0x1111_1111)
	}
}
