package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
)

var txtfile = "rtest1_bin_untranslated.txt"

const MAXopcodeSize = 11

// Struct to represent an instruction with its mnemonic and type
type Instruction struct {
	Mnemonic string
	Type     string
}

// Map to associate binary opcodes with their instruction mnemonics and types
var opcodeMap = map[string]Instruction{
	"000101":      {"B", "B"},
	"10001010000": {"AND", "R"},
	"10001011000": {"ADD", "R"},
	"1001000100":  {"ADDI", "I"},
	"10101010000": {"ORR", "R"},
	"10110100":    {"CBZ", "CB"},
	"10110101":    {"CBNZ", "CB"},
	"11001011000": {"SUB", "R"},
	"1101000100":  {"SUBI", "I"},
	"110100101":   {"MOVZ", "IM"},
	"111100101":   {"MOVK", "IM"},
	"11010011010": {"LSR", "R"},
	"11010011011": {"LSL", "R"},
	"11111000000": {"STUR", "D"},
	"11111000010": {"LDUR", "D"},
	"11010011100": {"ASR", "R"},
	"0":           {"NOP", "N/A"},
	"11101010000": {"EOR", "R"},
}

func main() {

	openfile, err := os.Open(txtfile)
	if err != nil {
		log.Fatal(err)
	}
	// Close file
	defer openfile.Close()

	scanner := bufio.NewScanner(openfile)
	for scanner.Scan() {
		// Current line
		fullline := scanner.Text()
		// Translate the binary instruction to its assembly representation
		defineOpcode(fullline)
	}
	// Check for any errors that occurred while reading the file
	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}
}

func defineOpcode(line string) {
	line = strings.ReplaceAll(line, " ", "")
	// Ensure the line is long enough to contain the opcode
	if len(line) >= MAXopcodeSize {
		// Extract the opcode from the line
		opcode := line[:MAXopcodeSize]
		// Check if the opcode exists in the opcodeMap
		if inst, exists := opcodeMap[opcode]; exists {
			// Determine the type of instruction and extract relevant bits
			switch inst.Type {
			case "R":
				// Extract register values for R-type instructions
				rm := extractBits(line, 11, 15)
				rn := extractBits(line, 22, 26)
				rd := extractBits(line, 27, 31)
				// Print the assembly representation of the instruction
				fmt.Printf("%s R%d, R%d, R%d\n", inst.Mnemonic, rd, rn, rm)
				// TODO: Handle other instruction types (I, B, CB, etc.)
			}
		} else {

			fmt.Println("Unknown instruction")
		}
	}
}

func extractBits(line string, start, end int) int {
	// Extract a substring of bits from the line based on the provided start and end positions
	bits := line[start : end+1]
	// Convert the binary string to an integer
	value, err := strconv.ParseInt(bits, 2, 64)
	if err != nil {

		log.Fatal(err)
	}
	return int(value)
}
