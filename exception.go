package main

type Exception struct {
	Code      int
	TrapValue uint64
}

func ExcpNone() *Exception {
	return &Exception{Code: -1, TrapValue: 0}
}

func ExcpInstructionAddressMisalighed(pc uint64) *Exception {
	return &Exception{Code: 0, TrapValue: pc}
}

func ExcpInstructionAccessFault(pc uint64) *Exception {
	return &Exception{Code: 1, TrapValue: pc}
}

func ExcpIllegalInstruction(v uint64) *Exception {
	return &Exception{Code: 2, TrapValue: v}
}

func ExcpBreakpoint(pc uint64) *Exception {
	return &Exception{Code: 3, TrapValue: pc}
}

func ExcpLoadAddressMisaligned(pc uint64) *Exception {
	return &Exception{Code: 4, TrapValue: pc}
}

func ExcpLoadAccessFault(pc uint64) *Exception {
	return &Exception{Code: 5, TrapValue: pc}
}

func ExcpStoreAMOAddressMisaligned(pc uint64) *Exception {
	return &Exception{Code: 6, TrapValue: pc}
}

func ExcpStoreAMOAccessFault(pc uint64) *Exception {
	return &Exception{Code: 7, TrapValue: pc}
}

func ExcpEnvironmentCallFromUmode() *Exception {
	return &Exception{Code: 8, TrapValue: 0}
}

func ExcpEnvironmentCallFromSmode() *Exception {
	return &Exception{Code: 9, TrapValue: 0}
}

func ExcpEnvironmentCallFromMmode() *Exception {
	return &Exception{Code: 11, TrapValue: 0}
}

func ExcpInstructionPageFault(v uint64) *Exception {
	return &Exception{Code: 12, TrapValue: v}
}

func ExcpLoadPageFault(v uint64) *Exception {
	return &Exception{Code: 13, TrapValue: v}
}

func ExcpStoreAMOPageFault(v uint64) *Exception {
	return &Exception{Code: 15, TrapValue: v}
}
