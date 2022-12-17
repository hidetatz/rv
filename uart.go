package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"sync"
)

const (
	ierRxintBit   = 0x1
	ierThreintBit = 0x2

	iirThrEmpty    = 0x2
	iirRdAvailable = 0x4
	iirNoInterrupt = 0x7

	lsrDataAvailable = 0x1
	lsrThrEmpty      = 0x20
)

type Uart struct {
	clock        uint64
	rbr          uint8 // receiver buffer register
	thr          uint8 // transmitter holding register
	ier          uint8 // interrupt enable register
	iir          uint8 // interrupt identification register
	lcr          uint8 // line control register
	mcr          uint8 // modem control register
	lsr          uint8 // line status register
	scr          uint8 // scratch,
	threip       bool
	interrupting bool

	sync.Mutex
	buffer []byte
}

func NewUart() *Uart {
	u := &Uart{
		clock:        0,
		rbr:          0,
		thr:          0,
		ier:          0,
		iir:          0,
		lcr:          0,
		mcr:          0,
		lsr:          lsrThrEmpty,
		scr:          0,
		threip:       false,
		interrupting: false,

		buffer: []byte{}, // stdin buffer
	}
	// read input
	go func() {
		r := bufio.NewReader(os.Stdin)
		for {
			b, err := r.ReadByte()
			if err != nil {
				if err == io.EOF {
					continue
				}
				fmt.Fprintf(os.Stderr, "read stdin: %s", err)
				continue
			}

			u.Lock()
			u.buffer = append(u.buffer, b)
			u.Unlock()
		}
	}()

	return u
}

func (u *Uart) Tick() {
	u.clock++
	rxip := false

	// read input and store into register
	if u.clock%0x38400 == 0 && u.rbr == 0 {
		// get single byte
		u.Lock()
		b := u.buffer[0]
		u.buffer = u.buffer[1:]
		u.Unlock()

		if b != 0 {
			u.rbr = b
			u.lsr |= lsrDataAvailable
			u.updateIir()
			if u.ier&ierRxintBit != 0 {
				rxip = true
			}
		}
	}

	// write reg value to stdout
	if u.clock%0x10 == 0 && u.thr != 0 {
		fmt.Fprint(os.Stdout, string(u.thr))
		u.thr = 0
		u.lsr |= lsrThrEmpty
		u.updateIir()
		if u.ier&ierThreintBit != 0 {
			u.threip = true
		}
	}

	if u.threip || rxip {
		u.interrupting = true
		u.threip = false
	} else {
		u.interrupting = false
	}
}

func (u *Uart) updateIir() {
	rxip := u.ier&ierRxintBit != 0 && u.rbr != 0
	threip := u.ier&ierThreintBit != 0 && u.thr == 0

	if rxip {
		u.iir = iirRdAvailable
	} else if threip {
		u.iir = iirThrEmpty
	} else {
		u.iir = iirNoInterrupt
	}
}

func (u *Uart) read(address uint64) uint8 {
	switch address {
	case 0x10000000:
		if (u.lcr >> 7) == 0 {
			rbr := u.rbr
			u.rbr = 0
			u.lsr &= ^uint8(lsrDataAvailable)
			u.updateIir()
			return rbr
		}
	case 0x10000001:
		if (u.lcr >> 7) == 0 {
			return u.ier
		}
	case 0x10000002:
		return u.iir
	case 0x10000003:
		return u.lcr
	case 0x10000004:
		return u.mcr
	case 0x10000005:
		return u.lsr
	case 0x10000007:
		return u.scr
	}
	return 0
}

func (u *Uart) write(address uint64, value uint8) {
	switch address {
	case 0x10000000:
		if (u.lcr >> 7) == 0 {
			u.thr = value
			u.lsr &= ^uint8(lsrThrEmpty)
			u.updateIir()
		}
	case 0x10000001:
		if u.ier&ierThreintBit == 0 && value&ierThreintBit != 0 && u.thr == 0 {
			u.threip = true
		}

		u.ier = value
		u.updateIir()
	case 0x10000003:
		u.lcr = value
	case 0x10000004:
		u.mcr = value
	case 0x10000007:
		u.scr = value
	}
}
