package main

import (
	"errors"
	"fmt"
)

func Compile(unit *Unit) ([]uint8, error) {
	if _, found := unit.Sections["main"]; !found {
		return nil, errors.New("label 'main' is not defined")
	}

	labelAddrs := map[string]uint16{
		"rst_00":     0x0000,
		"rst_08":     0x0008,
		"rst_10":     0x0010,
		"rst_18":     0x0018,
		"rst_20":     0x0020,
		"rst_28":     0x0028,
		"rst_30":     0x0030,
		"rst_38":     0x0038,
		"int_vblank": 0x0040,
		"int_lcdc":   0x0048,
		"int_timer":  0x0050,
		"int_serial": 0x0058,
		"int_keys":   0x0060,
		"main":       0x0150,
	}

	// enough space for all header stuff
	output := make([]uint8, 0x0100)

	// compile special sections
	for _, label := range []string{
		"rst_00",
		"rst_08",
		"rst_10",
		"rst_18",
		"rst_20",
		"rst_28",
		"rst_30",
		"rst_38",
		"int_vblank",
		"int_lcdc",
		"int_timer",
		"int_serial",
		"int_keys",
	} {
		bytes, err := compileSpecial(unit, label)
		if err != nil {
			return nil, err
		}
		for i := 0; i < len(bytes); i++ {
			output[int(labelAddrs[label])+i] = bytes[i]
		}
	}

	// copy header
	header := generateHeader()
	output = append(output, header[:]...)

	// generate main
	bytes, err := compileSection(unit, "main")
	if err != nil {
		return nil, err
	}
	output = append(output, bytes...)

	// generate everything else
	for _, label := range unit.Labels {
		// skip already compiled sections
		if _, found := labelAddrs[label]; found {
			continue
		}

		bytes, err := compileSection(unit, label)
		if err != nil {
			return nil, err
		}
		output = append(output, bytes...)
	}

	// calculate checksum
	var checksum uint = 0
	for _, b := range output {
		checksum += uint(b)
	}
	output[0x014e] = uint8((checksum >> 8) & 0xff)
	output[0x014f] = uint8(checksum & 0xff)

	return output, nil
}

func compileSection(unit *Unit, label string) ([]uint8, error) {
	if section, found := unit.Sections[label]; found {
		bytes, err := Assemble(section.Insns)
		return bytes, err
	} else {
		return nil, errors.New(fmt.Sprintf("unknown label '%s'", label))
	}
}

func compileSpecial(unit *Unit, label string) ([8]uint8, error) {
	output := [8]uint8{}

	if section, found := unit.Sections[label]; found {
		bytes, err := Assemble(section.Insns)
		if err != nil {
			return output, err
		}

		for i := 0; i < 8 && i < len(bytes); i++ {
			output[i] = bytes[i]
		}
	}

	return output, nil
}

func generateHeader() [0x50]uint8 {
	return [0x50]uint8{
		0x00, 0xc3, 0x50, 0x01, 0xce, 0xed, 0x66, 0x66, 0xcc, 0x0d, 0x00, 0x0b, 0x03, 0x73, 0x00, 0x83,
		0x00, 0x0c, 0x00, 0x0d, 0x00, 0x08, 0x11, 0x1f, 0x88, 0x89, 0x00, 0x0e, 0xdc, 0xcc, 0x6e, 0xe6,
		0xdd, 0xdd, 0xd9, 0x99, 0xbb, 0xbb, 0x67, 0x63, 0x6e, 0x0e, 0xec, 0xcc, 0xdd, 0xdc, 0x99, 0x9f,
		0xbb, 0xb9, 0x33, 0x3e, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0xe7, 0x00, 0x00,
	}
}
