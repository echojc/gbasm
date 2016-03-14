package main

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"regexp"
	"strings"
	"unicode"
)

type Section struct {
	Label      string
	LineNumber uint
	Data       []uint8
	Insns      []Insn
}

type Insn struct {
	Name       string
	Args       []string
	LineNumber uint
	Err        error
}

type LabelUsage struct {
	TargetLabel     string
	SourceSection   string
	SourceInsnIndex int
}

type Unit struct {
	Sections    map[string]*Section
	Labels      []string
	LabelUsages []*LabelUsage
}

var labelRegex = regexp.MustCompile("^[a-z_][a-z0-9_]*$")
var dataLabelReplaceRegex = regexp.MustCompile("[^a-z0-9_]+")

func Parse(lines []string) (*Unit, error) {
	var currentSection *Section

	sections := make(map[string]*Section)
	definedLabels := make([]string, 0)
	labelUsages := make([]*LabelUsage, 0)

	for i, text := range lines {
		lineNumber := uint(i + 1)

		// drop comments
		if i := strings.Index(text, ";"); i >= 0 {
			text = text[0:i]
		}

		text = strings.TrimSpace(strings.ToLower(text))
		if text == "" {
			continue
		}

		if text[0] == '.' { // label
			label := text[1:]
			if _, alreadyExists := sections[label]; alreadyExists {
				return nil, errors.New(fmt.Sprintf("duplicate label '%s' (labels are case insensitive)", label))
			}

			section, err := newSection(label)
			if err != nil {
				return nil, err
			}

			section.LineNumber = lineNumber
			definedLabels = append(definedLabels, label)

			if currentSection != nil {
				sections[currentSection.Label] = currentSection
			}
			currentSection = section

		} else if text[0] == '<' { // data
			dataFile, err := os.Open(text[1:])
			if err != nil {
				return nil, err
			}
			defer dataFile.Close()

			// the '.' is intentional, and becomes a '_' after the regex replace
			label := "data." + text[1:]
			label = dataLabelReplaceRegex.ReplaceAllLiteralString(label, "_")
			if _, alreadyExists := sections[label]; alreadyExists {
				return nil, errors.New(fmt.Sprintf("duplicate label '%s' (labels are case insensitive)", label))
			}

			section, err := newSection(label)
			if err != nil {
				return nil, err
			}

			data, err := ioutil.ReadAll(dataFile)
			if err != nil {
				return nil, err
			}

			section.Data = data
			section.LineNumber = lineNumber
			definedLabels = append(definedLabels, label)

			if currentSection != nil {
				sections[currentSection.Label] = currentSection
			}
			currentSection = section

		} else if currentSection == nil {
			return nil, errors.New("all asm must be under some label")
		} else {
			insn := ParseInsn(text, lineNumber)

			// replace label usage with placeholder
			for argIndex, targetLabel := range insn.Args {
				if !isSpecialName(targetLabel) && isValidLabel(targetLabel) {
					labelUsages = append(labelUsages, &LabelUsage{
						targetLabel,
						currentSection.Label,
						len(currentSection.Insns),
					})

					// replace with appropriate placeholder
					switch insn.Name {
					case "jr":
						insn.Args[argIndex] = "$66"
					case "ld":
						if len(insn.Args) == 2 && isReg16(insn.Args[0]) {
							insn.Args[argIndex] = "$6666"
						} else {
							insn.Args[argIndex] = "($6666)"
						}
					default:
						insn.Args[argIndex] = "$6666"
					}
				}
			}

			currentSection.Insns = append(currentSection.Insns, insn)
		}
	}

	if currentSection != nil {
		sections[currentSection.Label] = currentSection
	}

	if len(sections) == 0 {
		return nil, errors.New("there was nothing to parse")
	}

	// validating labels
	missingLabels := make([]string, 0)
	for _, labelUsage := range labelUsages {
		usedLabel := labelUsage.TargetLabel

		if _, found := sections[usedLabel]; !found {
			missingLabels = append(missingLabels, usedLabel)
		}
	}
	if len(missingLabels) > 0 {
		return nil, errors.New(fmt.Sprintf("found undefined labels %s", missingLabels))
	}

	return &Unit{sections, definedLabels, labelUsages}, nil
}

func ParseInsn(line string, num uint) Insn {
	spaceOrComma := func(c rune) bool { return unicode.IsSpace(c) || c == ',' }
	parts := strings.FieldsFunc(line, spaceOrComma)

	insn := Insn{}
	insn.Name = parts[0]
	insn.Args = parts[1:]
	insn.LineNumber = num
	return insn
}

func (i *Insn) Error() string {
	return fmt.Sprintf("%d: %s", i.LineNumber, i.Err.Error())
}

func (i *Insn) expectedNumberArgs(expected ...uint) error {
	str := fmt.Sprintf("'%s' has wrong number of args, expected %d", i.Name, expected)
	i.Err = errors.New(str)
	return i
}

func newSection(label string) (*Section, error) {
	if isSpecialName(label) {
		return nil, errors.New(fmt.Sprintf("'%s' is reserved and can't be used as a label name", label))
	} else if !isValidLabel(label) {
		return nil, errors.New(fmt.Sprintf("label '%s' is invalid (alphanumeric + underscore)", label))
	}

	section := new(Section)
	section.Label = label
	return section, nil
}

func isValidLabel(name string) bool {
	return labelRegex.MatchString(name)
}

func isSpecialName(name string) bool {
	return "b" == name ||
		"c" == name ||
		"d" == name ||
		"e" == name ||
		"h" == name ||
		"l" == name ||
		"a" == name ||
		"af" == name ||
		"nz" == name ||
		"z" == name ||
		"nc" == name ||
		isReg16(name)
}

func isReg16(name string) bool {
	return "bc" == name || "de" == name || "hl" == name || "sp" == name
}
