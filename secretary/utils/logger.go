package utils

import (
	"encoding/json"
	"fmt"
	"os"
	"runtime/debug"
	"strings"

	"golang.org/x/term"
)

const (
	colorNone = "\033[0m"

	red     = "\033[0;31m"
	green   = "\033[38;5;76m"
	blue    = "\033[38;5;39m"
	magenta = "\x1b[35m"

	black  = "\033[0;30m"
	yellow = "\033[0;33m"
	cyan   = "\033[0;36m"
	white  = "\033[0;37m"

	orange      = "\033[38;5;214m" // Bright Orange
	pink        = "\033[38;5;206m" // Hot Pink
	lightBlue   = "\033[38;5;45m"  // Light Blue
	teal        = "\033[38;5;44m"  // Teal
	purple      = "\033[38;5;129m" // Soft Purple
	lightGreen  = "\033[38;5;83m"  // Light Green
	gray        = "\033[38;5;245m" // Gray
	brightWhite = "\033[38;5;231m" // Bright White

	whiteBg   = "\033[40;5;135m"
	redBg     = "\033[41;5;135m"
	greenBg   = "\033[42;5;135m"
	blueBg    = "\033[44;5;135m"
	magentaBg = "\033[45;5;135m"
	purpleBg  = "\033[48;5;135m"
)

var (
	colorIndex = 0
	colors     = []string{
		// magenta, blue, green,
		blueBg, whiteBg, magentaBg, greenBg, purpleBg,
	}
)

const (
	PROJECTNAME     = "/secretary/"
	PROJECTFUNCNAME = "/secretary."
)

func Log(msgs ...any) {
	if len(msgs) == 0 || (len(msgs) == 1 && msgs[0] == "\n") {
		fmt.Println()
		return
	}

	colorIndex++
	randomColor := colors[colorIndex%len(colors)]
	hello := ""

	{
		lines := strings.Split(string(debug.Stack()), "\n")
		loc := 0

		nameLoc := ""

		for i, line := range lines {
			if strings.Contains(line, PROJECTNAME) && strings.Contains(line, ".go") {
				if loc == 1 {
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
				loc++
			}
		}

		hello += nameLoc
	}

	hello += randomColor

	for i, msg := range msgs {
		if _, ok := msg.(error); ok {
			// fmt.Fprintf(os.Stderr, "%s", "\n"+red+err.Error()+"\n")

			// lines := strings.Split(string(debug.Stack()), "\n")
			// for _, line := range lines {
			// 	if strings.Contains(line, PROJECTNAME) && strings.Contains(line, ".go") {
			// 		fmt.Fprintf(os.Stderr, "%s", line+"\n")
			// 	}
			// }
			// fmt.Print(colorNone)
		} else {
			if i%2 == 1 {
				hello += " "
			}
			if i%2 == 0 && len(msgs) > 2 {
				hello += "\n"
			}

			switch v := msg.(type) {
			case int, int8, int16, int32, int64,
				float32, float64,
				uint, uint8, uint16, uint32, uint64,
				string, []string,
				[]int, []float32:
				hello += fmt.Sprint(msg)
			default:
				data, err := json.MarshalIndent(v, "", "  ")
				if err != nil {
					hello += fmt.Sprint(msg)
				} else {
					hello += string(data)
				}
			}
		}
	}

	hello = processParagraph(hello, len(randomColor)) + colorNone

	fmt.Println(hello)
}

// Pads a single line with "_" to the nearest multiple of terminal width
func padLine(line string, width int) string {
	lineLen := len(line)

	// Calculate the next multiple of width
	targetWidth := ((lineLen / width) + 1) * width

	// Calculate how many "_" are needed to reach targetWidth
	padding := targetWidth - lineLen

	return line + strings.Repeat(" ", padding)
}

// Cleans and processes a paragraph
func processParagraph(paragraph string, colorlen int) string {
	width, _, err := term.GetSize(int(os.Stdout.Fd()))
	if err != nil {
		width = 80 // Default width if terminal size can't be determined
	}

	lines := strings.Split(paragraph, "\n") // Split paragraph by newlines
	for i, line := range lines {
		// fmt.Println(len(line), width)
		// fmt.Println(line)
		if i == 0 {
			lines[i] = padLine(line, width+colorlen) // Pad each line to a multiple of width
		} else {
			lines[i] = padLine(line, width) // Pad each line to a multiple of width
		}
		// fmt.Println(len(lines[i]), width)
		// fmt.Println(lines[i])
	}
	return strings.Join(lines, "\n") // Merge lines back
}
