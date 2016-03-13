package main

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
	"unicode"
)

type Section struct {
	Label      string
	LineNumber uint
	Insns      []Insn
}

type Insn struct {
	Name       string
	Args       []string
	LineNumber uint
	Err        error
}

type LabelLocation struct {
	Section   string
	InsnIndex int
}

type Unit struct {
	Sections    map[string]*Section
	Labels      []string
	LabelUsages map[string]*LabelLocation
}

func Parse(lines []string) (*Unit, error) {
	var currentSection *Section

	sections := make(map[string]*Section)
	definedLabels := make([]string, 0)
	labelUsages := make(map[string]*LabelLocation)

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

		if text[0] == ':' {
			if currentSection != nil {
				sections[currentSection.Label] = currentSection
			}

			label := strings.ToLower(text[1:])
			if isSpecialName(label) {
				return nil, errors.New(fmt.Sprintf("'%s' is reserved and can't be used as a label name", label))
			} else if !isValidLabel(label) {
				return nil, errors.New(fmt.Sprintf("label '%s' is invalid (alphanumeric + underscore)", label))
			} else if _, alreadyExists := sections[label]; alreadyExists {
				return nil, errors.New(fmt.Sprintf("duplicate label '%s' (labels are case insensitive)", label))
			}

			currentSection = new(Section)
			currentSection.Label = label
			currentSection.LineNumber = lineNumber
			definedLabels = append(definedLabels, label)

		} else if currentSection == nil {
			return nil, errors.New("all asm must be under some label")
		} else {
			insn := ParseInsn(text, lineNumber)

			// replace label usage with placeholder
			for argIndex, label := range insn.Args {
				if !isSpecialName(label) && isValidLabel(label) {
					labelUsages[label] = &LabelLocation{
						currentSection.Label,
						len(currentSection.Insns),
					}

					// replace with appropriate placeholder
					switch insn.Name {
					case "jr":
						insn.Args[argIndex] = "$66"
					case "ld":
						if len(insn.Args) == 2 && (insn.Args[0] == "bc" || insn.Args[0] == "de") {
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
	for usedLabel, _ := range labelUsages {
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

var labelRegex = regexp.MustCompile("^[a-z_][a-z0-9_]")

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
		"bc" == name ||
		"de" == name ||
		"hl" == name ||
		"sp" == name ||
		"af" == name ||
		"nz" == name ||
		"z" == name ||
		"nc" == name
}
