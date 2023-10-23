package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"math"
	"os"
	"strconv"
	"strings"
)

const MAXopcodeSize = 11

// Struct to represent an instruction with its mnemonic and type
type Instruction struct {
	Mnemonic string
	Type     string
}

// Map to associate binary opcodes with their instruction mnemonics and types
var opcodeMap = map[string]Instruction{
	// 6-bit opcodes
	"000101": {"B", "B"},

	// 8-bit opcodes
	"10110100": {"CBZ", "CB"},
	"10110101": {"CBNZ", "CB"},

	// 9-bit opcodes
	"110100101": {"MOVZ", "IM"},
	"111100101": {"MOVK", "IM"},

	// 10-bit opcodes
	"1001000100": {"ADDI", "I"},
	"1101000100": {"SUBI", "I"},

	// 11-bit opcodes
	"10001010000": {"AND", "R"},
	"10001011000": {"ADD", "R"},
	"10101010000": {"ORR", "R"},
	"11001011000": {"SUB", "R"},
	"11010011010": {"LSR", "R"},
	"11010011011": {"LSL", "R"},
	"11111000000": {"STUR", "D"},
	"11111000010": {"LDUR", "D"},
	"11010011100": {"ASR", "R"},
	"11101010000": {"EOR", "R"},

	// BREAK code
	"11111110110111101111111111100111": {"BREAK", "BREAK"},
}

func main() {
	// Define flags
	inputFile := flag.String("i", "", "Input file name")
	outputFile := flag.String("o", "", "Output file name")

	// Parse flags
	flag.Parse()

	// Check if both inputFile and outputFile have values
	if *inputFile == "" || *outputFile == "" {
		fmt.Println("Both input and output file names must be provided")
		os.Exit(1)
	}

	memCounter := 96

	// Open the input file for reading
	openfile, err := os.Open(*inputFile)
	if err != nil {
		log.Fatal(err)
	}
	defer openfile.Close()

	outFileNameWithSuffix := *outputFile + "_dis.txt"
	outFile, err := os.Create(outFileNameWithSuffix)
	if err != nil {
		log.Fatal(err)
	}
	defer outFile.Close()

	scanner := bufio.NewScanner(openfile)
	for scanner.Scan() {
		fullline := scanner.Text()
		result := defineOpcode(fullline, &memCounter)
		memCounter += 4

		// Write the result to the output file
		_, err := outFile.WriteString(result + "\n")
		if err != nil {
			log.Fatal(err)
		}
	}

	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}
}

