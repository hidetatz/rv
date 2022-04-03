package main

import "fmt"

type ELFFile struct {
	Header   *ELFHeader
	Sections []*Section
	Programs []*Program

	// ToHost is a rv specific information.
	// In riscv-tests, tohost is used to tell the emulator that the
	// test program ends at the address.
	// If ToHost is not 0, the program is likely the test program from riscv-tests.
	// https://riscv.org/wp-content/uploads/2015/01/riscv-testing-frameworks-bootcamp-jan2015.pdf
	ToHost uint64
}

type ELFHeader struct {
	Class      uint8
	Data       uint8
	ELFVersion uint8
	OSABI      uint8
	ABIVersion uint8
	Type       uint16
	Machine    uint16
	Version    uint32
	Entry      uint64
	PhOff      uint64
	ShOff      uint64
	Flags      uint32
	EhSize     uint16
	PhEntSize  uint16
	PhNum      uint16
	ShEntSize  uint16
	ShNum      uint16
	ShStrNdx   uint16
}

type Section struct {
	Name      uint32
	Type      uint32
	Flags     uint64
	Addr      uint64
	Offset    uint64
	Size      uint64
	Link      uint32
	Info      uint32
	AddrAlign uint64
	EntSize   uint64
}

type Program struct {
	Type   uint32
	Flags  uint32
	Offset uint64
	VAddr  uint64
	PAddr  uint64
	Filesz uint64
	Memsz  uint64
	Align  uint64
}

