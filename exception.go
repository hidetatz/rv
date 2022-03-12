package main

type Exception uint8

const (
	ExcpInstructionAddressMisalighed Exception = iota + 1
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
