package progress

import (
	"time"

	"github.com/fatih/color"
)

var (
	spinnerText = []rune("⠋⠙⠹⠸⠼⠴⠦⠧⠇⠏")
)

var (
	colorDone    = color.New(color.FgHiGreen)
	colorError   = color.New(color.FgHiRed)
	colorSpinner = color.New(color.FgHiCyan)
)

const (
	refreshRate = time.Millisecond * 50
	doneTail    = "Done"
	errorTail   = "Error"
)

// Bar controls how a bar is displayed, for both single bar or multi bar item.
type Bar interface {
	UpdateDisplay(newDisplay *DisplayProps)
}
