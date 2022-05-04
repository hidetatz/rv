package main

// Trap represents a trap including exceptions and interrupts.
type Trap int

const (
	// TrapContained is a trap which is visible to, and handled by, software running inside the execution environment.
	TrapContained Trap = iota + 1

	// TrapRequested is a trap which is a synchronous exception that is an explicit call to the execution
	// environment requesting an action on behalf of software inside the execution environment.
	// An example is a system call.
	TrapRequested

	// TrapInvisible is a trap which is handled transparently by the execution environment and execution
	// resumes normally after the trap is handled. Examples include emulating missing instructions.
	TrapInvisible

	// TrapFatal is a trap which is a fatal failure and causes the execution environment to terminate
	// execution. Examples include failing a virtual-memory page-protection check.
	TrapFatal
)
