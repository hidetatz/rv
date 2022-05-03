package main

type ExceptionCode int

const (
	ExcpCodeNone                         ExceptionCode = -1
	ExcpCodeInstructionAddressMisalighed ExceptionCode = 0
	ExcpCodeInstructionAccessFault       ExceptionCode = 1
	ExcpCodeIllegalInstruction           ExceptionCode = 2
	ExcpCodeBreakpoint                   ExceptionCode = 3
	ExcpCodeLoadAddressMisaligned        ExceptionCode = 4
	ExcpCodeLoadAccessFault              ExceptionCode = 5
	ExcpCodeStoreAMOAddressMisaligned    ExceptionCode = 6
	ExcpCodeStoreAMOAccessFault          ExceptionCode = 7
	ExcpCodeEnvironmentCallFromUmode     ExceptionCode = 8
	ExcpCodeEnvironmentCallFromSmode     ExceptionCode = 9
	ExcpCodeEnvironmentCallFromMmode     ExceptionCode = 11
	ExcpCodeInstructionPageFault         ExceptionCode = 12
	ExcpCodeLoadPageFault                ExceptionCode = 13
	ExcpCodeStoreAMOPageFault            ExceptionCode = 15
)

type Exception struct {
	Code      ExceptionCode
	TrapValue uint64
}

func ExcpNone() *Exception {
	return &Exception{Code: ExcpCodeNone, TrapValue: 0}
}

func ExcpInstructionAddressMisalighed(pc uint64) *Exception {
	return &Exception{Code: ExcpCodeInstructionAddressMisalighed, TrapValue: pc}
}

func ExcpInstructionAccessFault(pc uint64) *Exception {
	return &Exception{Code: ExcpCodeInstructionAccessFault, TrapValue: pc}
}

func ExcpIllegalInstruction(v uint64) *Exception {
	return &Exception{Code: ExcpCodeIllegalInstruction, TrapValue: v}
}

func ExcpBreakpoint(pc uint64) *Exception {
	return &Exception{Code: ExcpCodeBreakpoint, TrapValue: pc}
}

func ExcpLoadAddressMisaligned(pc uint64) *Exception {
	return &Exception{Code: ExcpCodeLoadAddressMisaligned, TrapValue: pc}
}

func ExcpLoadAccessFault(pc uint64) *Exception {
	return &Exception{Code: ExcpCodeLoadAccessFault, TrapValue: pc}
}

func ExcpStoreAMOAddressMisaligned(pc uint64) *Exception {
	return &Exception{Code: ExcpCodeStoreAMOAddressMisaligned, TrapValue: pc}
}

func ExcpStoreAMOAccessFault(pc uint64) *Exception {
	return &Exception{Code: ExcpCodeStoreAMOAccessFault, TrapValue: pc}
}

func ExcpEnvironmentCallFromUmode() *Exception {
	return &Exception{Code: ExcpCodeEnvironmentCallFromUmode, TrapValue: 0}
}

func ExcpEnvironmentCallFromSmode() *Exception {
	return &Exception{Code: ExcpCodeEnvironmentCallFromSmode, TrapValue: 0}
}

func ExcpEnvironmentCallFromMmode() *Exception {
	return &Exception{Code: ExcpCodeEnvironmentCallFromMmode, TrapValue: 0}
}

func ExcpInstructionPageFault(v uint64) *Exception {
	return &Exception{Code: ExcpCodeInstructionPageFault, TrapValue: v}
}

func ExcpLoadPageFault(v uint64) *Exception {
	return &Exception{Code: ExcpCodeLoadPageFault, TrapValue: v}
}

func ExcpStoreAMOPageFault(v uint64) *Exception {
	return &Exception{Code: ExcpCodeStoreAMOPageFault, TrapValue: v}
}
