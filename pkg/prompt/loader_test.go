package prompt_test

import (
	"embed"
	"io/fs"
	"testing"

	"github.com/WinPooh32/go-coder/pkg/prompt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

//go:embed testdata/*.tpl
var templateFS embed.FS

func TestLoad(t *testing.T) {
	t.Parallel()

	prompts, err := prompt.Load(templateFS, "testdata/*.tpl")
	require.NoError(t, err, "should load templates without error")

	assert.Len(t, prompts, 2, "should have two templates loaded")
	assert.Contains(t, prompts, "greeting", "should contain 'greeting' template")
	assert.Contains(t, prompts, "farewell", "should contain 'farewell' template")
}

func TestExecute_Greeting(t *testing.T) {
	t.Parallel()

	prompts, err := prompt.Load(templateFS, "testdata/*.tpl")
	require.NoError(t, err, "should load templates without error")

	prompt, ok := prompts["greeting"]
	assert.True(t, ok, "should find 'greeting' template")

	result, err := prompt.Execute(map[string]any{"Name": "Alice"})
	require.NoError(t, err, "should execute template without error")

	expected := "Hello, Alice! Welcome to our service."
	assert.Equal(t, expected, result, "should produce correct output")
}

func TestExecute_Farewell(t *testing.T) {
	t.Parallel()

	prompts, err := prompt.Load(templateFS, "testdata/*.tpl")
	require.NoError(t, err, "should load templates without error")

	prompt, ok := prompts["farewell"]
	assert.True(t, ok, "should find 'farewell' template")

	result, err := prompt.Execute(map[string]any{"Name": "Bob"})
	require.NoError(t, err, "should execute template without error")

	expected := "Goodbye, Bob! We hope to see you again soon."
	assert.Equal(t, expected, result, "should produce correct output")
}

func TestExecute_MissingTemplateField(t *testing.T) {
	t.Parallel()

	prompts, err := prompt.Load(templateFS, "testdata/*.tpl")
	require.NoError(t, err, "should load templates without error")

	prompt, ok := prompts["greeting"]
	assert.True(t, ok, "should find 'greeting' template")

	_, err = prompt.Execute(map[string]any{})
	if assert.Error(t, err) {
		expectedErr := "has no entry for key"
		assert.ErrorContainsf(t, err, expectedErr, "should produce correct error message")
	}
}

func TestLoad_NoTemplatesFound(t *testing.T) {
	t.Parallel()

	var emptyFS fs.FS = embed.FS{}

	prompts, err := prompt.Load(emptyFS)
	if assert.Error(t, err) {
		expectedErr := "no files"
		assert.ErrorContains(t, err, expectedErr)
	}

	assert.Empty(t, prompts, "should not contain any templates")
}
