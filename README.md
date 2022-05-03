# rv

rv is a 64-bit RISC-V emulator written in Go.
It only supports running ELF binary.

## Installation

```
go install github.com/hidetatz/rv@latest
```

## Usage

Pass an ELF program to run to `rv`.

```
rv -p ./hello
```

Debug log will be enabled if `-d` option is passed (note that this dumps all the executed instructions and some other information).

## Supported instructions

- [ ] RV64G ISA
  - [x] RV64I
  - [ ] RV64M
  - [ ] RV64A
  - [ ] RV64F
  - [ ] RV64D
  - [x] Zifencei
  - [x] Zicsr
- [x] RV64C ISA
- [x] Privileged ISA

For the full list of the implemented instructions, see [instruction.go](./instruction.go).

## Supported features

- [x] ELF binary load
- [ ] Sv39 (Virtual memory)
- [x] CSR
- [ ] Devices

## LICENSE

MIT
