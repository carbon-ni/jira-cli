package root

import (
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/ankitpokhrel/jira-cli/internal/cmd/board"
	"github.com/ankitpokhrel/jira-cli/internal/cmd/completion"
	"github.com/ankitpokhrel/jira-cli/internal/cmd/epic"
	initCmd "github.com/ankitpokhrel/jira-cli/internal/cmd/init"
	"github.com/ankitpokhrel/jira-cli/internal/cmd/issue"
	"github.com/ankitpokhrel/jira-cli/internal/cmd/man"
	"github.com/ankitpokhrel/jira-cli/internal/cmd/me"
	"github.com/ankitpokhrel/jira-cli/internal/cmd/open"
	"github.com/ankitpokhrel/jira-cli/internal/cmd/project"
	"github.com/ankitpokhrel/jira-cli/internal/cmd/release"
	"github.com/ankitpokhrel/jira-cli/internal/cmd/serverinfo"
	"github.com/ankitpokhrel/jira-cli/internal/cmd/sprint"
	"github.com/ankitpokhrel/jira-cli/internal/cmd/version"
	"github.com/ankitpokhrel/jira-cli/internal/cmdutil"
	jiraConfig "github.com/ankitpokhrel/jira-cli/internal/config"
	"github.com/ankitpokhrel/jira-cli/pkg/jira"
	"github.com/ankitpokhrel/jira-cli/pkg/netrc"

	"github.com/zalando/go-keyring"
)

const (
	jiraCLIHelpLink          = "https://github.com/ankitpokhrel/jira-cli#getting-started"
	jiraAPITokenLink         = "https://id.atlassian.com/manage-profile/security/api-tokens"
	atlassianCookieConfigDir = "atlassian"
)

var (
	config string
	debug  bool
)

func init() {
	cobra.OnInitialize(func() {
		if config != "" {
			// 1. Command line flag has the highest priority
			viper.SetConfigFile(config)
		} else if configFile := os.Getenv("JIRA_CONFIG_FILE"); configFile != "" {
			// 2. Environment variable has second priority
			viper.SetConfigFile(configFile)
		} else {
			// 3. Default location has the lowest priority
			home, err := cmdutil.GetConfigHome()
			if err != nil {
				cmdutil.Failed("Error: %s", err)
				return
			}

			viper.AddConfigPath(fmt.Sprintf("%s/%s", home, jiraConfig.Dir))
			viper.SetConfigName(jiraConfig.FileName)
			viper.SetConfigType(jiraConfig.FileType)
		}

		viper.AutomaticEnv()
		viper.SetEnvPrefix("jira")

		if err := viper.ReadInConfig(); err == nil && debug {
			fmt.Printf("Using config file: %s\n", viper.ConfigFileUsed())
		}
	})
}

// NewCmdRoot is a root command.
func NewCmdRoot() *cobra.Command {
	cmd := cobra.Command{
		Use:   "jira <command> <subcommand>",
		Short: "Interactive Jira CLI",
		Long:  "Interactive Jira command line.",
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
		PersistentPreRun: func(cmd *cobra.Command, _ []string) {
			subCmd := cmd.Name()
			if !cmdRequireToken(subCmd) {
				return
			}

			switch viper.GetString("auth_type") {
			case string(jira.AuthTypeMTLS):
				// mTLS doesn't need Jira API Token.
			case string(jira.AuthTypeCookie):
				checkForJiraCookies()
			default:
				checkForJiraToken(viper.GetString("server"), viper.GetString("login"))
			}

			configFile := viper.ConfigFileUsed()
			if !jiraConfig.Exists(configFile) {
				cmdutil.Failed("Missing configuration file.\nRun 'jira init' to configure the tool.")
			}
		},
	}

	configHome, err := cmdutil.GetConfigHome()
	if err != nil {
		cmdutil.Failed("Error: %s", err)
	}

	cmd.PersistentFlags().StringVarP(
		&config, "config", "c", "",
		fmt.Sprintf("Config file (default is %s/%s/%s.yml, can be overridden with JIRA_CONFIG_FILE env var)", configHome, jiraConfig.Dir, jiraConfig.FileName),
	)
	cmd.PersistentFlags().StringP(
		"project", "p", "",
		fmt.Sprintf(
			"Jira project to look into (defaults to %s/%s/%s.yml)",
			configHome, jiraConfig.Dir, jiraConfig.FileName,
		),
	)
	cmd.PersistentFlags().BoolVar(&debug, "debug", false, "Turn on debug output")

	cmd.SetHelpFunc(helpFunc)

	_ = viper.BindPFlag("config", cmd.PersistentFlags().Lookup("config"))
	_ = viper.BindPFlag("project.key", cmd.PersistentFlags().Lookup("project"))
	_ = viper.BindPFlag("debug", cmd.PersistentFlags().Lookup("debug"))

	addChildCommands(&cmd)

	return &cmd
}

func addChildCommands(cmd *cobra.Command) {
	cmd.AddCommand(
		initCmd.NewCmdInit(),
		issue.NewCmdIssue(),
		epic.NewCmdEpic(),
		sprint.NewCmdSprint(),
		board.NewCmdBoard(),
		project.NewCmdProject(),
		open.NewCmdOpen(),
		me.NewCmdMe(),
		serverinfo.NewCmdServerInfo(),
		completion.NewCmdCompletion(),
		version.NewCmdVersion(),
		release.NewCmdRelease(),
		man.NewCmdMan(),
	)
}

func cmdRequireToken(cmd string) bool {
	allowList := []string{
		"init",
		"help",
		"jira",
		"version",
		"completion",
		"__complete", "__completeNoDesc", // Subcommand name during autocompletion call.
		"man",
	}
	return !slices.Contains(allowList, cmd)
}

func checkForJiraCookies() {
	if viper.GetString("cookies") != "" {
		return
	}

	configDir, err := cmdutil.GetConfigHome()
	if err == nil {
		for _, path := range []string{
			filepath.Join(configDir, jiraConfig.Dir, "cookies.txt"),
			filepath.Join(configDir, atlassianCookieConfigDir, "cookies.txt"),
		} {
			cookies, err := os.ReadFile(path)
			if err == nil && strings.TrimSpace(string(cookies)) != "" {
				return
			}
		}
	}

	cmdutil.Warn(`The tool needs Jira browser cookies to function with cookie auth.

You can either:
  - Export cookies to your shell as a JIRA_COOKIES env variable
  - Or, write the full Cookie header to ~/.config/.jira/cookies.txt
  - Or, share the Atlassian cookie file at ~/.config/atlassian/cookies.txt

You can extract cookies from browser devtools or a HAR file.
`)
	os.Exit(1)
}

func checkForJiraToken(server string, login string) {
	if os.Getenv("JIRA_API_TOKEN") != "" {
		return
	}

	netrcConfig, _ := netrc.Read(server, login)
	if netrcConfig != nil {
		return
	}

	secret, _ := keyring.Get("jira-cli", login)
	if secret != "" {
		return
	}

	msg := fmt.Sprintf(`The tool needs a Jira API token to function.

For cloud server: you can generate the token using this link: %s
For local server: you can use the password you use to log in to Jira for basic auth or get a token from your Jira profile for PAT.

After generating the token, you can either:
  - Export API token to your shell as a JIRA_API_TOKEN env variable
  - Or, you can use a .netrc file to define required machine details

Once you are done with the above steps, run 'jira init' to generate the config if you haven't already.

For more details, see: %s
`, jiraAPITokenLink, jiraCLIHelpLink)

	cmdutil.Warn(msg)
	os.Exit(1)
}
