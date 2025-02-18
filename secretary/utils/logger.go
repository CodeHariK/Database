package utils

import (
	"encoding/json"
	"fmt"
	"math/rand/v2"
	"os"
	"runtime/debug"
	"strings"

	"golang.org/x/term"
)

const (
	COLORRESET = "\033[0m"

	RED = "\033[0;31m"
)

var colorIndex = 0

const (
	PROJECTNAME     = "/secretary/"
	PROJECTFUNCNAME = "/secretary."
)

func Log(msgs ...any) {
	if len(msgs) == 0 || (len(msgs) == 1 && msgs[0] == "\n") {
		fmt.Println()
		return
	}

	width, _, err := term.GetSize(int(os.Stdout.Fd()))
	if err != nil {
		width = 80 // Default width if terminal size can't be determined
	}

	colorIndex++
	randomColor := Ternary(
		colorIndex%2 == 1,
		nightColor(),
		lightColor())

	log := randomColor

	extracTrace := func(lines []string, i int) string {
		nameLoc := ""
		line := lines[i]

		if strings.Contains(line, PROJECTNAME) && strings.Contains(line, ".go") {
			parts := strings.Split(strings.Trim(line, "\t"), " ")
			if len(parts) > 0 {
				projectPart := strings.Split(parts[0], PROJECTNAME)
				if len(projectPart) > 1 {
					nameLoc += fmt.Sprint(projectPart[1], " ")
				}
			}

			if i > 0 { // Check index range
				prevParts := strings.Split(strings.Trim(lines[i-1], "\t"), " ")
				if len(prevParts) > 0 {
					funcPart := strings.Split(prevParts[0], PROJECTFUNCNAME)
					if len(funcPart) > 1 {
						funcName := strings.Split(funcPart[1], "(0x")
						if len(funcName) > 0 {
							nameLoc += fmt.Sprint(funcName[0], " ")
						}
					}
				}
			}
		}
		return nameLoc
	}

	{
		lines := strings.Split(string(debug.Stack()), "\n")
		loc := 0

		nameLoc := ""

		for i := range lines {
			if strings.Contains(lines[i], PROJECTNAME) && strings.Contains(lines[i], ".go") {
				if loc == 1 {
					nameLoc += extracTrace(lines, i)
				}
				loc++
			}
		}

		p := nameLoc
		if len(msgs) > 2 {
			p = padLine("> "+nameLoc, width-1, "-", false)
		}
		log += p
	}

	for i, msg := range msgs {
		if err, ok := msg.(error); ok {
			fmt.Fprintf(os.Stderr, "%s", "\n"+RED+err.Error()+"\n")

			lines := strings.Split(string(debug.Stack()), "\n")

			for i := range lines {
				l := extracTrace(lines, i)
				if len(l) > 0 {
					log += "\n" + l
				}
			}
		} else {
			if i%2 == 1 {
				log += " "
			}
			if i%2 == 0 && len(msgs) > 2 {
				log += "\n"
			}

			switch v := msg.(type) {
			case int, int8, int16, int32, int64,
				float32, float64,
				uint, uint8, uint16, uint32, uint64,
				string, []string,
				[]int, []float32:
				log += fmt.Sprint(msg)
			default:
				data, err := json.MarshalIndent(v, "", "  ")
				if err != nil {
					log += fmt.Sprint(msg)
				} else {
					log += string(data)
				}
			}
		}
	}

	log = processParagraph(log, len(randomColor), width) + COLORRESET

	fmt.Println(log)
}

func padLine(line string, width int, repeat string, suffix bool) string {
	lineLen := len(line)

	// Calculate the next multiple of width
	targetWidth := ((lineLen / width) + 1) * width

	// Calculate how many "_" are needed to reach targetWidth
	padding := targetWidth - lineLen

	if suffix {
		return line + strings.Repeat(repeat, padding)
	}
	return strings.Repeat(repeat, padding) + line
}

// Cleans and processes a paragraph
func processParagraph(paragraph string, colorlen int, width int) string {
	lines := strings.Split(paragraph, "\n") // Split paragraph by newlines
	for i, line := range lines {
		// fmt.Println(len(line), width)
		// fmt.Println(line)
		if i == 0 {
			lines[i] = padLine(line, width+colorlen, " ", true) // Pad each line to a multiple of width
		} else {
			lines[i] = padLine(line, width, " ", true) // Pad each line to a multiple of width
		}
		// fmt.Println(len(lines[i]), width)
		// fmt.Println(lines[i])
	}
	return strings.Join(lines, "\n") // Merge lines back
}

func lightColor() string {
	dlFR, dlFG, dlFB := randomColor(0, 30)
	dlBR, dlBG, dlBB := randomColor(180, 240)

	return fmt.Sprintf("\033[38;2;%d;%d;%dm\033[48;2;%d;%d;%dm",
		dlFR, dlFG, dlFB, dlBR, dlBG, dlBB)
}

func nightColor() string {
	ldFR, ldFG, ldFB := randomColor(220, 250)
	ldBR, ldBG, ldBB := randomColor(10, 50)

	return fmt.Sprintf("\033[38;2;%d;%d;%dm\033[48;2;%d;%d;%dm",
		ldFR, ldFG, ldFB, ldBR, ldBG, ldBB)
}

func randomColor(min, max int) (int, int, int) {
	return rand.IntN(max-min) + min, rand.IntN(max-min) + min, rand.IntN(max-min) + min
}
