package main

const (
	// U
	CsrUSTATUS uint64 = 0x000 // User status register.
	CsrUTVEC   uint64 = 0x005 // User trap handler base address.
	CsrUEPC    uint64 = 0x041 // User exception program counter.
	CsrUCAUSE  uint64 = 0x042 // User trap cause.
	CsrUTVAL   uint64 = 0x043 // User bad address or instruction.
	CsrFFLAGS  uint64 = 0x001 // Floating-Point Accrued Exceptions.
	CsrFRM     uint64 = 0x002 // Floating-Point Dynamic Rounding Mode.
	CsrFCSR    uint64 = 0x003 // Floating-Point Control and Status Register (frm + fflags).

	// S
	CsrSSTATUS uint64 = 0x100 // Supervisor status register.
	CsrSEDELEG uint64 = 0x102 // Supervisor exception delegation register.
	CsrSIE     uint64 = 0x104 // Supervisor interrupt-enable register.
	CsrSTVEC   uint64 = 0x105 // Supervisor trap handler base address.
	CsrSEPC    uint64 = 0x141 // Supervisor exception program counter.
	CsrSCAUSE  uint64 = 0x142 // Supervisor trap cause.
	CsrSTVAL   uint64 = 0x143 // Supervisor bad address or instruction.
	CsrSIP     uint64 = 0x144 // Supervisor interrupt pending.
	CsrSATP    uint64 = 0x180 // Supervisor address translation and protection.

	//// M
	CsrMSTATUS uint64 = 0x300 // Machine status register.
	CsrMEDELEG uint64 = 0x302 // Machine exception delegation register.
	CsrMIE     uint64 = 0x304 // Machine interrupt-enable register.
	CsrMTVEC   uint64 = 0x305 // Machine trap-handler base address.
	CsrMEPC    uint64 = 0x341 // Machine exception program counter.
	CsrMCAUSE  uint64 = 0x342 // Machine trap cause.
	CsrMTVAL   uint64 = 0x343 // Machine bad address or instruction.
	CsrMIP     uint64 = 0x344 // Machine interrupt pending.

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
