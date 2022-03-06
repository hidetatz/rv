package main

const (
	// The number of CSR registers.
	CSRRegsCount = 4096

	// well-known CSRs for Machine-level.
	// Note that MXXX starts from 0x300.

	// Read-only CSRs.
	CsrMVENDORID = 0xf11
	CsrMARCHID   = 0xf12
	CsrMIMPID    = 0xf13
	CsrMHARTID   = 0xf14

	// Normal machine-level CSRs.
	CsrMSTATUS = 0x300
	CsrMIE     = 0x304
	CsrMTVEC   = 0x305
	CsrMEPC    = 0x341
	CsrMCAUSE  = 0x342
	CsrMTVAL   = 0x343
	CsrMIP     = 0x344

	// well-known CSRs for Supervisor-level.
	CsrSSTATUS = 0x100
	CsrSIE     = 0x104
	CsrSIP     = 0x144

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
	// CsrStatusMask is the location which SSTATUS can access (= the access level is under the supervisor).
	CsrSstatusMask = CsrSstatusUie | CsrSstatusSie | CsrSstatusUpie | CsrSstatusSpie | CsrSstatusSpp | CsrSstatusFs | CsrSstatusXs | CsrSstatusSum | CsrSstatusMxr | CsrSstatusUxl | CsrSstatusSd

	// Named SIP fields.
	CarSipUSIP = 0x0000_0000_0000_0001 // sstatus[0]
	CarSipSSIP = 0x0000_0000_0000_0002 // sstatus[1]
	CarSipUTIP = 0x0000_0000_0000_0010 // sstatus[4]
	CarSipSTIP = 0x0000_0000_0000_0020 // sstatus[5]
	CarSipUEIP = 0x0000_0000_0000_0100 // sstatus[8]
	CarSipSEIP = 0x0000_0000_0000_0200 // sstatus[9]
	// CsrSipMask is the location which SIP can access (= the access level is under the supervisor).
	CsrSipMask = CarSipUSIP | CarSipSSIP | CarSipUTIP | CarSipSTIP | CarSipUEIP | CarSipSEIP

	// Named SIE fields.
	CarSieUSIE = 0x0000_0000_0000_0001 // sstatus[0]
	CarSieSSIE = 0x0000_0000_0000_0002 // sstatus[1]
	CarSieUTIE = 0x0000_0000_0000_0010 // sstatus[4]
	CarSieSTIE = 0x0000_0000_0000_0020 // sstatus[5]
	CarSieUEIE = 0x0000_0000_0000_0100 // sstatus[8]
	CarSieSEIE = 0x0000_0000_0000_0200 // sstatus[9]
	// CsrSieMask is the location which SIE can access (= the access level is under the supervisor).
	CsrSieMask = CarSieUSIE | CarSieSSIE | CarSieUTIE | CarSieSTIE | CarSieUEIE | CarSieSEIE
)

// ReadCSR reads CSR by the given address. CSR address is 12-bit.
// This method does not validate the CPU mode. The validation should be the caller's responsibility.
func ReadCSR(csr [CSRRegsCount]uint64, addr uint16) uint64 {
	// assuming addr is small enough, not checking the index

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
