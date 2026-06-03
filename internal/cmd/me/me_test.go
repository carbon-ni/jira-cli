package me

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/spf13/viper"

	"github.com/ankitpokhrel/jira-cli/pkg/jira"
)

func TestMePrintsConfiguredLogin(t *testing.T) {
	viper.Reset()
	viper.Set("login", "configured-user")

	got := captureStdout(t, func() { me(nil, nil) })

	if got != "configured-user\n" {
		t.Fatalf("expected configured login, got %q", got)
	}
}

func TestMeFetchesCurrentUserWhenLoginIsEmpty(t *testing.T) {
	viper.Reset()
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

	got := captureStdout(t, func() { me(nil, nil) })

	if got != "cookie-user\n" {
		t.Fatalf("expected fetched login, got %q", got)
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
