package version

import "testing"

func TestDefaultVersionUsesNextMajorDevelopmentVersion(t *testing.T) {
	if Version != "v1.0.0-dev" {
		t.Fatalf("Version = %q, want %q", Version, "v1.0.0-dev")
	}
}
