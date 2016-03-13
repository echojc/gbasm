package main

import (
	"errors"
	"fmt"
	"reflect"
	"strconv"
)

func Assemble(insns []Insn) ([]uint8, error) {
	out := make([]uint8, 0, len(insns))

	for _, insn := range insns {
		asm, err := assembleInsn(&insn)
		if err != nil {
			return nil, err
		}
		out = append(out, asm...)
	}

	return out, nil
}

func assembleInsn(insn *Insn) ([]uint8, error) {
	switch insn.Name {
	case "ld":
		if len(insn.Args) == 2 {
			addrHi, errAddrHi := asmAddr16(insn.Args[0])
			addrLo, errAddrLo := asmAddr16(insn.Args[1])
			reg16Hi, errReg16Hi := asmReg16(insn.Args[0])
			num16Lo, errNum16Lo := asmUint16(insn.Args[1])
			addrRegHi, errAddrRegHi := asmAddrReg(insn.Args[0])
			addrRegLo, errAddrRegLo := asmAddrReg(insn.Args[1])
			reg8Hi, errReg8Hi := asmReg8Hi(insn.Args[0])
			reg8Lo, errReg8Lo := asmReg8Lo(insn.Args[1])
			num8Lo, errNum8Lo := asmUint8(insn.Args[1])

			switch {
			case errReg8Hi == nil && errReg8Lo == nil:
				return []uint8{0x40 | reg8Hi | reg8Lo}, nil
			case errReg8Hi == nil && errNum8Lo == nil:
				return []uint8{0x06 | reg8Hi, num8Lo}, nil
			case errReg16Hi == nil && errNum16Lo == nil:
				return []uint8{0x01 | reg16Hi, uint8(num16Lo & 0xff), uint8(num16Lo >> 8)}, nil
			case errAddrRegHi == nil && insn.Args[1] == "a":
				return []uint8{0x02 | addrRegHi}, nil
			case insn.Args[0] == "a" && errAddrRegLo == nil:
				return []uint8{0x0a | addrRegLo}, nil
			case insn.Args[0] == "a" && errAddrLo == nil:
				return []uint8{0xfa, uint8(addrLo & 0xff), uint8(addrLo >> 8)}, nil
			case errAddrHi == nil && insn.Args[1] == "a":
				return []uint8{0xea, uint8(addrHi & 0xff), uint8(addrHi >> 8)}, nil
			case errAddrHi == nil && insn.Args[1] == "sp":
				return []uint8{0x08, uint8(addrHi & 0xff), uint8(addrHi >> 8)}, nil
			default:
				insn.Err = errors.New(fmt.Sprintf("ld has invalid args"))
				return nil, insn
			}
		} else {
			return nil, insn.expectedNumberArgs(2)
		}
	case "ldi":
		if reflect.DeepEqual(insn.Args, []string{"(hl)", "a"}) {
			return []uint8{0x22}, nil
		} else if reflect.DeepEqual(insn.Args, []string{"a", "(hl)"}) {
			return []uint8{0x2a}, nil
		} else {
			validArgs := []string{"(hl), a", "a, (hl)"}
			insn.Err = errors.New(fmt.Sprintf("ldi expects %s only", validArgs))
			return nil, insn
		}
	case "ldd":
		if reflect.DeepEqual(insn.Args, []string{"(hl)", "a"}) {
			return []uint8{0x32}, nil
		} else if reflect.DeepEqual(insn.Args, []string{"a", "(hl)"}) {
			return []uint8{0x3a}, nil
		} else {
			validArgs := []string{"(hl), a", "a, (hl)"}
			insn.Err = errors.New(fmt.Sprintf("ldd expects %s only", validArgs))
			return nil, insn
		}
	case "ldh":
		if reflect.DeepEqual(insn.Args, []string{"a", "(c)"}) {
			return []uint8{0xf2}, nil
		} else if reflect.DeepEqual(insn.Args, []string{"(c)", "a"}) {
			return []uint8{0xe2}, nil
		} else if len(insn.Args) == 2 && insn.Args[0] == "a" {
			addr, err := asmAddr8(insn.Args[1])
			if err != nil {
				insn.Err = err
				return nil, insn
			}
			return []uint8{0xf0, addr}, nil
		} else if len(insn.Args) == 2 && insn.Args[1] == "a" {
			addr, err := asmAddr8(insn.Args[0])
			if err != nil {
				insn.Err = err
				return nil, insn
			}
			return []uint8{0xe0, addr}, nil
		} else {
			validArgs := []string{"(c), a", "a, (c)", "(n), a", "a, (n)"}
			insn.Err = errors.New(fmt.Sprintf("ldd expects %s only", validArgs))
			return nil, insn
		}
	case "ldhl":
		if len(insn.Args) == 2 && insn.Args[0] == "sp" {
			num, err := asmInt8(insn.Args[1])
			if err != nil {
				insn.Err = err
				return nil, insn
			}
			return []uint8{0xf8, uint8(num)}, nil
		} else {
			validArgs := []string{"sp, n"}
			insn.Err = errors.New(fmt.Sprintf("ldhl expects %s only", validArgs))
			return nil, insn
		}
	case "inc":
		if len(insn.Args) == 1 {
			reg, err := asmReg8Hi(insn.Args[0])
			if err != nil {
				insn.Err = err
				return nil, insn
			}
			return []uint8{0x04 | reg}, nil
		} else {
			return nil, insn.expectedNumberArgs(1)
		}
	case "dec":
		if len(insn.Args) == 1 {
			reg, err := asmReg8Hi(insn.Args[0])
			if err != nil {
				insn.Err = err
				return nil, insn
			}
			return []uint8{0x05 | reg}, nil
		} else {
			return nil, insn.expectedNumberArgs(1)
		}
	case "add":
		if len(insn.Args) == 2 {
			switch insn.Args[0] {
			case "hl":
				reg, err := asmReg16(insn.Args[1])
				if err != nil {
					insn.Err = err
					return nil, insn
				}
				return []uint8{0x09 | reg}, nil
			case "sp":
				num, err := asmInt8(insn.Args[1])
				if err != nil {
					insn.Err = err
					return nil, insn
				}
				return []uint8{0xe8, uint8(num)}, nil
			case "a":
				reg, err1 := asmReg8Lo(insn.Args[1])
				if err1 != nil {
					num, err2 := asmUint8(insn.Args[1])
					if err2 != nil {
						insn.Err = errors.New(fmt.Sprintf("%s or %s", err1.Error(), err2.Error()))
						return nil, insn
					}
					return []uint8{0xc6, num}, nil
				} else {
					return []uint8{0x80 | reg}, nil
				}
			default:
				validArgs := []string{"hl", "sp", "a"}
				insn.Err = errors.New(fmt.Sprintf("add expects %d as first arg", validArgs))
			}
		} else {
			return nil, insn.expectedNumberArgs(2)
		}
	case "adc":
		if len(insn.Args) == 2 {
			if insn.Args[0] == "a" {
				reg, err1 := asmReg8Lo(insn.Args[1])
				if err1 != nil {
					num, err2 := asmUint8(insn.Args[1])
					if err2 != nil {
						insn.Err = errors.New(fmt.Sprintf("%s or %s", err1.Error(), err2.Error()))
						return nil, insn
					}
					return []uint8{0xce, num}, nil
				} else {
					return []uint8{0x88 | reg}, nil
				}
			} else {
				validArgs := []string{"a"}
				insn.Err = errors.New(fmt.Sprintf("adc expects %d as first arg", validArgs))
			}
		} else {
			return nil, insn.expectedNumberArgs(2)
		}
	case "sub":
		if len(insn.Args) == 2 {
			if insn.Args[0] == "a" {
				reg, err1 := asmReg8Lo(insn.Args[1])
				if err1 != nil {
					num, err2 := asmUint8(insn.Args[1])
					if err2 != nil {
						insn.Err = errors.New(fmt.Sprintf("%s or %s", err1.Error(), err2.Error()))
						return nil, insn
					}
					return []uint8{0xd6, num}, nil
				} else {
					return []uint8{0x90 | reg}, nil
				}
			} else {
				validArgs := []string{"a"}
				insn.Err = errors.New(fmt.Sprintf("sub expects %d as first arg", validArgs))
			}
		} else {
			return nil, insn.expectedNumberArgs(2)
		}
	case "sbc":
		if len(insn.Args) == 2 {
			if insn.Args[0] == "a" {
				reg, err1 := asmReg8Lo(insn.Args[1])
				if err1 != nil {
					num, err2 := asmUint8(insn.Args[1])
					if err2 != nil {
						insn.Err = errors.New(fmt.Sprintf("%s or %s", err1.Error(), err2.Error()))
						return nil, insn
					}
					return []uint8{0xde, num}, nil
				} else {
					return []uint8{0x98 | reg}, nil
				}
			} else {
				validArgs := []string{"a"}
				insn.Err = errors.New(fmt.Sprintf("sub expects %d as first arg", validArgs))
			}
		} else {
			return nil, insn.expectedNumberArgs(2)
		}
	case "and":
		if len(insn.Args) == 1 {
			reg, err1 := asmReg8Lo(insn.Args[0])
			if err1 != nil {
				num, err2 := asmUint8(insn.Args[0])
				if err2 != nil {
					insn.Err = errors.New(fmt.Sprintf("%s or %s", err1.Error(), err2.Error()))
					return nil, insn
				}
				return []uint8{0xe6, num}, nil
			} else {
				return []uint8{0xa0 | reg}, nil
			}
		} else {
			return nil, insn.expectedNumberArgs(1)
		}
	case "xor":
		if len(insn.Args) == 1 {
			reg, err1 := asmReg8Lo(insn.Args[0])
			if err1 != nil {
				num, err2 := asmUint8(insn.Args[0])
				if err2 != nil {
					insn.Err = errors.New(fmt.Sprintf("%s or %s", err1.Error(), err2.Error()))
					return nil, insn
				}
				return []uint8{0xee, num}, nil
			} else {
				return []uint8{0xa8 | reg}, nil
			}
		} else {
			return nil, insn.expectedNumberArgs(1)
		}
	case "or":
		if len(insn.Args) == 1 {
			reg, err1 := asmReg8Lo(insn.Args[0])
			if err1 != nil {
				num, err2 := asmUint8(insn.Args[0])
				if err2 != nil {
					insn.Err = errors.New(fmt.Sprintf("%s or %s", err1.Error(), err2.Error()))
					return nil, insn
				}
				return []uint8{0xf6, num}, nil
			} else {
				return []uint8{0xb0 | reg}, nil
			}
		} else {
			return nil, insn.expectedNumberArgs(1)
		}
	case "cp":
		if len(insn.Args) == 1 {
			reg, err1 := asmReg8Lo(insn.Args[0])
			if err1 != nil {
				num, err2 := asmUint8(insn.Args[0])
				if err2 != nil {
					insn.Err = errors.New(fmt.Sprintf("%s or %s", err1.Error(), err2.Error()))
					return nil, insn
				}
				return []uint8{0xfe, num}, nil
			} else {
				return []uint8{0xb8 | reg}, nil
			}
		} else {
			return nil, insn.expectedNumberArgs(1)
		}
	case "rlca":
		return []uint8{0x07}, nil
	case "rla":
		return []uint8{0x17}, nil
	case "rrca":
		return []uint8{0x0f}, nil
	case "rra":
		return []uint8{0x1f}, nil
	case "jr":
		switch len(insn.Args) {
		case 1:
			addr, err := asmInt8(insn.Args[0])
			if err != nil {
				insn.Err = err
				return nil, insn
			}
			return []uint8{0x18, uint8(addr)}, nil
		case 2:
			cond, err := asmCond(insn.Args[0])
			if err != nil {
				insn.Err = err
				return nil, insn
			}
			addr, err := asmInt8(insn.Args[1])
			if err != nil {
				insn.Err = err
				return nil, insn
			}
			return []uint8{0x20 | cond, uint8(addr)}, nil
		default:
			return nil, insn.expectedNumberArgs(1, 2)
		}
	case "jp":
		switch len(insn.Args) {
		case 1:
			if insn.Args[0] == "hl" {
				return []uint8{0xe9}, nil
			} else {
				addr, err := asmUint16(insn.Args[0])
				if err != nil {
					insn.Err = err
					return nil, insn
				}
				return []uint8{0xc3, uint8(addr & 0xff), uint8(addr >> 8)}, nil
			}
		case 2:
			cond, err := asmCond(insn.Args[0])
			if err != nil {
				insn.Err = err
				return nil, insn
			}
			addr, err := asmUint16(insn.Args[1])
			if err != nil {
				insn.Err = err
				return nil, insn
			}
			return []uint8{0xc2 | cond, uint8(addr & 0xff), uint8(addr >> 8)}, nil
		default:
			return nil, insn.expectedNumberArgs(1, 2)
		}
	case "daa":
		return []uint8{0x27}, nil
	case "cpl":
		return []uint8{0x2f}, nil
	case "scf":
		return []uint8{0x37}, nil
	case "ccf":
		return []uint8{0x3f}, nil
	case "push":
		if len(insn.Args) == 1 {
			reg, err := asmReg16PushPop(insn.Args[0])
			if err != nil {
				insn.Err = err
				return nil, insn
			}
			return []uint8{0xc5 | reg}, nil
		} else {
			return nil, insn.expectedNumberArgs(1)
		}
	case "pop":
		if len(insn.Args) == 1 {
			reg, err := asmReg16PushPop(insn.Args[0])
			if err != nil {
				insn.Err = err
				return nil, insn
			}
			return []uint8{0xc1 | reg}, nil
		} else {
			return nil, insn.expectedNumberArgs(1)
		}
	case "call":
		switch len(insn.Args) {
		case 1:
			addr, err := asmUint16(insn.Args[0])
			if err != nil {
				insn.Err = err
				return nil, insn
			}
			return []uint8{0xcd, uint8(addr & 0xff), uint8(addr >> 8)}, nil
		case 2:
			cond, err := asmCond(insn.Args[0])
			if err != nil {
				insn.Err = err
				return nil, insn
			}
			addr, err := asmUint16(insn.Args[1])
			if err != nil {
				insn.Err = err
				return nil, insn
			}
			return []uint8{0xc4 | cond, uint8(addr & 0xff), uint8(addr >> 8)}, nil
		default:
			return nil, insn.expectedNumberArgs(1, 2)
		}
	case "ret":
		switch len(insn.Args) {
		case 0:
			return []uint8{0xc9}, nil
		case 1:
			cond, err := asmCond(insn.Args[0])
			if err != nil {
				insn.Err = err
				return nil, insn
			}
			return []uint8{0xc0 | cond}, nil
		default:
			return nil, insn.expectedNumberArgs(0, 1)
		}
	case "reti":
		return []uint8{0xd9}, nil
	case "di":
		return []uint8{0xf3}, nil
	case "ei":
		return []uint8{0xfb}, nil
	case "rst":
		if len(insn.Args) == 1 {
			addr, err := asmUint16(insn.Args[0])
			if err != nil {
				insn.Err = err
				return nil, insn
			} else if addr > 0x38 || addr%0x08 != 0 {
				validAddrs := []string{"$00", "$08", "$10", "$18", "$20", "$28", "$30", "$38"}
				insn.Err = errors.New(fmt.Sprintf("rst expects %d only", validAddrs))
				return nil, insn
			}
			return []uint8{0xc7 | (uint8(addr) >> 3)}, nil
		} else {
			return nil, insn.expectedNumberArgs(1)
		}
	case "halt":
		return []uint8{0x76, 0x00}, nil
	case "stop":
		return []uint8{0x10, 0x00}, nil
	case "rlc":
		if len(insn.Args) == 1 {
			reg, err := asmReg8Lo(insn.Args[0])
			if err != nil {
				insn.Err = err
				return nil, insn
			}
			return []uint8{0xcb, 0x00 | reg}, nil
		} else {
			return nil, insn.expectedNumberArgs(1)
		}
	case "rl":
		if len(insn.Args) == 1 {
			reg, err := asmReg8Lo(insn.Args[0])
			if err != nil {
				insn.Err = err
				return nil, insn
			}
			return []uint8{0xcb, 0x10 | reg}, nil
		} else {
			return nil, insn.expectedNumberArgs(1)
		}
	case "rrc":
		if len(insn.Args) == 1 {
			reg, err := asmReg8Lo(insn.Args[0])
			if err != nil {
				insn.Err = err
				return nil, insn
			}
			return []uint8{0xcb, 0x08 | reg}, nil
		} else {
			return nil, insn.expectedNumberArgs(1)
		}
	case "rr":
		if len(insn.Args) == 1 {
			reg, err := asmReg8Lo(insn.Args[0])
			if err != nil {
				insn.Err = err
				return nil, insn
			}
			return []uint8{0xcb, 0x18 | reg}, nil
		} else {
			return nil, insn.expectedNumberArgs(1)
		}
	case "sla":
		if len(insn.Args) == 1 {
			reg, err := asmReg8Lo(insn.Args[0])
			if err != nil {
				insn.Err = err
				return nil, insn
			}
			return []uint8{0xcb, 0x20 | reg}, nil
		} else {
			return nil, insn.expectedNumberArgs(1)
		}
	case "sra":
		if len(insn.Args) == 1 {
			reg, err := asmReg8Lo(insn.Args[0])
			if err != nil {
				insn.Err = err
				return nil, insn
			}
			return []uint8{0xcb, 0x28 | reg}, nil
		} else {
			return nil, insn.expectedNumberArgs(1)
		}
	case "swap":
		if len(insn.Args) == 1 {
			reg, err := asmReg8Lo(insn.Args[0])
			if err != nil {
				insn.Err = err
				return nil, insn
			}
			return []uint8{0xcb, 0x30 | reg}, nil
		} else {
			return nil, insn.expectedNumberArgs(1)
		}
	case "srl":
		if len(insn.Args) == 1 {
			reg, err := asmReg8Lo(insn.Args[0])
			if err != nil {
				insn.Err = err
				return nil, insn
			}
			return []uint8{0xcb, 0x38 | reg}, nil
		} else {
			return nil, insn.expectedNumberArgs(1)
		}
	case "bit":
		if len(insn.Args) == 2 {
			bit, err := asmBit(insn.Args[0])
			if err != nil {
				insn.Err = err
				return nil, insn
			}
			reg, err := asmReg8Lo(insn.Args[1])
			if err != nil {
				insn.Err = err
				return nil, insn
			}
			return []uint8{0xcb, 0x40 | bit | reg}, nil
		} else {
			return nil, insn.expectedNumberArgs(2)
		}
	case "res":
		if len(insn.Args) == 2 {
			bit, err := asmBit(insn.Args[0])
			if err != nil {
				insn.Err = err
				return nil, insn
			}
			reg, err := asmReg8Lo(insn.Args[1])
			if err != nil {
				insn.Err = err
				return nil, insn
			}
			return []uint8{0xcb, 0x80 | bit | reg}, nil
		} else {
			return nil, insn.expectedNumberArgs(2)
		}
	case "set":
		if len(insn.Args) == 2 {
			bit, err := asmBit(insn.Args[0])
			if err != nil {
				insn.Err = err
				return nil, insn
			}
			reg, err := asmReg8Lo(insn.Args[1])
			if err != nil {
				insn.Err = err
				return nil, insn
			}
			return []uint8{0xcb, 0xc0 | bit | reg}, nil
		} else {
			return nil, insn.expectedNumberArgs(2)
		}
	case "nop":
		return []uint8{0x00}, nil
	}

	insn.Err = errors.New(fmt.Sprintf("unknown instruction '%s'", insn.Name))
	return nil, insn
}

