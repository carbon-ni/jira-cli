package cmdutil

import (
	"os"
	"path/filepath"
	"testing"
)

func TestReadCookieFileUsesJiraCookieFileFirst(t *testing.T) {
	configHome := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", configHome)

	writeCookieFile(t, filepath.Join(configHome, cookieConfigDir, cookieFileName), "jira-cookie\n")
	writeCookieFile(t, filepath.Join(configHome, atlassianCookieConfigDir, cookieFileName), "atlassian-cookie\n")

	got := readCookieFile()

	if got != "jira-cookie" {
		t.Fatalf("expected Jira cookie file to be used first, got %q", got)
	}
}

func TestReadCookieFileFallsBackToAtlassianCookieFile(t *testing.T) {
	configHome := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", configHome)

	writeCookieFile(t, filepath.Join(configHome, atlassianCookieConfigDir, cookieFileName), "atlassian-cookie\n")

	got := readCookieFile()

	if got != "atlassian-cookie" {
		t.Fatalf("expected Atlassian cookie fallback, got %q", got)
	}
}

func TestReadCookieFileReturnsEmptyWhenNoCookieFileExists(t *testing.T) {
	configHome := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", configHome)

	got := readCookieFile()

	if got != "" {
		t.Fatalf("expected empty cookies, got %q", got)
	}
}

func writeCookieFile(t *testing.T, path string, content string) {
	t.Helper()

	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("create cookie dir: %v", err)
	}
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatalf("write cookie file: %v", err)
	}
}
