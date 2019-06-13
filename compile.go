package main

import (
	"errors"
	"fmt"
)

type LabelOffset struct {
	Label       string
	Offset      uint16
	InsnOffsets []int
}

func Compile(unit *Unit) ([]uint8, error) {
	if _, found := unit.Sections["main"]; !found {
		return nil, errors.New("label 'main' is not defined")
	}

	// for resolving labels
	labelOffsets := map[string]LabelOffset{}

	// enough space for all header stuff
	output := make([]uint8, 0x0150)

	// compile special sections
	for idx, label := range []string{
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
		bytes, insnOffsets, err := compileSpecial(unit, label)
		if err != nil {
			return nil, err
		}

		labelOffset := idx * 0x08
		for i := 0; i < len(bytes); i++ {
			output[labelOffset+i] = bytes[i]
		}

		labelOffsets[label] = LabelOffset{
			label,
			uint16(labelOffset),
			insnOffsets,
		}
	}

	// copy header
	header := generateHeader()
	for i := 0; i < len(header); i++ {
		output[0x0100+i] = header[i]
	}

	// generate main
	bytes, insnOffsets, err := compileSection(unit, "main")
	if err != nil {
		return nil, err
	}
	labelOffsets["main"] = LabelOffset{
		"main",
		0x0150,
		insnOffsets,
	}
	output = append(output, bytes...)

	// generate everything else
	for _, label := range unit.Labels {
		// skip already compiled sections
		if _, found := labelOffsets[label]; found {
			continue
		}

		bytes, insnOffsets, err := compileSection(unit, label)
		if err != nil {
			return nil, err
		}

		// align a section to the closest 0x100 (for lookup tables, etc.)
		offset := uint16(len(output))
		if unit.Sections[label].IsAligned && (offset&0x00ff) != 0 {
			alignedOffset := (offset + 0x100) & 0xff00
			output = append(output, make([]uint8, alignedOffset-offset)...)
			offset = alignedOffset
		}

		labelOffsets[label] = LabelOffset{
			label,
			offset,
			insnOffsets,
		}
		output = append(output, bytes...)
	}

	// resolve labels
	for _, labelUsage := range unit.LabelUsages {
		targetLabel := labelUsage.TargetLabel
		targetAddr := uint16(labelOffsets[targetLabel].Offset)

		usage := labelOffsets[labelUsage.SourceSection]
		usageOffset := usage.Offset + uint16(usage.InsnOffsets[labelUsage.SourceInsnIndex])

		insn := unit.Sections[labelUsage.SourceSection].Insns[labelUsage.SourceInsnIndex]
		if insn.Name == "jr" {
			// calculate relative
			startAddr := usageOffset + 2
			delta := int(targetAddr) - int(startAddr)

			if delta > 127 || delta < -128 { // int8 range
				insn.Err = errors.New(fmt.Sprintf("target label '%s' is out of range (%d)", targetLabel, delta))
				return nil, &insn
			}

			output[usageOffset+1] = uint8(int8(delta))
		} else {
			// inject absolute
			output[usageOffset+1] = uint8(targetAddr & 0xff)
			output[usageOffset+2] = uint8(targetAddr >> 8)
		}
	}

	// calculate checksum
	var checksum uint = 0
	for _, b := range output {
		checksum += uint(b)
	}
	output[0x014e] = uint8(checksum >> 8)
	output[0x014f] = uint8(checksum & 0xff)

	return output, nil
}

func compileSection(unit *Unit, label string) ([]uint8, []int, error) {
	if section, found := unit.Sections[label]; found {
		output := make([]uint8, len(section.Data))

		bytes, offsets, err := Assemble(section.Insns)
		if err != nil {
			return nil, nil, err
		}

		// prepend data block if necessary
		if len(section.Data) > 0 {
			copy(output, section.Data)
			for i, offset := range offsets {
				offsets[i] = offset + len(section.Data)
			}
		}

		output = append(output, bytes...)
		return output, offsets, nil
	} else {
		return nil, nil, errors.New(fmt.Sprintf("unknown label '%s'", label))
	}
}

func compileSpecial(unit *Unit, label string) ([]uint8, []int, error) {
	output := []uint8{}
	var offsets []int = nil

	if section, found := unit.Sections[label]; found {
		bytes, _offsets, err := Assemble(section.Insns)
		offsets = _offsets
		if err != nil {
			return output, nil, err
		}

		output = bytes
	}

	return output, offsets, nil
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
