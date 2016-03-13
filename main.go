package main

import (
	"bufio"
	"log"
	"os"
	"strings"
)

func main() {
	log.SetFlags(0)

	if len(os.Args) < 2 {
		log.Fatalf("Usage: %s <input> [<output>]\n", os.Args[0])
	}

	inputFilename := os.Args[1]
	input, err := os.Open(inputFilename)
	if err != nil {
		log.Fatalf("Could not open input file '%s'\n", inputFilename)
	}
	defer input.Close()

	outputFilename := ""
	if len(os.Args) > 2 {
		outputFilename = os.Args[2]
	} else {
		if i := strings.LastIndex(inputFilename, "."); i >= 0 {
			outputFilename = inputFilename[0:i] + ".bin"
		} else {
			outputFilename = inputFilename + ".bin"
		}
	}

	output, err := os.OpenFile(outputFilename, os.O_CREATE|os.O_TRUNC|os.O_RDWR, 0664)
	if err != nil {
		log.Fatalf("Could not open output file '%s'\n", outputFilename)
	}
	defer output.Close()

	scanner := bufio.NewScanner(input)
	lines := make([]string, 0)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	if err := scanner.Err(); err != nil {
		log.Fatalf("Error reading file: %v\n", err)
	}

	unit, err := Parse(lines)
	if err != nil {
		log.Fatalln(err)
	}

	bytes, err := Compile(unit)
	if err != nil {
		log.Fatalln(err)
	}

	count, err := output.Write(bytes)
	if err != nil {
		log.Fatalln(err)
	}

	if count < 0x8000 {
		padding := make([]uint8, 0x8000-count)
		output.Write(padding)
	}
}