func defineOpcode(line string, memCounter *int) string {

	line = strings.ReplaceAll(line, " ", "")
	var opcode string = ""
	var exists bool
	var inst Instruction

	// Ensure the line is long enough to contain the opcode
	if len(line) >= 6 { // Minimum opcode size is 6

		if len(line) >= 32 {
			opcode = line[:32]
			inst, exists = opcodeMap[opcode]
		}

		// Check for 11-bit opcode
		if !exists && len(line) >= 11 {
			opcode = line[:11]
			inst, exists = opcodeMap[opcode]
		}

		// If not found, check for 10-bit opcode
		if !exists && len(line) >= 10 {
			opcode = line[:10]
			inst, exists = opcodeMap[opcode]
		}

		// If not found, check for 9-bit opcode
		if !exists && len(line) >= 9 {
			opcode = line[:9]
			inst, exists = opcodeMap[opcode]
		}

		// If not found, check for 8-bit opcode
		if !exists && len(line) >= 8 {
			opcode = line[:8]
			inst, exists = opcodeMap[opcode]
		}

		// If not found, check for 6-bit opcode
		if !exists && len(line) >= 6 {
			opcode = line[:6]
			inst, exists = opcodeMap[opcode]
		}

		// Check if the opcode exists in the opcodeMap
		if exists {
			// Determine the type of instruction and extract relevant bits
			switch inst.Type {

			case "R":
				rm := extractBits(line, 11, 15)
				rn := extractBits(line, 22, 26)
				rd := extractBits(line, 27, 31)
				switch inst.Mnemonic {
				case "LSR", "LSL":
					imm := extractBits(line, 16, 21)
					return fmt.Sprintf("%s %s %s %s %s"+"\t%d\t%s\tR%d, R%d, #%d", line[:11], line[11:16], line[16:22], line[22:27], line[27:], *memCounter, inst.Mnemonic, rd, rn, imm)
				default:
					return fmt.Sprintf("%s %s %s %s %s"+"\t%d\t%s\tR%d, R%d, R%d", line[:11], line[11:16], line[16:22], line[22:27], line[27:], *memCounter, inst.Mnemonic, rd, rn, rm)
				}

			case "CB":
				imm := extractBits(line, 8, 26)
				rt := extractBits(line, 27, 31)

				return fmt.Sprintf("%s %s %s"+"\t%d\t%s\tR%d, #%d", line[:8], line[8:27], line[27:], *memCounter, inst.Mnemonic, rt, imm)

			case "I":
				imm := extractBits(line, 10, 21)
				rn := extractBits(line, 22, 26)
				rd := extractBits(line, 27, 31)

				return fmt.Sprintf("%s %s %s %s"+"\t%d\t%s\tR%d, R%d, #%d", line[:10], line[10:22], line[22:27], line[27:], *memCounter, inst.Mnemonic, rd, rn, imm)

			case "IM":
				immlo := extractBits(line, 9, 10)
				immhi := extractBits(line, 11, 26)
				rd := extractBits(line, 27, 31)
				shiftAmount := immlo * 16
				if inst.Mnemonic == "MOVZ" {
					return fmt.Sprintf("%s %s %s %s"+"\t%d\t%s\tR%d, %d, LSL %d", line[:9], line[9:11], line[11:27], line[27:], *memCounter, inst.Mnemonic, rd, immhi, shiftAmount)
				} else if inst.Mnemonic == "MOVK" {
					return fmt.Sprintf("%s %s %s %s"+"\t%d\t%s\tR%d, %d, LSL %d", line[:9], line[9:11], line[11:27], line[27:], *memCounter, inst.Mnemonic, rd, immhi, shiftAmount)
				}

			case "D":
				imm := extractBits(line, 11, 19)
				rn := extractBits(line, 22, 26)
				rt := extractBits(line, 27, 31)
				return fmt.Sprintf("%s %s %s %s %s"+"\t%d\t%s\tR%d, [R%d, #%d]", line[:11], line[11:20], line[20:22], line[22:27], line[27:], *memCounter, inst.Mnemonic, rt, rn, imm)

			case "B":
				opcodePart := line[:6]
				rawOffset := extractBits(line, 7, 31)
				signBit := extractBits(line, 6, 6)
				signSym := ""

				//If the sign bit is 1, it's a negative number
				if signBit == 1 {
					rawOffset = -(^(rawOffset - 1))
					signSym = "-"
				}

				// Calculate the actual offset for display
				displayOffset := rawOffset

				return fmt.Sprintf("%s %s"+"\t%d\t%s\t#%s%d", opcodePart, line[6:], *memCounter, inst.Mnemonic, signSym, displayOffset)

			case "N/A":
				return fmt.Sprintf("%s"+"\t%d\tNOP", line, *memCounter)
			case "BREAK":
				return fmt.Sprintf("%s"+"\t%d\t%s", line, *memCounter, inst.Mnemonic)
			}

		}
	} else {

		return fmt.Sprintf("Unknown instruction with opcode: %s at address %d", opcode, *memCounter)
	}

	if len(line) == 32 {
		decInt := twosComplement(line)
		return fmt.Sprintf("%s"+"\t%d\t%d", line, *memCounter, decInt)
	}

	return fmt.Sprintf("Invalid instruction at address %d", *memCounter)
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

func binToDec(binline string) int {

	index := 31
	decimalNum := 0

	tempdecimalNum := 0
	for index != 0 {
		for index == 31 {
			templine := binline[index-1 : index]
			ttempline, _ := strconv.Atoi(templine)
			tempdecimalNum = tempdecimalNum + (ttempline * int(math.Pow(2, float64(index))))
			index--
		}
		templine := binline[index-1 : index]
		ttempline, _ := strconv.Atoi(templine)
		decimalNum = decimalNum + (ttempline * int(math.Pow(2, float64(index))))
		index--

	}
	decimalNum = decimalNum - tempdecimalNum
	return decimalNum
}

func twosComplement(binStr string) int {

	num, _ := strconv.ParseInt(binStr, 2, 64)

	num = (1 << len(binStr)) - num

	return -int(num)
}
