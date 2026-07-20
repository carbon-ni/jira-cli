package me

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/ankitpokhrel/jira-cli/api"
	"github.com/ankitpokhrel/jira-cli/internal/cmdutil"
	"github.com/ankitpokhrel/jira-cli/pkg/jira"
)

func TestMePrintsConfiguredLogin(t *testing.T) {
	viper.Reset()
	viper.Set("login", "configured-user")

	cmd := humanCmd(t)
	got := captureStdout(t, func() { me(cmd, nil) })

	if got != "configured-user\n" {
		t.Fatalf("expected configured login, got %q", got)
	}
}

func TestMeFetchesCurrentUserWhenLoginIsEmpty(t *testing.T) {
	viper.Reset()
	api.ResetClient()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/rest/api/2/myself" {
			t.Fatalf("expected /rest/api/2/myself, got %s", r.URL.Path)
		}
		if r.Header.Get("Cookie") != "cloud.session=abc" {
			t.Fatalf("expected cookie auth header, got %q", r.Header.Get("Cookie"))
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"name":"cookie-user","displayName":"Cookie User","emailAddress":"cookie@example.com"}`))
	}))
	defer server.Close()

	viper.Set("server", server.URL)
	viper.Set("auth_type", "cookie")
	viper.Set("cookies", "cloud.session=abc")

	cmd := humanCmd(t)
	got := captureStdout(t, func() { me(cmd, nil) })

	if got != "cookie-user\n" {
		t.Fatalf("expected fetched login, got %q", got)
	}
}

func TestMeStructuredTOONEmitsFullProfile(t *testing.T) {
	viper.Reset()
	api.ResetClient()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/rest/api/2/myself" {
			t.Fatalf("expected /rest/api/2/myself, got %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"name":"ada","accountId":"5e4:abc","displayName":"Ada Lovelace","emailAddress":"ada@example.com","timeZone":"UTC"}`))
	}))
	defer server.Close()

	viper.Set("server", server.URL)
	viper.Set("auth_type", "cookie")
	viper.Set("cookies", "cloud.session=abc")

	cmd := &cobra.Command{}
	cmd.Flags().String("format", cmdutil.FormatAuto, "")
	if err := cmd.Flags().Set("format", cmdutil.FormatTOON); err != nil {
		t.Fatalf("set format: %v", err)
	}

	got := captureStdout(t, func() { me(cmd, nil) })

	// TOON object preserving jira.Me json-tag order; accountId is quoted because
	// it contains a colon (TOON §7.2). No trailing newline.
	want := "name: ada\naccountId: \"5e4:abc\"\ndisplayName: Ada Lovelace\nemailAddress: ada@example.com\ntimeZone: UTC"
	if got != want {
		t.Fatalf("TOON mismatch\nwant: %q\n got: %q", want, got)
	}
	if strings.HasSuffix(got, "\n") {
		t.Fatalf("TOON output must not end with a newline, got %q", got)
	}
}

func TestMeStructuredJSONEmitsValidJSON(t *testing.T) {
	viper.Reset()
	api.ResetClient()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"name":"ada","accountId":"id","displayName":"Ada","emailAddress":"ada@example.com","timeZone":"UTC"}`))
	}))
	defer server.Close()

	viper.Set("server", server.URL)
	viper.Set("auth_type", "cookie")
	viper.Set("cookies", "cloud.session=abc")

	cmd := &cobra.Command{}
	cmd.Flags().String("format", cmdutil.FormatAuto, "")
	if err := cmd.Flags().Set("format", cmdutil.FormatJSON); err != nil {
		t.Fatalf("set format: %v", err)
	}

	got := captureStdout(t, func() { me(cmd, nil) })

	if !strings.HasSuffix(got, "\n") {
		t.Fatalf("JSON output should end with a newline, got %q", got)
	}
	if got != strings.TrimRight(got, "\n")+"\n" {
		t.Fatalf("JSON output should have exactly one trailing newline, got %q", got)
	}
}

func TestCurrentUserIdentifierFallsBackWhenLoginIsEmpty(t *testing.T) {
	tests := []struct {
		name string
		user *jira.Me
		want string
	}{
		{name: "email", user: &jira.Me{Email: "cookie@example.com", Name: "Cookie User", AccountID: "account-id"}, want: "cookie@example.com"},
		{name: "display name", user: &jira.Me{Name: "Cookie User", AccountID: "account-id"}, want: "Cookie User"},
		{name: "account id", user: &jira.Me{AccountID: "account-id"}, want: "account-id"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := currentUserIdentifier(tt.user); got != tt.want {
				t.Fatalf("expected %q, got %q", tt.want, got)
			}
		})
	}
}

// humanCmd returns a command pinned to the legacy (auto) rendering, used by
// tests that exercise the non-structured path now that TOON is the default.
func humanCmd(t *testing.T) *cobra.Command {
	t.Helper()
	cmd := &cobra.Command{}
	cmd.Flags().String("format", cmdutil.FormatAuto, "")
	if err := cmd.Flags().Set("format", cmdutil.FormatAuto); err != nil {
		t.Fatalf("set format: %v", err)
	}
	return cmd
}

func captureStdout(t *testing.T, fn func()) string {
	t.Helper()

	oldStdout := os.Stdout
	reader, writer, err := os.Pipe()
	if err != nil {
		t.Fatalf("pipe stdout: %v", err)
	}

	os.Stdout = writer
	fn()
	_ = writer.Close()
	os.Stdout = oldStdout

	var buffer bytes.Buffer
	_, _ = io.Copy(&buffer, reader)
	return buffer.String()
}
