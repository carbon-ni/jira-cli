package view

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/ankitpokhrel/jira-cli/pkg/jira"
)

func TestSprintTableLayoutData(t *testing.T) {
	sprint := SprintList{
		Project: "TEST",
		Board:   "Test Board",
		Server:  "https://test.local",
		Data: []*jira.Sprint{
			{
				ID:           1,
				Name:         "Sprint 1",
				Status:       "closed",
				StartDate:    "2020-12-07T16:12:00.000Z",
				EndDate:      "2020-12-13T16:12:00.000Z",
				CompleteDate: "2020-12-13T16:12:00.000Z",
				BoardID:      1,
			},
			{
				ID:        2,
				Name:      "Sprint 2",
				Status:    "active",
				StartDate: "2020-12-13T16:12:00.000Z",
				EndDate:   "2020-12-19T16:12:00.000Z",
				BoardID:   1,
			},
		},
	}

	expected := [][]string{
		{"ID", "NAME", "START", "END", "COMPLETE", "STATE"},
		{"1", "Sprint 1", "2020-12-07 16:12:00", "2020-12-13 16:12:00", "2020-12-13 16:12:00", "closed"},
		{"2", "Sprint 2", "2020-12-13 16:12:00", "2020-12-19 16:12:00", "", "active"},
	}
	assert.Equal(t, expected, sprint.tableData())
}

func TestSprintRenderInPlainView(t *testing.T) {
	var b bytes.Buffer

	sprint := SprintList{
		Project: "TEST",
		Board:   "Test Board",
		Server:  "https://test.local",
		Data: []*jira.Sprint{
			{
				ID:           1,
				Name:         "Sprint 1",
				Status:       "closed",
				StartDate:    "2020-12-07T16:12:00.000Z",
				EndDate:      "2020-12-13T16:12:00.000Z",
				CompleteDate: "2020-12-13T16:12:00.000Z",
				BoardID:      1,
			},
			{
				ID:        2,
				Name:      "Sprint 2",
				Status:    "active",
				StartDate: "2020-12-13T16:12:00.000Z",
				EndDate:   "2020-12-19T16:12:00.000Z",
				BoardID:   1,
			},
		},
		Display: DisplayFormat{
			Plain:     true,
			NoHeaders: false,
		},
	}
	assert.NoError(t, sprint.renderPlain(&b))

	expected := `ID	NAME	START	END	COMPLETE	STATE
1	Sprint 1	2020-12-07 16:12:00	2020-12-13 16:12:00	2020-12-13 16:12:00	closed
2	Sprint 2	2020-12-13 16:12:00	2020-12-19 16:12:00		active
`
	assert.Equal(t, expected, b.String())
}

func TestSprintRenderInPlainViewWithoutHeaders(t *testing.T) {
	var b bytes.Buffer

	sprint := SprintList{
		Project: "TEST",
		Board:   "Test Board",
		Server:  "https://test.local",
		Data: []*jira.Sprint{
			{
				ID:           1,
				Name:         "Sprint 1",
				Status:       "closed",
				StartDate:    "2020-12-07T16:12:00.000Z",
				EndDate:      "2020-12-13T16:12:00.000Z",
				CompleteDate: "2020-12-13T16:12:00.000Z",
				BoardID:      1,
			},
			{
				ID:        2,
				Name:      "Sprint 2",
				Status:    "active",
				StartDate: "2020-12-13T16:12:00.000Z",
				EndDate:   "2020-12-19T16:12:00.000Z",
				BoardID:   1,
			},
		},
		Display: DisplayFormat{
			Plain:     true,
			NoHeaders: true,
		},
	}
	assert.NoError(t, sprint.renderPlain(&b))

	expected := `1	Sprint 1	2020-12-07 16:12:00	2020-12-13 16:12:00	2020-12-13 16:12:00	closed
2	Sprint 2	2020-12-13 16:12:00	2020-12-19 16:12:00		active
`
	assert.Equal(t, expected, b.String())
}

func TestSprintRenderInPlainViewWithFewColumns(t *testing.T) {
	var b bytes.Buffer

	sprint := SprintList{
		Project: "TEST",
		Board:   "Test Board",
		Server:  "https://test.local",
		Data: []*jira.Sprint{
			{
				ID:           1,
				Name:         "Sprint 1",
				Status:       "closed",
				StartDate:    "2020-12-07T16:12:00.000Z",
				EndDate:      "2020-12-13T16:12:00.000Z",
				CompleteDate: "2020-12-13T16:12:00.000Z",
				BoardID:      1,
			},
			{
				ID:        2,
				Name:      "Sprint 2",
				Status:    "active",
				StartDate: "2020-12-13T16:12:00.000Z",
				EndDate:   "2020-12-19T16:12:00.000Z",
				BoardID:   1,
			},
		},
		Display: DisplayFormat{
			Plain:     true,
			NoHeaders: false,
			Columns:   []string{"name", "start", "end"},
		},
	}
	assert.NoError(t, sprint.renderPlain(&b))

	expected := `NAME	START	END
Sprint 1	2020-12-07 16:12:00	2020-12-13 16:12:00
Sprint 2	2020-12-13 16:12:00	2020-12-19 16:12:00
`
	assert.Equal(t, expected, b.String())
}
