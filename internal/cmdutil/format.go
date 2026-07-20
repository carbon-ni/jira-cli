package cmdutil

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/toon-format/toon-go"
)

// Supported structured output formats. FormatAuto keeps the CLI's existing
// (human-oriented) rendering for a given command.
const (
	FormatAuto = "auto"
	FormatTOON = "toon"
	FormatJSON = "json"
)

// validFormats is the closed set accepted by the --format flag.
var validFormats = map[string]struct{}{
	FormatAuto: {},
	FormatTOON: {},
	FormatJSON: {},
}

// FormatFlagUsage is the help text for the global --format flag.
const FormatFlagUsage = "Output format for agent consumers: toon (default), json, or auto. " +
	"toon/json emit deterministic structured stdout; auto keeps a command's legacy rendering."

// DefaultFormat is the format used when --format is not specified. This CLI is
// agent-facing, so structured TOON is the default.
const DefaultFormat = FormatTOON

// OutputFormat resolves the --format value for a command, applying the
// agent-facing default (TOON) when the flag is absent or empty.
func OutputFormat(cmd *cobra.Command) string {
	if cmd == nil {
		return DefaultFormat
	}
	f, err := cmd.Flags().GetString("format")
	if err != nil || f == "" {
		return DefaultFormat
	}
	return f
}

// IsStructured reports whether the requested format is a machine-readable
// structured format (toon or json) rather than the human-oriented default.
func IsStructured(format string) bool {
	return format == FormatTOON || format == FormatJSON
}

// IsValidFormat reports whether format is one of the accepted --format values.
func IsValidFormat(format string) bool {
	_, ok := validFormats[format]
	return ok
}

// MarshalStructured encodes v for the requested structured format.
// TOON is the default structured format (AXI); JSON is offered for compatibility.
func MarshalStructured(v any, format string) ([]byte, error) {
	switch format {
	case FormatTOON:
		return toon.Marshal(v)
	case FormatJSON:
		return json.MarshalIndent(v, "", "  ")
	default:
		return nil, fmt.Errorf("unsupported structured format %q (want %q or %q)",
			format, FormatTOON, FormatJSON)
	}
}

// PrintStructured writes the encoded value to stdout without a trailing newline
// for TOON (spec §12) and with a single trailing newline for JSON.
func PrintStructured(v any, format string) error {
	b, err := MarshalStructured(v, format)
	if err != nil {
		return err
	}
	if _, err := os.Stdout.Write(b); err != nil {
		return err
	}
	if format == FormatJSON {
		if _, err := os.Stdout.Write([]byte("\n")); err != nil {
			return err
		}
	}
	return nil
}

// ErrorEnvelope is the structured error shape emitted to stdout when a command
// runs in a structured (toon/json) mode and fails. It keeps errors in the same
// format family as successful data so agents can parse a single contract.
type ErrorEnvelope struct {
	Error ErrorBody `json:"error" toon:"error"`
}

// ErrorBody describes a recoverable operational failure in domain terms.
type ErrorBody struct {
	// Code is a stable, domain-language error code (e.g. "auth-missing").
	Code string `json:"code" toon:"code"`
	// Message is a short, human-readable explanation in domain language.
	Message string `json:"message" toon:"message"`
	// Recovery is one actionable next step, or empty when none applies.
	Recovery string `json:"recovery,omitempty" toon:"recovery,omitempty"`
	// Hint is an optional copy-runnable command suggestion.
	Hint string `json:"hint,omitempty" toon:"hint,omitempty"`
}

// PrintStructuredError writes a structured error envelope to stdout in the
// requested format and returns the exit code the caller should use.
// Per AXI: operational failures use exit code 1, usage errors use 2.
func PrintStructuredError(env ErrorEnvelope, format string, usageError bool) int {
	if err := PrintStructured(env, format); err != nil {
		// Fall back to stderr so we never swallow the failure silently.
		fmt.Fprintf(os.Stderr, "%s\n", env.Error.Message)
	}
	if usageError {
		return 2
	}
	return 1
}
