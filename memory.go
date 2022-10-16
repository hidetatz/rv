package main

const DramBase = 0x0

type Memory struct {
	// 3GiB
	Mem [3 * 1024 * 1024 * 1024]uint8
}

func NewMemory() *Memory {
	return &Memory{Mem: [3 * 1024 * 1024 * 1024]uint8{}}
}

func (mem *Memory) Read(addr uint64, size int) uint64 {
	index := addr - DramBase
	switch size {
	case byt:
		return uint64(mem.Mem[index])
	case halfword:
		return uint64(mem.Mem[index]) | uint64(mem.Mem[index+1])<<8
	case word:
		return uint64(mem.Mem[index]) |
			uint64(mem.Mem[index+1])<<8 |
			uint64(mem.Mem[index+2])<<16 |
			uint64(mem.Mem[index+3])<<24
	case doubleword:
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

func (mem *Memory) Write(addr, val uint64, size int) {
	index := addr - DramBase
	switch size {
	case byt:
		mem.Mem[index] = uint8(val)
	case halfword:
		mem.Mem[index] = uint8(val & 0b1111_1111)
		mem.Mem[index+1] = uint8((val >> 8) & 0b1111_1111)
	case word:
		mem.Mem[index] = uint8(val & 0b1111_1111)
		mem.Mem[index+1] = uint8((val >> 8) & 0b1111_1111)
		mem.Mem[index+2] = uint8((val >> 16) & 0b1111_1111)
		mem.Mem[index+3] = uint8((val >> 24) & 0b1111_1111)
	case doubleword:
		mem.Mem[index] = uint8(val & 0b1111_1111)
		mem.Mem[index+1] = uint8((val >> 8) & 0b1111_1111)
		mem.Mem[index+2] = uint8((val >> 16) & 0b1111_1111)
		mem.Mem[index+3] = uint8((val >> 24) & 0b1111_1111)
		mem.Mem[index+4] = uint8((val >> 32) & 0b1111_1111)
		mem.Mem[index+5] = uint8((val >> 40) & 0b1111_1111)
		mem.Mem[index+6] = uint8((val >> 48) & 0b1111_1111)
		mem.Mem[index+7] = uint8((val >> 56) & 0b1111_1111)
	}
}
