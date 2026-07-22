package cmdutil

import (
	"os"
	"testing"

	"github.com/mitchellh/go-homedir"
	"github.com/stretchr/testify/assert"
)

func TestGetConfigHome(t *testing.T) {
	t.Parallel()

	userHome, err := homedir.Dir()
	assert.NoError(t, err)

	os.Clearenv()

	configHome, err := GetConfigHome()
	assert.NoError(t, err)
	assert.Equal(t, userHome+"/.config", configHome)

	assert.NoError(t, os.Setenv("XDG_CONFIG_HOME", "./test"))

	configHome, err = GetConfigHome()
	assert.NoError(t, err)
	assert.Equal(t, "./test", configHome)
}
