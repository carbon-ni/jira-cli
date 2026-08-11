package view

import (
	"io"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/ankitpokhrel/jira-cli/pkg/jira"
)

// DisplayFormat is a issue display type.
type DisplayFormat struct {
	Plain      bool
	Delimiter  string
	CSV        bool
	NoHeaders  bool
	NoTruncate bool
	Columns    []string
	Comments   uint
	Timezone   string
}

// IssueList is a list view for issues.
type IssueList struct {
	Project    string
	Server     string
	Data       []*jira.Issue
	Display    DisplayFormat
	FooterText string
}

// Render renders the view.
func (l *IssueList) Render() error {
	// Prioritize CSV format when explicitly requested.
	if l.Display.CSV {
		return l.renderCSV(os.Stdout)
	}

	// Custom delimiter is used only in plain mode, otherwise \t is used.
	delimeter := "\t"
	if l.Display.Plain {
		delimeter = l.Display.Delimiter
	}
	w := tabwriter.NewWriter(os.Stdout, 0, tabWidth, 1, '\t', 0)
	return l.renderPlain(w, delimeter)
}

// renderPlain renders the issue in plain view.
func (l *IssueList) renderPlain(w io.Writer, delimeter string) error {
	return renderPlain(w, l.data(), delimeter)
}

// renderCSV renders issues in csv format.
func (l *IssueList) renderCSV(w io.Writer) error {
	return renderCSV(w, l.data())
}

func (*IssueList) validColumnsMap() map[string]struct{} {
	columns := ValidIssueColumns()
	out := make(map[string]struct{}, len(columns))

	for _, c := range columns {
		out[c] = struct{}{}
	}

	return out
}

func (l *IssueList) header() []string {
	if len(l.Display.Columns) == 0 {
		validColumns := ValidIssueColumns()
		if l.Display.NoTruncate || !l.Display.Plain {
			return validColumns
		}
		return validColumns[0:4]
	}

	return l.upperColumns(l.Display.Columns)
}

func (l *IssueList) upperColumns(cols []string) []string {
	var headers []string

	columnsMap := l.validColumnsMap()
	for _, c := range cols {
		c = strings.ToUpper(c)
		if _, ok := columnsMap[c]; ok {
			headers = append(headers, c)
		}
	}

	return headers
}

func (l *IssueList) data() [][]string {
	var data [][]string

	headers := l.header()
	if !l.Display.NoHeaders {
		data = append(data, headers)
	}
	for _, iss := range l.Data {
		data = append(data, l.assignColumns(headers, iss))
	}

	return data
}

func (l *IssueList) assignColumns(columns []string, issue *jira.Issue) []string {
	var bucket []string

	for _, column := range columns {
		switch column {
		case fieldType:
			bucket = append(bucket, issue.Fields.IssueType.Name)
		case fieldKey:
			bucket = append(bucket, issue.Key)
		case fieldSummary:
			bucket = append(bucket, prepareTitle(issue.Fields.Summary))
		case fieldStatus:
			bucket = append(bucket, issue.Fields.Status.Name)
		case fieldAssignee:
			bucket = append(bucket, issue.Fields.Assignee.Name)
		case fieldReporter:
			bucket = append(bucket, issue.Fields.Reporter.Name)
		case fieldPriority:
			bucket = append(bucket, issue.Fields.Priority.Name)
		case fieldResolution:
			bucket = append(bucket, issue.Fields.Resolution.Name)
		case fieldCreated:
			bucket = append(bucket, formatDateTime(issue.Fields.Created, jira.RFC3339, l.Display.Timezone))
		case fieldUpdated:
			bucket = append(bucket, formatDateTime(issue.Fields.Updated, jira.RFC3339, l.Display.Timezone))
		case fieldLabels:
			bucket = append(bucket, strings.Join(issue.Fields.Labels, ","))
		}
	}

	return bucket
}