func asmCond(cond string) (uint8, error) {
	switch cond {
	case "nz":
		return 0 << 3, nil
	case "z":
		return 1 << 3, nil
	case "nc":
		return 2 << 3, nil
	case "c":
		return 3 << 3, nil
	default:
		validConds := []string{"nz", "z", "nc", "c"}
		return 0xff, errors.New(fmt.Sprintf("unknown condition '%s', expected %s", cond, validConds))
	}
}

func asmAddr8(addr string) (uint8, error) {
	if addr[0] == '(' && addr[len(addr)-1] == ')' {
		return asmUint8(addr[1 : len(addr)-1])
	} else {
		return 0xff, errors.New("expected address in parens")
	}
}

func asmAddr16(addr string) (uint16, error) {
	if addr[0] == '(' && addr[len(addr)-1] == ')' {
		return asmUint16(addr[1 : len(addr)-1])
	} else {
		return 0xff, errors.New("expected address in parens")
	}
}

func asmAddrReg(reg string) (uint8, error) {
	if reg[0] == '(' && reg[len(reg)-1] == ')' {
		switch reg[1 : len(reg)-1] {
		case "bc":
			return 0 << 4, nil
		case "de":
			return 1 << 4, nil
		}
	}
	validRegs := []string{"bc", "de"}
	return 0xff, errors.New(fmt.Sprintf("unknown register '%s', expected %s", reg, validRegs))
}

