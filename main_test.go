package main

import (
	"errors"
	"os"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

func mockGetGitBranch(branch string, err error) func() (string, error) {
	return func() (string, error) {
		return branch, err
	}
}

func TestDetermineTask_WithTaskFlag(t *testing.T) {
	config := Config{
		Aliases: make(map[string]string),
	}
	cmd := &cobra.Command{}
	cmd.Flags().String("task", "PROJ-123", "Jira task ID")

	task, err := determineTask(config, cmd)

	assert.NoError(t, err)
	assert.Equal(t, "PROJ-123", task)
}

func TestDetermineTask_WithTaskFlagAndFullURL(t *testing.T) {
	config := Config{
		Aliases: make(map[string]string),
	}
	cmd := &cobra.Command{}
	cmd.Flags().String("task", "https://some-jira.host.com/PROJ-123", "Jira task ID")

	task, err := determineTask(config, cmd)

	assert.NoError(t, err)
	assert.Equal(t, "PROJ-123", task)
}

func TestDetermineTask_WithAlias(t *testing.T) {
	config := Config{
		Aliases: map[string]string{"bugfix": "BUG-456"},
	}
	cmd := &cobra.Command{}
	cmd.Flags().String("alias", "bugfix", "Task alias")

	task, err := determineTask(config, cmd)

	assert.NoError(t, err)
	assert.Equal(t, "BUG-456", task)
}

// Test when Git branch contains task ID
func TestDetermineTask_WithGitBranch(t *testing.T) {
	getGitBranch = mockGetGitBranch("FEAT-789", nil) // Mocking Git branch
	config := Config{Aliases: make(map[string]string)}
	cmd := &cobra.Command{}
	originalStdin := os.Stdin
	defer func() { os.Stdin = originalStdin }()

	r, w, _ := os.Pipe()
	w.Write([]byte("y"))
	w.Close()

	os.Stdin = r
	task, err := determineTask(config, cmd)

	assert.NoError(t, err)
	assert.Equal(t, "FEAT-789", task)
}

func TestDetermineTask_WithNotOnlyGitBranch(t *testing.T) {
	getGitBranch = mockGetGitBranch("feature/FEAT-789", nil) // Mocking Git branch
	config := Config{Aliases: make(map[string]string)}
	cmd := &cobra.Command{}
	originalStdin := os.Stdin
	defer func() { os.Stdin = originalStdin }()

	r, w, _ := os.Pipe()
	w.Write([]byte("y"))
	w.Close()

	os.Stdin = r
	task, err := determineTask(config, cmd)

	assert.NoError(t, err)
	assert.Equal(t, "FEAT-789", task)
}

func TestDetermineTask_WithInvalidGitBranch(t *testing.T) {
	getGitBranch = mockGetGitBranch("invalid-branch", nil) // Mocking Git branch
	config := Config{Aliases: make(map[string]string)}
	cmd := &cobra.Command{}

	task, err := determineTask(config, cmd)

	assert.Error(t, err)
	assert.Equal(t, "", task)
}

func TestDetermineTask_WithInvalidGitBranchAndPassedTask(t *testing.T) {
	getGitBranch = mockGetGitBranch("invalid-branch", nil) // Mocking Git branch
	config := Config{Aliases: make(map[string]string)}
	cmd := &cobra.Command{}
	originalStdin := os.Stdin
	defer func() { os.Stdin = originalStdin }()

	r, w, _ := os.Pipe()
	w.Write([]byte("PRO-123"))
	w.Close()

	os.Stdin = r

	task, err := determineTask(config, cmd)

	assert.NoError(t, err)
	assert.Equal(t, "PRO-123", task)
}

// Test when Git command fails
func TestDetermineTask_WithGitError(t *testing.T) {
	getGitBranch = mockGetGitBranch("", errors.New("Git error")) // Mock Git failure
	config := Config{Aliases: make(map[string]string)}
	cmd := &cobra.Command{}

	task, err := determineTask(config, cmd)

	assert.Error(t, err)
	assert.Equal(t, "", task)
}
