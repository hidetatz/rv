package main

// Mode is RISC-V machine status for privilege architecture.
type Mode int

const (
	// User is a mode for application which runs on operating system.
	User Mode = 0
	// Supervisor is a mode for operating system.
	Supervisor Mode = 1
	// Machine is a mode for RISC-V hart internal operation.
	// This sometimes is called kernal-mode or protect-mode in other architecture.
	Machine Mode = 3
)

func (m Mode) Code() int {
	return int(m)
}

func CodeToMode(code int) Mode {
	switch code {
	case 0:
		return User
	case 1:
		return Supervisor
	case 3:
		return Machine
	default:
		panic("unexpected code as mode")
	}
}
