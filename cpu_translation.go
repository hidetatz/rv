package main

import "fmt"

type MemoryAccessType int

const (
	MemoryAccessTypeInstruction MemoryAccessType = iota + 1
	MemoryAccessTypeLoad
	MemoryAccessTypeStore
)

const (
	// Page size for SV39.
	PageSize = 4096
	// Level for SV39.
	Levels = 3
	// PTE Size for SV39.
	PTESize = 8
)

// TranslateMem translates physical <-> virtual memory address based on SV39.
func (cpu *CPU) TranslateMem(virtualAddr uint64, at MemoryAccessType) (uint64, *Exception) {
	fault := func() *Exception {
		switch at {
		case MemoryAccessTypeInstruction:
			return ExcpInstructionPageFault(virtualAddr)
		case MemoryAccessTypeLoad:
			return ExcpLoadPageFault(virtualAddr)
		case MemoryAccessTypeStore:
			return ExcpStoreAMOPageFault(virtualAddr)
		}

		return ExcpNone() // should not come here
	}

	// SATP register must be active which means the privilege mode must be S or U.
	if cpu.Mode == Machine {
		return virtualAddr, ExcpNone()
	}

	if !cpu.PagingEnabled {
		return virtualAddr, ExcpNone()
	}

	// First, virtual Address is partitioned into a virtual page number (VPN) and offset.
	vpn := []uint64{
		bits(virtualAddr, 20, 12),
		bits(virtualAddr, 29, 21),
		bits(virtualAddr, 38, 30),
	}

	// a = SATP.PPN * PAGE_SIZE, i = LEVELS - 1.
	satp := cpu.CSR.Read(CsrSATP)
	a := bits(satp, 43, 0) * PageSize
	i := Levels - 1
	var pte uint64

	for {
		// pte = the value of the PTE at address a + vpn[i] * PTESize.
		pte = cpu.Bus.Read(a+vpn[i]*PTESize, DoubleWord)

		// PMA/PMP check is skipped.
		// TODO: Do PMA/PMP check.

		// Validate PTE
		pteV := bit(pte, 0)
		pteR := bit(pte, 1)
		pteW := bit(pte, 2)

		fmt.Printf("%b\n", pte)
		if pteV == 0 || (pteR == 0 && pteW == 1) {
			// "Any bits or encodings that are reserved for future standard use are set within pte" check is skipped.
			// TODO: Do the check.

			return 0, fault()

		}

		// PTE must be valid.

		pteX := bit(pte, 3)
		if pteR == 1 || pteX == 1 {
			// if pte.r = 1 or pte.w = 1, the PTE is the last entry. Go to next step.
			break
		}

		// otherwise, go to next PTE.
		i--
		if i < 0 {
			return 0, fault()
		}

		a = bits(pte, 53, 10) * PageSize
	}

	// Leaf PTE access permission check is skipped.
	// TODO: Do the check.

	ppn := []uint64{bits(pte, 18, 10), bits(pte, 27, 19), bits(pte, 53, 28)}
	if i > 0 {
		idx := i - 1
		for idx >= 0 {
			if ppn[idx] != 0 {
				// if i > 0 and ppn[i-1:0] != 0, this is a misalighed superpage.
				return 0, fault()
			}

			idx--
		}
	}

	pteA := bit(pte, 6)
	pteD := bit(pte, 7)
	if pteA == 0 || (at == MemoryAccessTypeStore && pteD == 0) {
		// PMA/PMP check is skipped.
		// TODO: Do the check.

		// set a and d in pte
		pte = setBit(pte, 6)
		if at == MemoryAccessTypeStore {
			pte = setBit(pte, 7)
		}
	}

	// Translation success.
	// * pa.offset == va.offset
	// * if i > 0, then this is a superpage translation. pa.ppn[i-1:0] = va.vpn[i-1:0].
	// * pa.ppn[LEVELS-1:i] = pte.ppn[LEVELS-1:i].

	offset := bits(virtualAddr, 11, 0)
	switch i {
	case 0:
		return (bits(pte, 53, 10) << 12) | offset, ExcpNone()
	case 1:
		return (ppn[2] << 30) | (ppn[1] << 21) | (vpn[0] << 12) | offset, ExcpNone()
	case 2:
		return (ppn[2] << 30) | (vpn[1] << 21) | (vpn[0] << 12) | offset, ExcpNone()
	default:
		return 0, fault()
	}
}
