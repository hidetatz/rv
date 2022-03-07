package main

const (
	// The number of CSR registers.
	CSRRegsCount = 4096

	// Listed CSRs. See privileged architecture spec 2.2.
	// U
	CsrUSTATUS  = 0x000 // User status register.
	CsrUIE      = 0x004 // User interrupt-enable register.
	CsrUTVEC    = 0x005 // User trap handler base address.
	CsrUSCRATCH = 0x040 // Scratch register for user trap handlers.
	CsrUEPC     = 0x041 // User exception program counter.
	CsrUCAUSE   = 0x042 // User trap cause.
	CsrUTVAL    = 0x043 // User bad address or instruction.
	CsrUIP      = 0x044 // User interrupt pending.
	CsrFFLAGS   = 0x001 // Floating-Point Accrued Exceptions.
	CsrFRM      = 0x002 // Floating-Point Dynamic Rounding Mode.
	CsrFCSR     = 0x003 // Floating-Point Control and Status Register (frm + fflags).
	CsrCYCLE    = 0xc00 // Read only. Cycle counter for RDCYCLE instruction.
	CsrTIME     = 0xc01 // Read only. Timer for RDTIME instruction.
	CsrINSTRET  = 0xc02 // Read only. Instructions-retired counter for RDINSTRET instruction.
	CsrCYCLEH   = 0xc80 // Read only. Upper 32 bits of cycle, RV32I only.
	CsrTIMEH    = 0xc81 // Read only. Upper 32 bits of time, RV32I only.
	CsrINSTRETH = 0xc82 // Read only. Upper 32 bits of instret, RV32I only.

	// S
	CsrSSTATUS    = 0x100 // Supervisor status register.
	CsrSEDELEG    = 0x102 // Supervisor exception delegation register.
	CsrSIDELEG    = 0x103 // Supervisor interrupt delegation register.
	CsrSIE        = 0x104 // Supervisor interrupt-enable register.
	CsrSTVEC      = 0x105 // Supervisor trap handler base address.
	CsrSCOUNTEREN = 0x106 // Supervisor counter enable.
	CsrSSCRATCH   = 0x140 // Scratch register for supervisor trap handlers.
	CsrSEPC       = 0x141 // Supervisor exception program counter.
	CsrSCAUSE     = 0x142 // Supervisor trap cause.
	CsrSTVAL      = 0x143 // Supervisor bad address or instruction.
	CsrSIP        = 0x144 // Supervisor interrupt pending.
	CsrSATP       = 0x180 // Supervisor address translation and protection.

	// M
	CsrMVENDORID  = 0xf11 // Read only. Vendor ID.
	CsrMARCHID    = 0xf12 // Read only. Architecture ID.
	CsrMIMPID     = 0xf13 // Read only. Implementation ID.
	CsrMHARTID    = 0xf14 // Read only. Hardware thread ID.
	CsrMSTATUS    = 0x300 // Machine status register.
	CsrMISA       = 0x301 // ISA and extensions
	CsrMEDELEG    = 0x302 // Machine exception delegation register.
	CsrMIDELEG    = 0x303 // Machine interrupt delegation register.
	CsrMIE        = 0x304 // Machine interrupt-enable register.
	CsrMTVEC      = 0x305 // Machine trap-handler base address.
	CsrMCOUNTEREN = 0x306 // Machine counter enable.
	CsrMSCRATCH   = 0x340 // Scratch register for machine trap handlers.
	CsrMEPC       = 0x341 // Machine exception program counter.
	CsrMCAUSE     = 0x342 // Machine trap cause.
	CsrMTVAL      = 0x343 // Machine bad address or instruction.
	CsrMIP        = 0x344 // Machine interrupt pending.
	CsrPMPCFG0    = 0x3a0 // Physical memory protection configuration.
	CsrPMPCFG1    = 0x3a1 // Physical memory protection configuration, RV32 only.
	CsrPMPCFG2    = 0x3a2 // Physical memory protection configuration.
	CsrPMPCFG3    = 0x3a3 // Physical memory protection configuration, RV32 only.

	// Named SSTATUS fields.
	CsrSstatusUie  = 0x0000_0000_0000_0001 // sstatus[0]
	CsrSstatusSie  = 0x0000_0000_0000_0002 // sstatus[1]
	CsrSstatusUpie = 0x0000_0000_0000_0010 // sstatus[4]
	CsrSstatusSpie = 0x0000_0000_0000_0020 // sstatus[5]
	CsrSstatusSpp  = 0x0000_0000_0000_0100 // sstatus[8]
	CsrSstatusFs   = 0x0000_0000_0000_6000 // sstatus[14:13]
	CsrSstatusXs   = 0x0000_0000_0001_8000 // sstatus[16:15]
	CsrSstatusSum  = 0x0000_0000_0004_0000 // sstatus[18]
	CsrSstatusMxr  = 0x0000_0000_0008_0000 // sstatus[19]
	CsrSstatusUxl  = 0x0000_0003_0000_0000 // sstatus[33:32]
	CsrSstatusSd   = 0x8000_0000_0000_0000 // sstatus[63]
	// CsrStatusMask is the field location which SSTATUS can access (= the access level is under the supervisor).
	CsrSstatusMask = CsrSstatusUie | CsrSstatusSie | CsrSstatusUpie | CsrSstatusSpie | CsrSstatusSpp | CsrSstatusFs | CsrSstatusXs | CsrSstatusSum | CsrSstatusMxr | CsrSstatusUxl | CsrSstatusSd

	// Named SIP fields.
	CsrSipUSIP = 0x0000_0000_0000_0001 // sip[0]
	CsrSipSSIP = 0x0000_0000_0000_0002 // sip[1]
	CsrSipUTIP = 0x0000_0000_0000_0010 // sip[4]
	CsrSipSTIP = 0x0000_0000_0000_0020 // sip[5]
	CsrSipUEIP = 0x0000_0000_0000_0100 // sip[8]
	CsrSipSEIP = 0x0000_0000_0000_0200 // sip[9]
	// CsrSipMask is the field location which SIP can access (= the access level is under the supervisor).
	CsrSipMask = CsrSipUSIP | CsrSipSSIP | CsrSipUTIP | CsrSipSTIP | CsrSipUEIP | CsrSipSEIP

	// Named SIE fields.
	CsrSieUSIE = 0x0000_0000_0000_0001 // sie[0]
	CsrSieSSIE = 0x0000_0000_0000_0002 // sie[1]
	CsrSieUTIE = 0x0000_0000_0000_0010 // sie[4]
	CsrSieSTIE = 0x0000_0000_0000_0020 // sie[5]
	CsrSieUEIE = 0x0000_0000_0000_0100 // sie[8]
	CsrSieSEIE = 0x0000_0000_0000_0200 // sie[9]
	// CsrSieMask is the field location which SIE can access (= the access level is under the supervisor).
	CsrSieMask = CsrSieUSIE | CsrSieSSIE | CsrSieUTIE | CsrSieSTIE | CsrSieUEIE | CsrSieSEIE
)

// ReadCSR reads CSR by the given address. CSR address is 12-bit.
// This method does not validate the CPU mode. The validation should be the caller's responsibility.
func ReadCSR(csr [CSRRegsCount]uint64, addr uint16) uint64 {
	// assuming addr is small enough, not checking the index

	if addr == CsrFFLAGS {
		// FCSR consists of FRM + FFLAGS
		return csr[CsrFCSR] & 0b1_1111
	}

	if addr == CsrFRM {
		// FCSR consists of FRM + FFLAGS
		return (csr[CsrFCSR] & 0b1110_0000) >> 5
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
func WriteCSR(csr [CSRRegsCount]uint64, addr uint16, value uint64) {
}
