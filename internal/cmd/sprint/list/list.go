package list

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/ankitpokhrel/jira-cli/internal/cmd/issue/list"
	"github.com/ankitpokhrel/jira-cli/internal/cmdutil"
	"github.com/ankitpokhrel/jira-cli/internal/query"
	"github.com/ankitpokhrel/jira-cli/internal/view"
	"github.com/ankitpokhrel/jira-cli/pkg/jira"
)

const (
	numSprints = 50 // This is the maximum result returned by Jira API at once.
	helpText   = `
Sprints are displayed in a plain table view by default.`

	examples = `$ jira sprint list

# List issues in a sprint
$ jira sprint list <SPRINT_ID>

# List sprints or sprint issues without headers
$ jira sprint list --no-headers
$ jira sprint list <SPRINT_ID> --no-headers

# Display some columns of sprints or sprint issues
$ jira sprint list --columns name,start,end
$ jira sprint list <SPRINT_ID> --columns type,key,summary

# Display sprint issues and show all fields
$ jira sprint list <SPRINT_ID> --no-truncate`
)

// NewCmdList is a sprint list command.
func NewCmdList() *cobra.Command {
	return &cobra.Command{
		Use:     "list [SPRINT_ID]",
		Short:   fmt.Sprintf("Sprint lists top %d sprints in a board", numSprints),
		Long:    fmt.Sprintf("Sprint lists top %d sprints in a board\n", numSprints) + helpText,
		Example: examples,
		Args:    cobra.MaximumNArgs(1),
		Aliases: []string{"lists", "ls"},
		Annotations: map[string]string{
			"help:args": "[SPRINT_ID]\tID of the sprint",
		},
		Run: sprintList,
	}
}

// SetFlags sets flags supported by a sprint list command.
func SetFlags(cmd *cobra.Command) {
	list.SetFlags(cmd)
	setFlags(cmd)
	hideFlags(cmd)
}

func sprintList(cmd *cobra.Command, args []string) {
	server := viper.GetString("server")
	project := viper.GetString("project.key")
	boardID := viper.GetInt("board.id")

	debug, err := cmd.Flags().GetBool("debug")
	cmdutil.ExitIfError(err)

	client := cmdutil.NewJiraClient(debug)
	format := cmdutil.OutputFormat(cmd)

	sprintQuery, err := query.NewSprint(cmd.Flags())
	cmdutil.ExitIfError(err)

	if len(args) == 0 {
		sprintExplorerView(sprintQuery, cmd.Flags(), boardID, project, server, client, format)
	} else {
		sprintID, err := strconv.Atoi(args[0])
		cmdutil.ExitIfError(err)

		singleSprintView(sprintQuery, cmd.Flags(), sprintID, project, server, client, nil, format)
	}
}

func singleSprintView(sprintQuery *query.Sprint, flags query.FlagParser, sprintID int, project, server string, client *jira.Client, sprint *jira.Sprint, format string) {
	issues, jql, err := func() ([]*jira.Issue, string, error) {
		s := cmdutil.Info("Fetching sprint issues...")
		defer s.Stop()

		q, err := query.NewIssue(project, flags)
		if err != nil {
			return nil, "", err
		}
		if sprintQuery.Params().ShowAllIssues {
			q.Params().JQL = "project IS NOT EMPTY"
		}
		jql := q.Get()
		resp, err := client.SprintIssues(sprintID, jql, q.Params().From, q.Params().Limit)
		if err != nil {
			return nil, "", err
		}
		return resp.Issues, jql, nil
	}()
	cmdutil.ExitIfError(err)

	if cmdutil.IsStructured(format) {
		list.RenderStructured(issues, jql, format)
		return
	}

	if len(issues) == 0 {
		fmt.Println()
		cmdutil.Failed("No result found for given query in project %q", project)
		return
	}

	plain, err := flags.GetBool("plain")
	cmdutil.ExitIfError(err)

	csv, err := flags.GetBool("csv")
	cmdutil.ExitIfError(err)

	delimiter, err := flags.GetString("delimiter")
	cmdutil.ExitIfError(err)

	noHeaders, err := flags.GetBool("no-headers")
	cmdutil.ExitIfError(err)

	noTruncate, err := flags.GetBool("no-truncate")
	cmdutil.ExitIfError(err)

	columns, err := flags.GetString("columns")
	cmdutil.ExitIfError(err)

	var ft string
	if sprint != nil {
		if sprint.Status == jira.SprintStateFuture {
			ft = fmt.Sprintf(
				"Showing %d results for project %q in sprint #%d ➤ %s (Future Sprint)",
				len(issues), project, sprint.ID, sprint.Name,
			)
		} else {
			ft = fmt.Sprintf(
				"Showing %d results for project %q in sprint #%d ➤ %s (%s - %s)",
				len(issues), project, sprint.ID, sprint.Name,
				view.FormatDateTimeHuman(sprint.StartDate, time.RFC3339),
				view.FormatDateTimeHuman(sprint.EndDate, time.RFC3339),
			)
		}
	} else {
		ft = fmt.Sprintf(
			"Showing %d results for project %q in sprint #%d",
			len(issues), project, sprintID,
		)
	}

	v := view.IssueList{
		Project:    project,
		Server:     server,
		Data:       issues,
		FooterText: ft,
		Display: view.DisplayFormat{
			Plain:      plain,
			Delimiter:  delimiter,
			CSV:        csv,
			NoHeaders:  noHeaders,
			NoTruncate: noTruncate,
			Columns: func() []string {
				if columns != "" {
					return strings.Split(columns, ",")
				}
				return []string{}
			}(),
			Timezone: viper.GetString("timezone"),
		},
	}

	cmdutil.ExitIfError(v.Render())
}

