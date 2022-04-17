package main

type CSR [CSRRegsCount]uint64

const (
	// The number of CSR registers.
	CSRRegsCount = 4096

	// Listed CSRs. See privileged architecture spec 2.2.
	// U
	CsrUSTATUS  uint64 = 0x000 // User status register.
	CsrUIE      uint64 = 0x004 // User interrupt-enable register.
	CsrUTVEC    uint64 = 0x005 // User trap handler base address.
	CsrUSCRATCH uint64 = 0x040 // Scratch register for user trap handlers.
	CsrUEPC     uint64 = 0x041 // User exception program counter.
	CsrUCAUSE   uint64 = 0x042 // User trap cause.
	CsrUTVAL    uint64 = 0x043 // User bad address or instruction.
	CsrUIP      uint64 = 0x044 // User interrupt pending.
	CsrFFLAGS   uint64 = 0x001 // Floating-Point Accrued Exceptions.
	CsrFRM      uint64 = 0x002 // Floating-Point Dynamic Rounding Mode.
	CsrFCSR     uint64 = 0x003 // Floating-Point Control and Status Register (frm + fflags).
	CsrCYCLE    uint64 = 0xc00 // Read only. Cycle counter for RDCYCLE instruction.
	CsrTIME     uint64 = 0xc01 // Read only. Timer for RDTIME instruction.
	CsrINSTRET  uint64 = 0xc02 // Read only. Instructions-retired counter for RDINSTRET instruction.
	CsrCYCLEH   uint64 = 0xc80 // Read only. Upper 32 bits of cycle, RV32I only.
	CsrTIMEH    uint64 = 0xc81 // Read only. Upper 32 bits of time, RV32I only.
	CsrINSTRETH uint64 = 0xc82 // Read only. Upper 32 bits of instret, RV32I only.

	// S
	CsrSSTATUS    uint64 = 0x100 // Supervisor status register.
	CsrSEDELEG    uint64 = 0x102 // Supervisor exception delegation register.
	CsrSIDELEG    uint64 = 0x103 // Supervisor interrupt delegation register.
	CsrSIE        uint64 = 0x104 // Supervisor interrupt-enable register.
	CsrSTVEC      uint64 = 0x105 // Supervisor trap handler base address.
	CsrSCOUNTEREN uint64 = 0x106 // Supervisor counter enable.
	CsrSSCRATCH   uint64 = 0x140 // Scratch register for supervisor trap handlers.
	CsrSEPC       uint64 = 0x141 // Supervisor exception program counter.
	CsrSCAUSE     uint64 = 0x142 // Supervisor trap cause.
	CsrSTVAL      uint64 = 0x143 // Supervisor bad address or instruction.
	CsrSIP        uint64 = 0x144 // Supervisor interrupt pending.
	CsrSATP       uint64 = 0x180 // Supervisor address translation and protection.

	// M
	CsrMVENDORID  uint64 = 0xf11 // Read only. Vendor ID.
	CsrMARCHID    uint64 = 0xf12 // Read only. Architecture ID.
	CsrMIMPID     uint64 = 0xf13 // Read only. Implementation ID.
	CsrMHARTID    uint64 = 0xf14 // Read only. Hardware thread ID.
	CsrMSTATUS    uint64 = 0x300 // Machine status register.
	CsrMISA       uint64 = 0x301 // ISA and extensions
	CsrMEDELEG    uint64 = 0x302 // Machine exception delegation register.
	CsrMIDELEG    uint64 = 0x303 // Machine interrupt delegation register.
	CsrMIE        uint64 = 0x304 // Machine interrupt-enable register.
	CsrMTVEC      uint64 = 0x305 // Machine trap-handler base address.
	CsrMCOUNTEREN uint64 = 0x306 // Machine counter enable.
	CsrMSCRATCH   uint64 = 0x340 // Scratch register for machine trap handlers.
	CsrMEPC       uint64 = 0x341 // Machine exception program counter.
	CsrMCAUSE     uint64 = 0x342 // Machine trap cause.
	CsrMTVAL      uint64 = 0x343 // Machine bad address or instruction.
	CsrMIP        uint64 = 0x344 // Machine interrupt pending.
	CsrPMPCFG0    uint64 = 0x3a0 // Physical memory protection configuration.
	CsrPMPCFG1    uint64 = 0x3a1 // Physical memory protection configuration, RV32 only.
	CsrPMPCFG2    uint64 = 0x3a2 // Physical memory protection configuration.
	CsrPMPCFG3    uint64 = 0x3a3 // Physical memory protection configuration, RV32 only.

	// Named SSTATUS fields index.
	// Not all are listed up because they just are not needed.
	CsrStatusUIE   = 0
	CsrStatusSIE   = 1
	CsrStatusMIE   = 3
	CsrStatusUPIE  = 4
	CsrStatusSPIE  = 5
	CsrStatusMPIE  = 7
	CsrStatusSPP   = 8
	CsrStatusMPPLo = 11
	CsrStatusMPPHi = 12
	CsrStatusFSLo  = 13
	CsrStatusFSHi  = 14
	CsrStatusXSLo  = 15
	CsrStatusXSHi  = 16
	CsrStatusMPRV  = 17
	CsrStatusSUM   = 18
	CsrStatusMXR   = 19
	CsrStatusUXLLo = 32
	CsrStatusUXLHi = 33
	CsrStatusSD    = 63
	// CsrStatusMask is the field location which SSTATUS can access (= the access level is under the supervisor).
	CsrSstatusMask = (1 << CsrStatusUIE) |
		(1 << CsrStatusSIE) |
		(1 << CsrStatusUPIE) |
		(1 << CsrStatusSPIE) |
		(1 << CsrStatusSPP) |
		(1 << CsrStatusFSLo) |
		(1 << CsrStatusFSHi) |
		(1 << CsrStatusXSLo) |
		(1 << CsrStatusXSHi) |
		(1 << CsrStatusSUM) |
		(1 << CsrStatusMXR) |
		(1 << CsrStatusUXLHi) |
		(1 << CsrStatusUXLLo) |
		(1 << CsrStatusSD)

	// Named SIP fields.
	CsrSipUSIP uint64 = 0x0000_0000_0000_0001 // sip[0]
	CsrSipSSIP uint64 = 0x0000_0000_0000_0002 // sip[1]
	CsrSipUTIP uint64 = 0x0000_0000_0000_0010 // sip[4]
	CsrSipSTIP uint64 = 0x0000_0000_0000_0020 // sip[5]
	CsrSipUEIP uint64 = 0x0000_0000_0000_0100 // sip[8]
	CsrSipSEIP uint64 = 0x0000_0000_0000_0200 // sip[9]
	// CsrSipMask is the field location which SIP can access (= the access level is under the supervisor).
	CsrSipMask = CsrSipUSIP | CsrSipSSIP | CsrSipUTIP | CsrSipSTIP | CsrSipUEIP | CsrSipSEIP

	// Named SIE fields.
	CsrSieUSIE uint64 = 0x0000_0000_0000_0001 // sie[0]
	CsrSieSSIE uint64 = 0x0000_0000_0000_0002 // sie[1]
	CsrSieUTIE uint64 = 0x0000_0000_0000_0010 // sie[4]
	CsrSieSTIE uint64 = 0x0000_0000_0000_0020 // sie[5]
	CsrSieUEIE uint64 = 0x0000_0000_0000_0100 // sie[8]
	CsrSieSEIE uint64 = 0x0000_0000_0000_0200 // sie[9]
	// CsrSieMask is the field location which SIE can access (= the access level is under the supervisor).
	CsrSieMask = CsrSieUSIE | CsrSieSSIE | CsrSieUTIE | CsrSieSTIE | CsrSieUEIE | CsrSieSEIE
)

