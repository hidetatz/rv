package main

type CSR struct {
	// There are 4096 (2 ^ 12) control status registers and
	// each of them is 32-bit.
	Regs [4096]uint32
}

const (
	// well-known CSRs
	// Mxxx represents Machine status register.

	// CsrMSTATUS represents the current processor state.
	CsrMSTATUS = 0x300

	// CsrMIE tells if interrupt is enabled.
	CsrMIE = 0x304

	// CsrMTVEC stores the pc to jump on an interrupt/exception.
	CsrMTVEC = 0x305

	// CsrMEPC stores the location which an exception occurs.
	// mret uses this to get back to the original place.
	CsrMEPC = 0x341

	// CsrMCAUSE represents the reason why an interrupt/exception occured.
	// If it is interrupt, the top bit will be 1.
	CsrMCAUSE = 0x342

	// CsrMTVAL stores a value, what is stored is determined by the type of the exception.
	CsrMTVAL = 0x343

	// CsrMIP represents the status of interrupt wait queue.
	CsrMIP = 0x344

	// Sxxx represents supervisor status register.

	CsrSSTATUS = 0x100

	// Other stuffs

	// Named SSTATUS fields.
	CsrSstatusSie  = 0x2                       // sstatus[1]
	CsrSstatusSpie = 0x20                      // sstatus[5]
	CsrSstatusUbe  = 0x40                      // sstatus[6]
	CsrSstatusSpp  = 0x10_0                    // sstatus[8]
	CsrSstatusFs   = 0x60_00                   // sstatus[14:13]
	CsrSstatusXs   = 0x18_00_0                 // sstatus[16:15]
	CsrSstatusSum  = 0x40_00_0                 // sstatus[18]
	CsrSstatusMxr  = 0x80_00_0                 // sstatus[19]
	CsrSstatusUxl  = 0x3_00_00_00_00           // sstatus[33:32]
	CsrSstatusSd   = 0x80_00_00_00_00_00_00_00 // sstatus[63]
	CsrSstatusMask = SSTATUS_SIE | SSTATUS_SPIE | SSTATUS_UBE | SSTATUS_SPP | SSTATUS_FS |
		SSTATUS_XS | SSTATUS_SUM | SSTATUS_MXR | SSTATUS_UXL | SSTATUS_SD
)

func NewCSR() *CSR {
	return &CSR{Regs: [4096]uint32{}}
}

// Read reads CSR by the given address.
// This method does not validate the CPU mode.
// CSR address is 12-bit.
func (csr *CSR) Read(addr uint16) uint32 {
	// assuming addr is small enough, not checking the index

	// when SSTATUS is requested, masked MSTATUS should be returned because
	// SSTATUS is a subset of MSTATUS. See RISC-V Privileged Architecture Spec 4.1
	if addr == CsrSSTATUS {
		return csr.Read(CsrMSTATUS) & CsrSSTATUSMask
	}
}