func sprintExplorerView(sprintQuery *query.Sprint, flags query.FlagParser, boardID int, project, server string, client *jira.Client, format string) {
	sprints := func() []*jira.Sprint {
		s := cmdutil.Info("Fetching sprints...")
		defer s.Stop()

		return client.SprintsInBoards([]int{boardID}, sprintQuery.Get(), numSprints)
	}()
	if len(sprints) == 0 {
		if cmdutil.IsStructured(format) {
			renderStructured(nil, project, viper.GetString("board.name"), format)
			return
		}
		fmt.Println()
		cmdutil.Failed("No result found for given query in project %q", project)
		return
	}

	if sprintQuery.Params().Current || sprintQuery.Params().Prev || sprintQuery.Params().Next {
		sprint := sprints[0]
		if sprintQuery.Params().Next {
			sprint = sprints[len(sprints)-1]
		}
		singleSprintView(sprintQuery, flags, sprint.ID, project, server, client, sprint, format)
		return
	}

	if cmdutil.IsStructured(format) {
		renderStructured(sprints, project, viper.GetString("board.name"), format)
		return
	}

	plain, err := flags.GetBool("plain")
	cmdutil.ExitIfError(err)

	noHeaders, err := flags.GetBool("no-headers")
	cmdutil.ExitIfError(err)

	columns, err := flags.GetString("columns")
	cmdutil.ExitIfError(err)

	v := view.SprintList{
		Project: project,
		Board:   viper.GetString("board.name"),
		Server:  server,
		Data:    sprints,
		Display: view.DisplayFormat{
			Plain:     plain,
			NoHeaders: noHeaders,
			Columns: func() []string {
				if columns != "" {
					return strings.Split(columns, ",")
				}
				return []string{}
			}(),
			Timezone: viper.GetString("timezone"),
		},
	}

	cmdutil.ExitIfError(v.Render())
}

func setFlags(cmd *cobra.Command) {
	cmd.Flags().String("state", "", "Filter sprint by its state (comma separated).\n"+
		"Valid values are future, active and closed.\n"+
		`Defaults to "active,closed"`)
	cmd.Flags().Bool("show-all-issues", false, "Show sprint issues from all projects")
	cmd.Flags().String("columns", "", "Comma separated list of columns to display in the plain mode.\n"+
		fmt.Sprintf("Accepts (for sprint list): %s", strings.Join(view.ValidSprintColumns(), ", "))+
		fmt.Sprintf("\nAccepts (for sprint issues): %s", strings.Join(view.ValidIssueColumns(), ", ")))
	cmd.Flags().Bool("current", false, "List issues in current active sprint")
	cmd.Flags().Bool("prev", false, "List issues in previous sprint")
	cmd.Flags().Bool("next", false, "List issues in next planned sprint")
}

func hideFlags(cmd *cobra.Command) {
	cmdutil.ExitIfError(cmd.Flags().MarkHidden("history"))
	cmdutil.ExitIfError(cmd.Flags().MarkHidden("watching"))
	cmdutil.ExitIfError(cmd.Flags().MarkHidden("type"))
	cmdutil.ExitIfError(cmd.Flags().MarkHidden("resolution"))
	cmdutil.ExitIfError(cmd.Flags().MarkHidden("status"))
	cmdutil.ExitIfError(cmd.Flags().MarkHidden("priority"))
	cmdutil.ExitIfError(cmd.Flags().MarkHidden("reporter"))
	cmdutil.ExitIfError(cmd.Flags().MarkHidden("assignee"))
	cmdutil.ExitIfError(cmd.Flags().MarkHidden("created"))
	cmdutil.ExitIfError(cmd.Flags().MarkHidden("updated"))
	cmdutil.ExitIfError(cmd.Flags().MarkHidden("created-after"))
	cmdutil.ExitIfError(cmd.Flags().MarkHidden("updated-after"))
	cmdutil.ExitIfError(cmd.Flags().MarkHidden("created-before"))
	cmdutil.ExitIfError(cmd.Flags().MarkHidden("updated-before"))
	cmdutil.ExitIfError(cmd.Flags().MarkHidden("label"))
	cmdutil.ExitIfError(cmd.Flags().MarkHidden("reverse"))
}
