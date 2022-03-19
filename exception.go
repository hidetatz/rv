package main

type Exception uint8

const (
	ExcpNone Exception = iota
	ExcpInstructionAddressMisalighed
	ExcpInstructionAccessFault
	ExcpIllegalInstruction
	ExcpBreakpoint
	ExcpLoadAddressMisaligned
	ExcpLoadAccessFault
	ExcpStoreAMOAddressMisaligned
	ExcpStoreAMOAccessFault
	ExcpEnvironmentCallFromUmode
	ExcpEnvironmentCallFromSmode
	ExcpEnvironmentCallFromMmode
	ExcpInstructionPageFault
	ExcpLoadPageFault
	ExcpStoreAMOPageFault
)
