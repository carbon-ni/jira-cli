package cmdutil

import (
	"os"
	"strings"

	"golang.org/x/term"
)

// IsDumbTerminal checks TERM/WT_SESSION environment variable and returns true
// if they indicate a dumb terminal.
//
// A dumb terminal has limited capability and may not handle special character
// sequences such as ANSI escape codes.
func IsDumbTerminal() bool {
	term := strings.ToLower(os.Getenv("TERM"))
	_, wtSession := os.LookupEnv("WT_SESSION")
	return !wtSession && (term == "" || term == "dumb")
}

// IsNotTTY returns true if stdout is not a TTY.
func IsNotTTY() bool {
	return !term.IsTerminal(int(os.Stdout.Fd()))
}