func NewCSR() CSR {
	return [CSRRegsCount]uint64{}
}

// Read reads CSR by the given address. CSR address is 12-bit.
// This method does not validate the CPU mode. The validation should be the caller's responsibility.
func (csr CSR) Read(addr uint64) uint64 {
	// assuming addr is small enough, not checking the index

	if addr == CsrFFLAGS {
		// FCSR consists of FRM (3-bit) + FFLAGS (5-bit)
		return csr[CsrFCSR] & 0x1f
	}

	if addr == CsrFRM {
		// FCSR consists of FRM (3-bit) + FFLAGS (5-bit)
		return (csr[CsrFCSR] & 0xe0) >> 5
	}

	// when any of SSTATUS, SIP, SIE is requested, masked MSTATUS, MIP, MIE should be returned because they are subsets.
	// See RISC-V Privileged Architecture Spec 4.1
	if addr == CsrSSTATUS {
		return csr[CsrMSTATUS] & CsrSstatusMask
	}

	if addr == CsrSIP {
		return csr[CsrMIP] & CsrSipMask
	}

	if addr == CsrSIE {
		return csr[CsrMIE] & CsrSieMask
	}

	return csr[addr]
}

// WriteCSR write the given value to the CSR.
// This method does not validate the CPU mode. The validation should be the caller's responsibility.
func (csr CSR) Write(addr uint64, value uint64) {
	if addr == CsrFFLAGS {
		// FCSR consists of FRM (3-bit) + FFLAGS (5-bit)
		csr[CsrFCSR] &= ^uint64(0x1f) // clear fcsr[4:0]
		csr[CsrFCSR] |= value & 0x1f  // write the value[4:0] to the fcsr[4:0]
		return
	}

	if addr == CsrFRM {
		// FCSR consists of FRM (3-bit) + FFLAGS (5-bit)
		csr[CsrFCSR] &= ^uint64(0xe0)       // clear fcsr[7:5]
		csr[CsrFCSR] |= (value << 5) & 0xe0 // write the value[2:0] to the fcsr[7:5]
		return
	}

	if addr == CsrSSTATUS {
		// SSTATUS is a subset of MSTATUS
		csr[CsrMSTATUS] &= ^uint64(CsrSstatusMask) // clear mask
		csr[CsrMSTATUS] |= value & CsrSstatusMask  // write only mask
	}

	if addr == CsrSIE {
		// SIE is a subset of MIE
		csr[CsrMIE] &= ^uint64(CsrSieMask)
		csr[CsrMIE] |= value & CsrSieMask
	}

	if addr == CsrSIP {
		// SIE is a subset of MIE
		csr[CsrMIP] &= ^uint64(CsrSieMask)
		csr[CsrMIP] |= value & CsrSieMask
	}

	csr[addr] = value
}
