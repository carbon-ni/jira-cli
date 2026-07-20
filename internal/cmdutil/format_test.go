package cmdutil

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/toon-format/toon-go"
)

func TestOutputFormatDefaultsToTOON(t *testing.T) {
	// Absent/empty flag yields the agent-facing default (TOON).
	cmd := &cobra.Command{}
	cmd.Flags().String("format", "", "")

	assert.Equal(t, FormatTOON, OutputFormat(cmd))
}

func TestOutputFormatRespectsFlag(t *testing.T) {
	cmd := &cobra.Command{}
	cmd.Flags().String("format", "", "")
	require.NoError(t, cmd.Flags().Set("format", FormatTOON))

	assert.Equal(t, FormatTOON, OutputFormat(cmd))
}

func TestOutputFormatNilCmdIsTOON(t *testing.T) {
	assert.Equal(t, FormatTOON, OutputFormat(nil))
}

func TestIsStructured(t *testing.T) {
	assert.True(t, IsStructured(FormatTOON))
	assert.True(t, IsStructured(FormatJSON))
	assert.False(t, IsStructured(FormatAuto))
	assert.False(t, IsStructured(""))
	assert.False(t, IsStructured("yaml"))
}

func TestIsValidFormat(t *testing.T) {
	for _, f := range []string{FormatAuto, FormatTOON, FormatJSON} {
		assert.True(t, IsValidFormat(f), "%q should be valid", f)
	}
	for _, f := range []string{"", "yaml", "toml", "TOON"} {
		assert.False(t, IsValidFormat(f), "%q should be invalid", f)
	}
}

func TestMarshalStructuredTOON(t *testing.T) {
	v := struct {
		ID   int    `json:"id" toon:"id"`
		Name string `json:"name" toon:"name"`
	}{ID: 1, Name: "Ada"}

	b, err := MarshalStructured(v, FormatTOON)
	require.NoError(t, err)
	assert.Equal(t, "id: 1\nname: Ada", string(b))
}

func TestMarshalStructuredJSON(t *testing.T) {
	v := struct {
		ID int `json:"id" toon:"id"`
	}{ID: 1}

	b, err := MarshalStructured(v, FormatJSON)
	require.NoError(t, err)

	var got map[string]int
	require.NoError(t, json.Unmarshal(b, &got))
	assert.Equal(t, 1, got["id"])
}

func TestMarshalStructuredTOONRoundTrips(t *testing.T) {
	type row struct {
		ID   int    `json:"id" toon:"id"`
		Name string `json:"name" toon:"name"`
	}
	type payload struct {
		Items []row `json:"items" toon:"items"`
		Total int   `json:"total" toon:"total"`
	}
	want := payload{
		Items: []row{{ID: 1, Name: "Ada"}, {ID: 2, Name: "Bob"}},
		Total: 2,
	}

	encoded, err := MarshalStructured(want, FormatTOON)
	require.NoError(t, err)

	var got payload
	require.NoError(t, toon.Unmarshal(encoded, &got))
	assert.Equal(t, want, got)
}

func TestMarshalStructuredRejectsUnknownFormat(t *testing.T) {
	_, err := MarshalStructured(struct{ X int }{X: 1}, "yaml")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "unsupported structured format")
}

func TestMarshalStructuredIsDeterministicAcrossFormats(t *testing.T) {
	v := struct {
		ID   int    `json:"id" toon:"id"`
		Name string `json:"name" toon:"name"`
	}{ID: 7, Name: "Bob"}

	first, err := MarshalStructured(v, FormatTOON)
	require.NoError(t, err)
	for i := 0; i < 4; i++ {
		b, err := MarshalStructured(v, FormatTOON)
		require.NoError(t, err)
		assert.Equal(t, first, b)
	}
}

func TestPrintStructuredErrorReturnsExitCodes(t *testing.T) {
	env := ErrorEnvelope{Error: ErrorBody{
		Code:     "auth-missing",
		Message:  "Jira API token is not configured",
		Recovery: "Run 'jira init' to configure authentication",
	}}

	t.Run("operational error exits 1", func(t *testing.T) {
		code := PrintStructuredError(env, FormatTOON, false)
		assert.Equal(t, 1, code)
	})

	t.Run("usage error exits 2", func(t *testing.T) {
		code := PrintStructuredError(env, FormatTOON, true)
		assert.Equal(t, 2, code)
	})
}

func TestErrorEnvelopeOmitsOptionalFieldsWhenEmpty(t *testing.T) {
	env := ErrorEnvelope{Error: ErrorBody{
		Code:    "not-found",
		Message: "issue does not exist",
	}}
	b, err := MarshalStructured(env, FormatTOON)
	require.NoError(t, err)
	// recovery and hint are omitted; error block keeps code+message only.
	got := string(b)
	assert.True(t, strings.HasPrefix(got, "error:"), "want error block, got %q", got)
	assert.NotContains(t, got, "recovery")
	assert.NotContains(t, got, "hint")
}
