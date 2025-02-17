package utils

import (
	"encoding/json"
	"fmt"
	"math/rand/v2"
	"os"
	"runtime/debug"
	"strings"
)

const (
	colorNone = "\033[0m"

	red     = "\033[0;31m"
	green   = "\033[38;5;76m"
	blue    = "\033[38;5;39m"
	magenta = "\x1b[35m"

	whiteBg   = "\033[40;5;135m"
	redBg     = "\033[41;5;135m"
	greenBg   = "\033[42;5;135m"
	blueBg    = "\033[44;5;135m"
	magentaBg = "\033[45;5;135m"
	purpleBg  = "\033[48;5;135m"
)

var colors = []string{magenta, blue, green}

const (
	PROJECTNAME = "/secretary/"
	TESTMODE    = false
)

func Log(msgs ...any) {
	randomColor := rand.IntN(len(colors))

	for _, msg := range msgs {
		if err, ok := msg.(error); ok {
			fmt.Fprintf(os.Stderr, "%s", red+err.Error()+colorNone+"\n")

			lines := strings.Split(string(debug.Stack()), "\n")
			for _, line := range lines {
				if strings.Contains(line, PROJECTNAME) && strings.Contains(line, ".go") {
					fmt.Fprintf(os.Stderr, "%s", red+line+colorNone+"\n")
				}
			}

			continue
		} else {
			b, err := json.MarshalIndent(msg, "", "  ")
			if err != nil {
				fmt.Print(msg)
			}
			fmt.Fprintf(os.Stdout, "%s", colors[randomColor%len(colors)]+string(b)+colorNone)
		}
	}
}
