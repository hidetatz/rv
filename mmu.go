package main

type MemoryAccessType int

const (
	MemoryAccessTypeInstruction MemoryAccessType = iota + 1
	MemoryAccessTypeLoad
	MemoryAccessTypeStore
)

type AddressingMode int

const (
	AddressingModeNone AddressingMode = iota
	AddressingModeSV32
	AddressingModeSV39
)

// MMU emulates memory management unit in a processor.
type MMU struct {
	Bus            *Bus
	XLen           XLen
	AddressingMode AddressingMode
	Mstatus        uint64
	PPN            uint64
}

func NewMMU(xlen XLen) *MMU {
	return &MMU{
		Bus:            NewBus(),
		XLen:           xlen,
		AddressingMode: AddressingModeNone,
		PPN:            0,
	}
}

func (mmu *MMU) Fetch(vAddr uint64, size Size, curMode Mode) (uint64, *Exception) {
	pAddr, excp := mmu.Translate(vAddr, MemoryAccessTypeInstruction, curMode)
	if excp.Code != ExcpCodeNone {
		return 0, excp
	}

	return mmu.Bus.Read(pAddr, size), ExcpNone()
}

func (mmu *MMU) Read(vAddr uint64, size Size, curMode Mode) (uint64, *Exception) {
	pAddr, excp := mmu.Translate(vAddr, MemoryAccessTypeLoad, curMode)
	if excp.Code != ExcpCodeNone {
		return 0, excp
	}

	return mmu.Bus.Read(pAddr, size), ExcpNone()
}

func (mmu *MMU) Write(vAddr, val uint64, size Size, curMode Mode) *Exception {
	pAddr, excp := mmu.Translate(vAddr, MemoryAccessTypeStore, curMode)
	if excp.Code != ExcpCodeNone {
		return excp
	}

	mmu.Bus.Write(pAddr, val, size)
	return ExcpNone()
}

func (mmu *MMU) Translate(vAddr uint64, at MemoryAccessType, curMode Mode) (uint64, *Exception) {
	vAddr = mmu.getEffectiveAddress(vAddr)
	switch mmu.AddressingMode {
	case AddressingModeNone:
		return vAddr, ExcpNone()
	case AddressingModeSV32:
		switch curMode {
		case Machine:
			if at == MemoryAccessTypeInstruction {
				return vAddr, ExcpNone()
			}

			if ((mmu.Mstatus >> 17) & 1) == 0 {
				return vAddr, ExcpNone()
			}

			newPrivMode := (mmu.Mstatus >> 9) & 3
			switch newPrivMode {
			case 3: // Machine
				return vAddr, ExcpNone()
			default:
				return mmu.Translate(vAddr, at, CodeToMode(int(newPrivMode)))
			}
		default:
			vpns := []uint64{
				(vAddr >> 12) & 0x3ff,
				(vAddr >> 22) & 0x3ff,
			}
			return mmu.TraversePage(vAddr, 2-1, mmu.PPN, vpns, at)
		}
	case AddressingModeSV39:
		switch curMode {
		case Machine:
			if at == MemoryAccessTypeInstruction {
				return vAddr, ExcpNone()
			}

			if ((mmu.Mstatus >> 17) & 1) == 0 {
				return vAddr, ExcpNone()
			}

			newPrivMode := (mmu.Mstatus >> 9) & 3
			switch newPrivMode {
			case 3:
				return vAddr, ExcpNone()
			default:
				return mmu.Translate(vAddr, at, CodeToMode(int(newPrivMode)))
			}
		default:
			vpns := []uint64{
				(vAddr >> 12) & 0x1ff,
				(vAddr >> 21) & 0x1ff,
				(vAddr >> 30) & 0x1ff,
			}
			return mmu.TraversePage(vAddr, 3-1, mmu.PPN, vpns, at)
		}
	default:
		panic("should not come here")
	}
}

