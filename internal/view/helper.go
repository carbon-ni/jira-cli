package view

import (
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/charmbracelet/glamour"
	"github.com/fatih/color"
	"github.com/mgutz/ansi"
)

const (
	wordWrap = 120
	tabWidth = 8
)

// ValidIssueColumns returns valid columns for issue list.
func ValidIssueColumns() []string {
	return []string{
		fieldType,
		fieldKey,
		fieldSummary,
		fieldStatus,
		fieldAssignee,
		fieldReporter,
		fieldPriority,
		fieldResolution,
		fieldCreated,
		fieldUpdated,
		fieldLabels,
	}
}

// ValidSprintColumns returns valid columns for sprint list.
func ValidSprintColumns() []string {
	return []string{
		fieldID,
		fieldName,
		fieldStartDate,
		fieldEndDate,
		fieldCompleteDate,
		fieldState,
	}
}

// MDRenderer constructs markdown renderer.
func MDRenderer() (*glamour.TermRenderer, error) {
	return glamour.NewTermRenderer(
		glamour.WithEnvironmentConfig(),
		glamour.WithWordWrap(wordWrap),
	)
}

func formatDateTime(dt, format, tz string) string {
	t, err := time.Parse(format, dt)
	if err != nil {
		return dt
	}
	if tz == "" {
		return t.Format("2006-01-02 15:04:05")
	}
	loc, err := time.LoadLocation(tz)
	if err != nil {
		return dt
	}
	return t.In(loc).Format("2006-01-02 15:04:05")
}

func prepareTitle(text string) string {
	return strings.TrimSpace(text)
}

func renderPlain(w io.Writer, data [][]string, delimiter string) error {
	for _, items := range data {
		n := len(items)
		for j, v := range items {
			_, _ = fmt.Fprintf(w, "%s", v)
			if j != n-1 {
				_, _ = fmt.Fprintf(w, "%s", delimiter)
			}
		}
		_, _ = fmt.Fprintln(w)
	}

	if _, ok := w.(*tabwriter.Writer); ok {
		return w.(*tabwriter.Writer).Flush()
	}
	return nil
}

func renderCSV(w io.Writer, data [][]string) error {
	csvwrt := csv.NewWriter(w)

	for _, items := range data {
		if err := csvwrt.Write(items); err != nil {
			return err
		}
	}

	csvwrt.Flush()
	if err := csvwrt.Error(); err != nil {
		return err
	}
	return nil
}

func coloredOut(msg string, clr color.Attribute, attrs ...color.Attribute) string {
	c := color.New(clr).Add(attrs...)
	return c.Sprint(msg)
}

func xterm256() bool {
	term := os.Getenv("TERM")
	return strings.Contains(term, "-256color")
}

func gray(msg string) string {
	if xterm256() {
		return gray256(msg)
	}
	return ansi.ColorFunc("black+h")(msg)
}

func gray256(msg string) string {
	return fmt.Sprintf("\x1b[38;5;242m%s\x1b[m", msg)
}

func shortenAndPad(msg string, limit int) string {
	if limit > 1 && len(msg) > limit {
		return msg[0:limit-1] + "…"
	}
	return pad(msg, limit)
}

func pad(msg string, limit int) string {
	var out strings.Builder
	out.WriteString(msg)
	for i := len(msg); i < limit; i++ {
		out.WriteRune(' ')
	}
	return out.String()
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
