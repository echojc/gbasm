package main

import (
	"fmt"
	"log"
)

func main() {

	//if len(os.Args) < 2 {
	//	log.Fatalf("Usage: %s <file>\n", os.Args[0])
	//}

	//file, err := os.Open(os.Args[1])
	//if err != nil {
	//	log.Fatalf("Could not open file '%s'\n", os.Args[1])
	//}
	//defer file.Close()

	//scanner := bufio.NewScanner(file)
	//lines := make([]string, 0)
	//for scanner.Scan() {
	//	lines = append(lines, scanner.Text())
	//}
	//if err := scanner.Err(); err != nil {
	//	log.Fatalf("Error reading file: %v\n", err)
	//}

	lines := []string{
		":rst_00",
		"jp main",
		":main",
		"ld a, $03",
		"di",
		"ldh ($ff), a",
		"ld a, $40",
		"ldh ($41), a",
		"xor a",
		"ldh ($40), a",
		":loop",
		"ldh a, ($44)",
		"cp $94",
		"jr nz, loop",
		"halt",
	}

	unit, err := Parse(lines)
	if err != nil {
		log.Fatalln(err)
	}

	bytes, err := Compile(unit)
	if err != nil {
		log.Fatalln(err)
	}

	for i, b := range bytes {
		fmt.Printf("%02x ", b)
		if (i+1)%0x10 == 0 {
			fmt.Println()
		}
	}
	fmt.Println()
}
