package cmdutil

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/mitchellh/go-homedir"
	"github.com/spf13/viper"
	"github.com/zalando/go-keyring"

	"github.com/ankitpokhrel/jira-cli/api"
	"github.com/ankitpokhrel/jira-cli/pkg/jira"
	"github.com/ankitpokhrel/jira-cli/pkg/netrc"
)

const (
	cookieConfigDir          = ".jira"
	atlassianCookieConfigDir = "atlassian"
	cookieFileName           = "cookies.txt"
)

// NewJiraClient resolves configuration from environment and creates a Jira client.
func NewJiraClient(debug bool) *jira.Client {
	return NewJiraClientWith(jira.Config{Debug: debug})
}

// NewJiraClientWith resolves missing config fields from environment and creates
// a Jira client. Supplied fields take precedence over environment resolution.
func NewJiraClientWith(config jira.Config) *jira.Client {
	resolveConfig(&config)
	return api.Client(config)
}

func resolveConfig(config *jira.Config) {
	if config.Server == "" {
		config.Server = viper.GetString("server")
	}
	if config.Login == "" {
		config.Login = viper.GetString("login")
	}
	if config.APIToken == "" {
		config.APIToken = viper.GetString("api_token")
	}
	if config.APIToken == "" {
		if netrcConfig, _ := netrc.Read(config.Server, config.Login); netrcConfig != nil {
			config.APIToken = netrcConfig.Password
		}
	}
	if config.APIToken == "" {
		secret, _ := keyring.Get("jira-cli", config.Login)
		config.APIToken = secret
	}
	if config.Cookies == "" {
		config.Cookies = viper.GetString("cookies")
	}
	if config.Cookies == "" {
		config.Cookies = readCookieFile()
	}
	if config.AuthType == nil {
		authType := jira.AuthType(viper.GetString("auth_type"))
		config.AuthType = &authType
	}
	if config.Insecure == nil {
		insecure := viper.GetBool("insecure")
		config.Insecure = &insecure
	}
	if config.Installation == "" {
		config.Installation = viper.GetString("installation")
	}
	if config.MTLSConfig.CaCert == "" {
		config.MTLSConfig.CaCert = viper.GetString("mtls.ca_cert")
	}
	if config.MTLSConfig.ClientCert == "" {
		config.MTLSConfig.ClientCert = viper.GetString("mtls.client_cert")
	}
	if config.MTLSConfig.ClientKey == "" {
		config.MTLSConfig.ClientKey = viper.GetString("mtls.client_key")
	}
}

func readCookieFile() string {
	home, err := getConfigDir()
	if err != nil {
		return ""
	}

	for _, path := range []string{
		filepath.Join(home, cookieConfigDir, cookieFileName),
		filepath.Join(home, atlassianCookieConfigDir, cookieFileName),
	} {
		cookies, err := os.ReadFile(path)
		if err == nil {
			return strings.TrimSpace(string(cookies))
		}
	}

	return ""
}

func getConfigDir() (string, error) {
	if home := os.Getenv("XDG_CONFIG_HOME"); home != "" {
		return home, nil
	}

	home, err := homedir.Dir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".config"), nil
}