func (mmu *MMU) TraversePage(vAddr uint64, level int, parentPPN uint64, vpns []uint64, at MemoryAccessType) (uint64, *Exception) {
	fault := func() *Exception {
		switch at {
		case MemoryAccessTypeInstruction:
			return ExcpInstructionPageFault(vAddr)
		case MemoryAccessTypeLoad:
			return ExcpLoadPageFault(vAddr)
		case MemoryAccessTypeStore:
			return ExcpStoreAMOPageFault(vAddr)
		}

		return ExcpNone() // should not come here
	}

	pageSize := 4096

	pteSize := 8
	if mmu.AddressingMode == AddressingModeSV32 {
		pteSize = 4
	}

	pteAddr := parentPPN*uint64(pageSize) + vpns[level]*uint64(pteSize)

	var pte uint64
	if mmu.AddressingMode == AddressingModeSV32 {
		pte = mmu.Bus.Read(pteAddr, Word)
	} else {
		pte = mmu.Bus.Read(pteAddr, DoubleWord)
	}

	var ppn uint64
	var ppns []uint64
	if mmu.AddressingMode == AddressingModeSV32 {
		ppn = (pte >> 10) & 0x3fffff
		ppns = []uint64{
			(pte >> 10) & 0x3ff,
			(pte >> 20) & 0xfff,
			0,
		}
	} else {
		ppn = (pte >> 10) & 0xfffffffffff
		ppns = []uint64{
			(pte >> 10) & 0x1ff,
			(pte >> 19) & 0x1ff,
			(pte >> 28) & 0x3ffffff,
		}
	}

	d := (pte >> 7) & 1
	a := (pte >> 6) & 1
	x := (pte >> 3) & 1
	w := (pte >> 2) & 1
	r := (pte >> 1) & 1
	v := pte & 1

	if v == 0 || (r == 0 && w == 1) {
		return 0, fault()
	}

	if r == 0 && x == 0 {
		if level == 0 {
			return 0, fault()
		}

		return mmu.TraversePage(vAddr, level-1, ppn, vpns, at)
	}

	// page found

	b := false
	if at == MemoryAccessTypeStore {
		b = d == 0
	}

	if a == 0 || b {
		newPTE := pte | (1 << 6)
		if at == MemoryAccessTypeStore {
			newPTE |= (1 << 7)
		}

		if mmu.AddressingMode == AddressingModeSV32 {
			mmu.Bus.Write(pteAddr, newPTE, Word)
		} else {
			mmu.Bus.Write(pteAddr, newPTE, DoubleWord)
		}
	}

	switch at {
	case MemoryAccessTypeInstruction:
		if x == 0 {
			return 0, fault()
		}
	case MemoryAccessTypeLoad:
		if r == 0 {
			return 0, fault()
		}
	case MemoryAccessTypeStore:
		if w == 0 {
			return 0, fault()
		}
	}

	offset := vAddr & 0xfff
	switch mmu.AddressingMode {
	case AddressingModeSV32:
		switch level {
		case 1:
			if ppns[0] != 0 {
				return 0, fault()
			}

			return ppns[1]<<22 | vpns[0]<<12 | offset, ExcpNone()
		case 0:
			return (ppn << 12) | offset, ExcpNone()
		default:
			panic("invalid level") // should not come here
		}
	default:
		switch level {
		case 2:
			if ppns[1] != 0 || ppns[0] != 0 {
				return 0, fault()
			}

			return (ppns[2] << 30) | (vpns[1] << 21) | (vpns[0] << 12) | offset, ExcpNone()
		case 1:
			if ppns[0] != 0 {
				return 0, fault()
			}

			return (ppns[2] << 30) | (ppns[1] << 21) | (vpns[0] << 12) | offset, ExcpNone()
		case 0:
			return (ppn << 12) | offset, ExcpNone()
		default:
			panic("invalid level") // should not come here
		}
	}
}

func (mmu *MMU) getEffectiveAddress(vAddr uint64) uint64 {
	if mmu.XLen == XLen64 {
		return vAddr
	}

	return vAddr & 0xffff_ffff
}
