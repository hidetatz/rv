package main

import "fmt"

// LoadELF parses the given data as valid ELF.
// When the ELF is loaded correctly, the program will be loaded on memory and
// the program counter is returned.
func LoadElf(data []uint8, memory *Memory) (uint64, error) {
	if len(data) < 4 {
		return 0, fmt.Errorf("too short, the file seems not to be a valid ELF")
	}

	if data[0] != 0x7f || data[1] != 'E' || data[2] != 'L' || data[3] != 'F' {
		return 0, fmt.Errorf("invalid magic number, the file seems not to be a valid ELF")
	}

	// Read necessary data from ELF header. Unused data are not read.
	// Assuming architecture is 64-bit.
	entryPoint := read64(data, 0x18)
	shOff := read64(data, 0x28)
	shNum := read16(data, 0x3c)

	offset := shOff
	for i := 0; i < int(shNum); i++ {
		shType := read32(data, offset+4)
		// SHT_PROGBITS
		if shType == 1 {
			shAddr := read64(data, offset+16)
			shOffset := read64(data, offset+24)
			shSize := read64(data, offset+32)

			fmt.Printf("[debug] shAddr: %x, shOffset: %x, shSize: %x\n", shAddr, shOffset, shSize)

			if shAddr >= 0x8000_0000 && shOffset > 0 && shSize > 0 {
				for j := 0; j < int(shSize); j++ {
					memory.Write(shAddr+uint64(j), uint64(data[shOffset+uint64(j)]), 8)
					fmt.Printf("[debug] writing %08b at %x\n", data[shOffset+uint64(j)], shAddr+uint64(j))
				}
			}
		}

		offset += 64

	}

	return entryPoint, nil
}

func read16(data []uint8, addr uint64) uint16 {
	return uint16(data[addr]) | uint16(data[addr+1])<<8
}

func read32(data []uint8, addr uint64) uint32 {
	return uint32(data[addr]) |
		uint32(data[addr+1])<<8 |
		uint32(data[addr+2])<<16 |
		uint32(data[addr+3])<<24
}

func read64(data []uint8, addr uint64) uint64 {
	return uint64(data[addr]) |
		uint64(data[addr+1])<<8 |
		uint64(data[addr+2])<<16 |
		uint64(data[addr+3])<<24 |
		uint64(data[addr+4])<<32 |
		uint64(data[addr+5])<<40 |
		uint64(data[addr+6])<<48 |
		uint64(data[addr+7])<<56
}