func asmUint16(num string) (uint16, error) {
	if num[0] == '$' {
		num, err := strconv.ParseUint(num[1:], 16, 16)
		return uint16(num), err
	} else {
		num, err := strconv.ParseUint(num, 10, 16)
		return uint16(num), err
	}
}

func asmUint8(num string) (uint8, error) {
	if num[0] == '$' {
		num, err := strconv.ParseUint(num[1:], 16, 8)
		return uint8(num), err
	} else {
		num, err := strconv.ParseUint(num, 10, 8)
		return uint8(num), err
	}
}

func asmInt8(num string) (int8, error) {
	if num[0] == '$' {
		num, err := strconv.ParseInt(num[1:], 16, 8)
		return int8(num), err
	} else {
		num, err := strconv.ParseInt(num, 10, 8)
		return int8(num), err
	}
}

func asmBit(num string) (uint8, error) {
	bit, err := strconv.ParseUint(num, 10, 8)
	if err != nil {
		return 0xff, err
	}
	if bit > 7 {
		return 0xff, errors.New("bit value must be 0..7 inclusive")
	}
	return uint8(bit) << 3, nil
}

func asmReg16PushPop(reg string) (uint8, error) {
	switch reg {
	case "bc":
		return 0 << 4, nil
	case "de":
		return 1 << 4, nil
	case "hl":
		return 2 << 4, nil
	case "af":
		return 3 << 4, nil
	default:
		validRegs := []string{"bc", "de", "hl", "af"}
		return 0xff, errors.New(fmt.Sprintf("unknown register '%s', expected %s", reg, validRegs))
	}
}

func asmReg16(reg string) (uint8, error) {
	switch reg {
	case "bc":
		return 0 << 4, nil
	case "de":
		return 1 << 4, nil
	case "hl":
		return 2 << 4, nil
	case "sp":
		return 3 << 4, nil
	default:
		validRegs := []string{"bc", "de", "hl", "sp"}
		return 0xff, errors.New(fmt.Sprintf("unknown register '%s', expected %s", reg, validRegs))
	}
}

func asmReg8Hi(reg string) (uint8, error) {
	out, err := asmReg8Lo(reg)
	if err != nil {
		return 0xff, err
	}
	return out << 3, nil
}

func asmReg8Lo(reg string) (uint8, error) {
	switch reg {
	case "b":
		return 0, nil
	case "c":
		return 1, nil
	case "d":
		return 2, nil
	case "e":
		return 3, nil
	case "h":
		return 4, nil
	case "l":
		return 5, nil
	case "(hl)":
		return 6, nil
	case "a":
		return 7, nil
	default:
		validRegs := []string{"b", "c", "d", "e", "h", "l", "(hl)", "a"}
		return 0xff, errors.New(fmt.Sprintf("unknown register '%s', expected %s", reg, validRegs))
	}
}
