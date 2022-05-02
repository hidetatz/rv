package main

type Exception int

const (
	// not using iota because code is not just a sequence number
	ExcpNone                         Exception = -1
	ExcpInstructionAddressMisalighed Exception = 0
	ExcpInstructionAccessFault       Exception = 1
	ExcpIllegalInstruction           Exception = 2
	ExcpBreakpoint                   Exception = 3
	ExcpLoadAddressMisaligned        Exception = 4
	ExcpLoadAccessFault              Exception = 5
	ExcpStoreAMOAddressMisaligned    Exception = 6
	ExcpStoreAMOAccessFault          Exception = 7
	ExcpEnvironmentCallFromUmode     Exception = 8
	ExcpEnvironmentCallFromSmode     Exception = 9
	ExcpEnvironmentCallFromMmode     Exception = 11
	ExcpInstructionPageFault         Exception = 12
	ExcpLoadPageFault                Exception = 13
	ExcpStoreAMOPageFault            Exception = 15
)

func (e Exception) Code() int {
	return int(e)
}