func LoadELF(data []uint8) (*ELFFile, error) {
	if len(data) < 4 {
		return nil, fmt.Errorf("too short, the file seems not to be a valid ELF")
	}

	if data[0] != 0x7f || data[1] != 'E' || data[2] != 'L' || data[3] != 'F' {
		return nil, fmt.Errorf("invalid magic number, the file seems not to be a valid ELF")
	}

	f := &ELFFile{
		Header: &ELFHeader{},
	}

	/*
	 * 1. Load ELFHeader
	 */
	f.Header.Class = read8(data, 4)
	f.Header.Data = read8(data, 5)
	f.Header.ELFVersion = read8(data, 6)
	f.Header.OSABI = read8(data, 7)
	f.Header.ABIVersion = read8(data, 8)
	f.Header.Type = read16(data, 16)
	f.Header.Machine = read16(data, 18)
	f.Header.Version = read32(data, 20)

	offset := uint64(24)

	switch f.Header.Class {
	case 1: // 32-bit
		f.Header.Entry = uint64(read32(data, offset))
		offset += 4
		f.Header.PhOff = uint64(read32(data, offset))
		offset += 4
		f.Header.ShOff = uint64(read32(data, offset))
		offset += 4
	case 2: // 64-bit
		f.Header.Entry = read64(data, offset)
		offset += 8
		f.Header.PhOff = read64(data, offset)
		offset += 8
		f.Header.ShOff = read64(data, offset)
		offset += 8
	}

	f.Header.Flags = read32(data, offset)
	offset += 4

	f.Header.EhSize = read16(data, offset)
	offset += 2

	f.Header.PhEntSize = read16(data, offset)
	offset += 2
	f.Header.PhNum = read16(data, offset)
	offset += 2
	f.Header.ShEntSize = read16(data, offset)
	offset += 2
	f.Header.ShNum = read16(data, offset)
	offset += 2
	f.Header.ShStrNdx = read16(data, offset)

	/*
	 * 2. Load Section Headers
	 */
	offset = f.Header.ShOff
	for i := 0; i < int(f.Header.ShNum); i++ {
		s := &Section{}
		s.Name = read32(data, offset)
		offset += 4
		s.Type = read32(data, offset)
		offset += 4

		switch f.Header.Class {
		case 1: // 32-bit
			s.Flags = uint64(read32(data, offset))
			offset += 4
			s.Addr = uint64(read32(data, offset))
			offset += 4
			s.Offset = uint64(read32(data, offset))
			offset += 4
			s.Size = uint64(read32(data, offset))
			offset += 4
		case 2: // 64-bit
			s.Flags = read64(data, offset)
			offset += 8
			s.Addr = read64(data, offset)
			offset += 8
			s.Offset = read64(data, offset)
			offset += 8
			s.Size = read64(data, offset)
			offset += 8
		}

		s.Link = read32(data, offset)
		offset += 4
		s.Info = read32(data, offset)
		offset += 4

		switch f.Header.Class {
		case 1: // 32-bit
			s.AddrAlign = uint64(read32(data, offset))
			offset += 4
			s.EntSize = uint64(read32(data, offset))
			offset += 4
		case 2: // 64-bit
			s.AddrAlign = read64(data, offset)
			offset += 8
			s.EntSize = read64(data, offset)
			offset += 8
		}

		f.Sections = append(f.Sections, s)
	}

	/*
	 * 3. Load Program Headers
	 */
	offset = f.Header.PhOff
	for i := 0; i < int(f.Header.PhNum); i++ {
		p := &Program{}
		p.Type = read32(data, offset)
		offset += 4

		if f.Header.Class == 2 {
			// if 64-bit, must read flags here.
			// flags place is not the same in 32-bit and 64-bit.
			// https://docs.oracle.com/cd/E19683-01/816-1386/chapter6-83432/index.html
			p.Flags = read32(data, offset)
			offset += 4
		}

		switch f.Header.Class {
		case 1: // 32-bit
			p.Offset = uint64(read32(data, offset))
			offset += 4
			p.VAddr = uint64(read32(data, offset))
			offset += 4
			p.PAddr = uint64(read32(data, offset))
			offset += 4
			p.Filesz = uint64(read32(data, offset))
			offset += 4
			p.Memsz = uint64(read32(data, offset))
			offset += 4
		case 2: // 64-bit
			p.Offset = read64(data, offset)
			offset += 8
			p.VAddr = read64(data, offset)
			offset += 8
			p.PAddr = read64(data, offset)
			offset += 8
			p.Filesz = read64(data, offset)
			offset += 8
			p.Memsz = read64(data, offset)
			offset += 8
		}

		if f.Header.Class == 1 {
			// if 32-bit, must read flags here.
			p.Flags = read32(data, offset)
			offset += 4
		}

		switch f.Header.Class {
		case 1: // 32-bit
			p.Align = uint64(read32(data, offset))
			offset += 4
		case 2: // 64-bit
			p.Align = read64(data, offset)
			offset += 8
		}

		f.Programs = append(f.Programs, p)
	}

	/*
	 * 4. Try to find ToHost Address
	 *    This is not necessarily required in just loading ELF,
	 *    but ToHost is used when running riscv-tests (https://github.com/riscv-software-src/riscv-tests) code.
	 */
	progDataSectionHeaders := []*Section{}
	stringTableSectionHeaders := []*Section{}
	for _, s := range f.Sections {
		if s.Type == 1 {
			progDataSectionHeaders = append(progDataSectionHeaders, s)
		} else if s.Type == 3 {
			stringTableSectionHeaders = append(stringTableSectionHeaders, s)
		}
	}

	tohost := []uint8{0x2e, 0x74, 0x6f, 0x68, 0x6f, 0x73, 0x74, 0x00} // ".tohost\null"
	for _, pds := range progDataSectionHeaders {
		addr := pds.Addr
		name := pds.Name
		for _, sts := range stringTableSectionHeaders {
			offset := sts.Offset
			size := sts.Size
			found := true
			for i := range tohost {
				a := offset + uint64(name) + uint64(i)
				if a >= offset+size || read8(data, a) != tohost[i] {
					found = false
					break
				}
			}

			if found {
				f.ToHost = addr
			}
		}
	}

	return f, nil
}

func read8(data []uint8, addr uint64) uint8 {
	return data[addr]
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
