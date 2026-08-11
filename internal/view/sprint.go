package view

import (
	"fmt"
	"io"
	"os"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/ankitpokhrel/jira-cli/pkg/jira"
)

// SprintList is a list view for sprints.
type SprintList struct {
	Project string
	Board   string
	Server  string
	Data    []*jira.Sprint
	Display DisplayFormat
}

// Render renders the sprint list as a plain table.
func (sl *SprintList) Render() error {
	w := tabwriter.NewWriter(os.Stdout, 0, tabWidth, 1, '\t', 0)
	return sl.renderPlain(w)
}

// renderPlain renders the sprints in plain view.
func (sl *SprintList) renderPlain(w io.Writer) error {
	// Sprint view supports only \t as delimiter, not custom.
	return renderPlain(w, sl.tableData(), "\t")
}

func (sl *SprintList) validColumnsMap() map[string]struct{} {
	columns := ValidSprintColumns()
	out := make(map[string]struct{}, len(columns))

	for _, c := range columns {
		out[c] = struct{}{}
	}

	return out
}

func (sl *SprintList) tableHeader() []string {
	if len(sl.Display.Columns) == 0 {
		return ValidSprintColumns()
	}

	var headers []string

	columnsMap := sl.validColumnsMap()
	for _, c := range sl.Display.Columns {
		c = strings.ToUpper(c)
		if _, ok := columnsMap[c]; ok {
			headers = append(headers, strings.ToUpper(c))
		}
	}

	return headers
}

func (sl *SprintList) tableData() [][]string {
	var data [][]string

	headers := sl.tableHeader()
	if !sl.Display.NoHeaders {
		data = append(data, headers)
	}
	if len(headers) == 0 {
		headers = ValidSprintColumns()
	}
	for _, s := range sl.Data {
		data = append(data, sl.assignColumns(headers, s))
	}

	return data
}

func (sl *SprintList) assignColumns(columns []string, sprint *jira.Sprint) []string {
	var bucket []string

	for _, column := range columns {
		switch column {
		case fieldID:
			bucket = append(bucket, fmt.Sprintf("%d", sprint.ID))
		case fieldName:
			bucket = append(bucket, sprint.Name)
		case fieldStartDate:
			bucket = append(bucket, formatDateTime(sprint.StartDate, time.RFC3339, sl.Display.Timezone))
		case fieldEndDate:
			bucket = append(bucket, formatDateTime(sprint.EndDate, time.RFC3339, sl.Display.Timezone))
		case fieldCompleteDate:
			bucket = append(bucket, formatDateTime(sprint.CompleteDate, time.RFC3339, sl.Display.Timezone))
		case fieldState:
			bucket = append(bucket, sprint.Status)
		}
	}

	return bucket
}
