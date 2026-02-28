package cmd

import "fmt"

const (
	ColorReset  = "\033[0m"
	ColorRed    = "\033[31m"
	ColorGreen  = "\033[32m"
	ColorYellow = "\033[33m"
	ColorCyan   = "\033[36m"
)

func Info(format string, a ...any) {
	fmt.Printf(ColorCyan+"→ "+format+ColorReset+"\n", a...)
}

func Success(format string, a ...any) {
	fmt.Printf(ColorGreen+"✓ "+format+ColorReset+"\n", a...)
}

func Warning(format string, a ...any) {
	fmt.Printf(ColorYellow+"! "+format+ColorReset+"\n", a...)
}

func ErrorLog(format string, a ...any) {
	fmt.Printf(ColorRed+"✗ "+format+ColorReset+"\n", a...)
}
