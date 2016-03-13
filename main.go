package main

import "fmt"

func main() {

	//if len(os.Args) < 2 {
	//	log.Printf("Usage: %s <file>\n", os.Args[0])
	//	os.Exit(2)
	//}

	//file, err := os.Open(os.Args[1])
	//if err != nil {
	//	log.Printf("Could not open file '%s'\n", os.Args[1])
	//	os.Exit(1)
	//}
	//defer file.Close()

	//scanner := bufio.NewScanner(file)
	//lines := make([]string, 0)
	//for scanner.Scan() {
	//	lines = append(lines, scanner.Text())
	//}
	//if err := scanner.Err(); err != nil {
	//	log.Printf("Error reading file: %v\n", err)
	//	os.Exit(3)
	//}

	lines := []string{
		"ld a, $03",
		"di",
		"ldh ($ff), a",
		"ld a, $40",
		"ldh ($41), a",
		"xor a",
		"ldh ($40), a",
		"ldh a, ($44)",
		"cp $94",
		"jr nz, $-06",
		"halt",
	}

	bytes, err := Assemble(lines)
	if err != nil {
		fmt.Println(err)
	} else {
		for _, b := range bytes {
			fmt.Printf("%02x ", b)
		}
		fmt.Println()
	}
}
