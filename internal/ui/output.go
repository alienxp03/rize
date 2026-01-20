package ui

import (
	"fmt"

	"github.com/fatih/color"
)

var (
	Green  = color.New(color.FgGreen).SprintFunc()
	Blue   = color.New(color.FgBlue).SprintFunc()
	Yellow = color.New(color.FgYellow).SprintFunc()
	Red    = color.New(color.FgRed).SprintFunc()
)

func Success(format string, a ...interface{}) {
	fmt.Printf("%s %s\n", Green("✓"), fmt.Sprintf(format, a...))
}

func Info(format string, a ...interface{}) {
	fmt.Printf("%s %s\n", Blue("→"), fmt.Sprintf(format, a...))
}

func Warning(format string, a ...interface{}) {
	fmt.Printf("%s %s\n", Yellow("!"), fmt.Sprintf(format, a...))
}

func Error(format string, a ...interface{}) {
	fmt.Fprintf(color.Output, "%s %s\n", Red("✗"), fmt.Sprintf(format, a...))
}
