package utils

import (
	"encoding/json"
	"fmt"
	"math/rand/v2"
	"os"
	"reflect"
	"regexp"
	"runtime/debug"
	"strings"
	"testing"

	"golang.org/x/term"
)

const (
	COLORRESET = "\033[0m"

	RED = "\033[0;31m"
)

var colorIndex = 0

const (
	MODE   = LIGHT
	NIGHT  = 0
	LIGHT  = 1
	SWITCH = -1
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

	log, t := LogMessage(msgs...)

	if t != nil {
		t.Fatal(log)
	} else {
		fmt.Println(log)
	}
}

func LogMessage(msgs ...any) (string, *testing.T) {
	var t *testing.T
	if len(msgs) > 0 {
		if tt, ok := msgs[0].(*testing.T); ok {
			t = tt
			msgs = msgs[1:] // Remove the first argument if it's *testing.T
		}
	}

	width, _, err := term.GetSize(int(os.Stdout.Fd()))
	if err != nil {
		width = 80 // Default width if terminal size can't be determined
	}

	colorIndex++
	color := Ternary(
		MODE == NIGHT || (MODE == SWITCH && colorIndex%2 == 0),
		NightColor().TermColor,
		LightColor().TermColor)

	log := color

	extracTrace := func(lines []string, i int) (name string, loc string) {
		line := lines[i]

		if strings.Contains(line, PROJECTNAME) && strings.Contains(line, ".go") {
			parts := strings.Split(strings.Trim(line, "\t"), " ")
			if len(parts) > 0 {
				projectPart := strings.Split(parts[0], PROJECTNAME)
				if len(projectPart) > 1 {
					loc = fmt.Sprint(projectPart[1], " ")
				}
			}

			if i > 0 { // Check index range
				prevParts := strings.Split(strings.Trim(lines[i-1], "\t"), " ")
				if len(prevParts) > 0 {
					funcPart := strings.Split(prevParts[0], PROJECTFUNCNAME)
					if len(funcPart) > 1 {
						funcName := strings.Split(funcPart[1], "(0x")
						if len(funcName) > 0 {
							name = fmt.Sprint(funcName[0], " ")
						}
					}
				}
			}
		}
		return name, loc
	}

	extractError := func(err error) {
		log += RED + err.Error() + COLORRESET + " "

		lines := strings.Split(string(debug.Stack()), "\n")

		for i := range lines {
			name, loc := extracTrace(lines, i)
			if loc != "" {
				log += name + loc + " "
			}
		}
		log += color + " \n"
	}

	{
		lines := strings.Split(string(debug.Stack()), "\n")
		lc := 0
		nameLoc := ""
		for i := range lines {
			if strings.Contains(lines[i], PROJECTNAME) && strings.Contains(lines[i], ".go") {
				if lc > 0 && lc < 6 {
					name, loc := extracTrace(lines, i)
					if lc == 1 {
						nameLoc += name
					}
					nameLoc += loc
				}
				lc++
			}
		}

		p := nameLoc
		if len(msgs) > 2 {
			p = padLine("> "+nameLoc, width-1, "-", false)
		}
		log += p
	}

	for i := 0; i < len(msgs); i++ {
		switch msg := msgs[i].(type) {

		case int, int8, int16, int32, int64,
			float32, float64,
			uint, uint8, uint16, uint32, uint64:
			log += fmt.Sprint(msg)
		case []string, []int, []float32:
			log += fmt.Sprintf("%d %v", reflect.ValueOf(msg).Len(), msg)
		case []byte:
			log += fmt.Sprintf("%d %v", reflect.ValueOf(msg).Len(), msg)
		case error:
			extractError(msg)
		case []error:
			for _, e := range msg {
				extractError(e)
			}

		case string:
			if strings.ContainsAny(msg, "%") { // Check if it contains format specifiers
				count := countFormatSpecifiers(msg)
				if i+count >= len(msgs) {
					log += fmt.Sprint("Error: Not enough values for format string:", msg)
					break
				}
				log += fmt.Sprintf(msg+"--\n--", msgs[i+1:i+1+count]...)
				i += count // Move index forward by number of consumed values
			} else {
				if msg == "\n" || msg == "" {
					log += "=\n="
				} else {
					log += "\n" + msg + " "
				}
			}
		default:
			data, err := json.MarshalIndent(msg, "", "  ")
			if err != nil || string(data) == "{}" {
				log += fmt.Sprint(msg, " ")
			} else {
				log += fmt.Sprintf("%v\n", string(data))
			}

		}
	}

	log = strings.TrimSuffix(log, "\n")

	log = processParagraph(log, len(color), width) + COLORRESET

	return log, t
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

type Color struct {
	TermColor string
	Hex       string
}

func LightColor() Color {
	dlFR, dlFG, dlFB := 0, 0, 0
	dlBR, dlBG, dlBB := randomColor(225, 256)

	hex := fmt.Sprintf("#%02X%02X%02X", dlBR, dlBG, dlBB)

	return Color{
		fmt.Sprintf("\033[38;2;%d;%d;%dm\033[48;2;%d;%d;%dm",
			dlFR, dlFG, dlFB, dlBR, dlBG, dlBB), hex,
	}
}

func NightColor() Color {
	ldFR, ldFG, ldFB := 255, 255, 255
	ldBR, ldBG, ldBB := randomColor(10, 60)

	hex := fmt.Sprintf("#%02X%02X%02X", ldBR, ldBG, ldBB)

	return Color{
		fmt.Sprintf("\033[38;2;%d;%d;%dm\033[48;2;%d;%d;%dm",
			ldFR, ldFG, ldFB, ldBR, ldBG, ldBB), hex,
	}
}

func randomColor(min, max int) (int, int, int) {
	return rand.IntN(max-min) + min, rand.IntN(max-min) + min, rand.IntN(max-min) + min
}

// Regular expression to count format specifiers (excluding "%%" escape)
var formatSpecifierRegex = regexp.MustCompile(`%([-+0# ]*\d*(?:\.\d*)?[vTtbcdoqxXUeEfFgGsp])`)

// countFormatSpecifiers counts how many format specifiers exist in a string.
func countFormatSpecifiers(format string) int {
	return len(formatSpecifierRegex.FindAllString(format, -1))
}
