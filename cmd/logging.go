package cmd

import (
	"fmt"
)

const (
	colorRed    = "\033[31m"
	colorGreen  = "\033[32m"
	colorYellow = "\033[33m"
	colorBlue   = "\033[34m"
	colorReset  = "\033[0m" // Resets all formatting attributes
)

// Helper functions for colored logging
func logError(msg string) {
	fmt.Printf("%s[ERROR]%s %s\n", colorRed, colorReset, msg)
}

func logWarn(msg string) {
	fmt.Printf("%s[WARN]%s %s\n", colorYellow, colorReset, msg)
}

func logInfo(msg string) {
	fmt.Printf("%s[INFO]%s %s\n", colorGreen, colorReset, msg)
}

func logPrompt(msg string) {
	fmt.Printf("%s[INFO]%s %s", colorGreen, colorReset, msg)
}

func logDebug(msg string) {
	if debugMode {
		fmt.Printf("%s[DEBUG]%s %s\n", colorBlue, colorReset, msg)
	}
}
