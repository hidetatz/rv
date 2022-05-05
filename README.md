# rv

rv is a 64-bit RISC-V emulator written in Go.
It only supports running ELF binary.

## Limitation

* 32-bit, 128-bit aren't supported.
* Multi-core emulation isn't supported.
* release/acquire bits are not handled in AMO instructions.
  - rv currently emulates only one hart, so this is not really a problem.
* fence instructions do nothing.
  - rv currently does not apply any optimizations and no out-of-order execution occurs, so it should be fine.

## Installation

```shell
go install github.com/hidetatz/rv@latest
```

## Usage

Pass an ELF program to run to `rv`.

```shell
rv -p ./hello
```

Debug log will be enabled if `-d` option is passed (note that this dumps all the executed instructions and some other information).

## Test

rv uses [riscv-tests](https://github.com/riscv-software-src/riscv-tests) as its E2E test.
You can run that by following command:

```shell
go test -v ./...
```

## Supported instructions

- [ ] RV64G ISA
  - [x] RV64I
  - [x] RV64M
  - [x] RV64A
  - [ ] RV64F
  - [ ] RV64D
  - [x] Zifencei
  - [x] Zicsr
- [x] RV64C ISA
- [x] Privileged ISA

For the full list of the implemented instructions, see [instruction.go](./instruction.go).

## Supported features

- [x] ELF binary load
- [ ] Sv39 (Virtual memory translation)
- [x] CSR
- [ ] Trap
  - [x] Exception
  - [ ] Interrupt
- [ ] Devices
  - [ ] UART
  - [ ] PLIC
  - [ ] CLINT
  - [ ] VirtIO

## LICENSE

MIT
