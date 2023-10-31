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
	"00000000000000000000000000000000": {"NOP", "NOP"},
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
	cycleCounter := 1

	// Open the input file for reading
	openfile, err := os.Open(*inputFile)
	if err != nil {
		log.Fatal(err)
	}
	defer openfile.Close()

	outFileNameWithSuffix := *outputFile + "_dis.txt"
	outFileSimNameWithSuffix := *outputFile + "_sim.txt"
	outFile, err := os.Create(outFileNameWithSuffix)
	if err != nil {
		log.Fatal(err)
	}
	defer outFile.Close()

	outFileSim, err := os.Create(outFileSimNameWithSuffix)
	if err != nil {
		log.Fatal(err)
	}
	defer outFileSim.Close()

	scanner := bufio.NewScanner(openfile)
	for scanner.Scan() {
		fullline := scanner.Text()
		result := defineOpcode(fullline, &memCounter)
		memCounter += 4
		cycleCounter += 4

		// Write the result to the output file
		_, err := outFile.WriteString(result + "\n")
		if err != nil {
			log.Fatal(err)
		}
		_, err = outFileSim.WriteString("====================" + "\n")
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
				case "LSR", "LSL", "ASR":
					imm := extractBits(line, 16, 21)
					return fmt.Sprintf("%s %s %s %s %s \t%d \t%s \tR%d, R%d, #%d", line[:11], line[11:16], line[16:22], line[22:27], line[27:], *memCounter, inst.Mnemonic, rd, rn, imm)
				default:
					return fmt.Sprintf("%s %s %s %s %s \t%d \t%s \tR%d, R%d, R%d", line[:11], line[11:16], line[16:22], line[22:27], line[27:], *memCounter, inst.Mnemonic, rd, rn, rm)
				}

			case "CB":
				imm := extractBits(line, 8, 26)
				rt := extractBits(line, 27, 31)
				var snum int32

				binaryImm := fmt.Sprintf("%019b", imm) // Convert imm to a 19-bit binary string
				if binaryImm[0] == '1' {               // Check if the most significant bit is 1
					// Convert from two's complement to positive binary number
					invertedBinaryImm := ""
					for _, bit := range binaryImm {
						if bit == '0' {
							invertedBinaryImm += "1"
						} else {
							invertedBinaryImm += "0"
						}
					}
					positiveBinaryImm := addBinary(invertedBinaryImm, "1")
					snum = -int32(binaryToDecimal(positiveBinaryImm))
				} else {
					snum = int32(binaryToDecimal(binaryImm))
				}

				return fmt.Sprintf("%s %s %s  \t%d \t%s \tR%d, #%d", line[:8], line[8:27], line[27:], *memCounter, inst.Mnemonic, rt, snum)

			case "I":
				imm := extractBits(line, 10, 21)
				rn := extractBits(line, 22, 26)
				rd := extractBits(line, 27, 31)
				negBitMask := 0x800 // figure out if 12 bit num is neg
				extendMask := 0xFFFFF000
				var simm int32
				simm = int32(imm)
				if (negBitMask & imm) > 0 { // is it?
					imm = imm | extendMask // if so extend with 1's
					imm = imm ^ 0xFFFFFFFF // 2s comp
					simm = int32(imm + 1)
					simm = simm * -1 // add neg sign
				}

				return fmt.Sprintf("%s %s %s %s \t%d  \t%s \tR%d, R%d, #%d", line[:10], line[10:22], line[22:27], line[27:], *memCounter, inst.Mnemonic, rd, rn, simm)

			case "IM":
				immlo := extractBits(line, 9, 10)
				immhi := extractBits(line, 11, 26)
				rd := extractBits(line, 27, 31)
				shiftAmount := immlo * 16
				if inst.Mnemonic == "MOVZ" {
					return fmt.Sprintf("%s %s %s %s \t%d \t%s \tR%d, %d, LSL %d", line[:9], line[9:11], line[11:27], line[27:], *memCounter, inst.Mnemonic, rd, immhi, shiftAmount)
				} else if inst.Mnemonic == "MOVK" {
					return fmt.Sprintf("%s %s %s %s \t%d \t%s \tR%d, %d, LSL %d", line[:9], line[9:11], line[11:27], line[27:], *memCounter, inst.Mnemonic, rd, immhi, shiftAmount)
				}

			case "D":
				imm := extractBits(line, 11, 19)
				rn := extractBits(line, 22, 26)
				rt := extractBits(line, 27, 31)
				return fmt.Sprintf("%s %s %s %s %s \t%d \t%s \tR%d, [R%d, #%d]", line[:11], line[11:20], line[20:22], line[22:27], line[27:], *memCounter, inst.Mnemonic, rt, rn, imm)

			case "B":
				opcodePart := line[:6]
				rawOffset := extractBits(line, 7, 31)
				var snum int32

				binaryOffset := fmt.Sprintf("%025b", rawOffset) // Convert rawOffset to a 25-bit binary string
				if binaryOffset[0] == '1' {                     // Check if the most significant bit is 1
					// Convert from two's complement to positive binary number
					invertedBinaryOffset := ""
					for _, bit := range binaryOffset {
						if bit == '0' {
							invertedBinaryOffset += "1"
						} else {
							invertedBinaryOffset += "0"
						}
					}
					positiveBinaryOffset := addBinary(invertedBinaryOffset, "1")
					snum = -int32(binaryToDecimal(positiveBinaryOffset))
				} else {
					snum = int32(binaryToDecimal(binaryOffset))
				}

				return fmt.Sprintf("%s %s   \t%d \t%s   \t#%d", opcodePart, line[6:], *memCounter, inst.Mnemonic, snum)

			case "NOP":
				return fmt.Sprintf("%s\t%d\tNOP", line, *memCounter)
			case "N/A":
				return fmt.Sprintf("%s \t%d \tNOP", line, *memCounter)
			case "BREAK":
				return fmt.Sprintf("%s %s %s %s %s %s \t%d \t%s", line[:8], line[8:11], line[11:16], line[16:21], line[21:26], line[26:], *memCounter, inst.Mnemonic)
			}

		}
	} else {

		return fmt.Sprintf("Unknown instruction with opcode: %s at address %d", opcode, *memCounter)
	}

	// Data after break
	if len(line) == 32 {
		binaryData := line        // Assuming the data after "BREAK" is the entire line
		if binaryData[0] == '1' { // Check if the most significant bit is 1
			// Convert from two's complement to positive binary number
			invertedBinaryData := ""
			for _, bit := range binaryData {
				if bit == '0' {
					invertedBinaryData += "1"
				} else {
					invertedBinaryData += "0"
				}
			}
			positiveBinaryData := addBinary(invertedBinaryData, "1")
			decInt := -binaryToDecimal(positiveBinaryData)
			return fmt.Sprintf("%s \t%d \t%d", line, *memCounter, decInt)
		} else {
			decInt := binaryToDecimal(binaryData)
			return fmt.Sprintf("%s \t%d \t%d", line, *memCounter, decInt)
		}
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

func twosComplement(binStr string, bitSize int) int {

	num, _ := strconv.ParseInt(binStr, 2, bitSize)

	num = (1 << len(binStr)) - num

	return -int(num)
}

func addBinary(a, b string) string {
	maxLength := max(len(a), len(b))
	a = padLeft(a, '0', maxLength)
	b = padLeft(b, '0', maxLength)

	carry := 0
	result := ""
	for i := maxLength - 1; i >= 0; i-- {
		bitA := int(a[i] - '0')
		bitB := int(b[i] - '0')
		sum := bitA + bitB + carry
		result = strconv.Itoa(sum%2) + result
		carry = sum / 2
	}
	if carry > 0 {
		result = "1" + result
	}
	return result
}

func binaryToDecimal(binaryStr string) int {
	result := 0
	length := len(binaryStr)
	for i, bit := range binaryStr {
		if bit == '1' {
			result += 1 << (length - 1 - i)
		}
	}
	return result
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func padLeft(str string, padChar byte, length int) string {
	for len(str) < length {
		str = string(padChar) + str
	}
	return str
}

// test line
