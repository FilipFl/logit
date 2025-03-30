package main

import (
	"context"
	"errors"
	"testing"

	"github.com/FilipFl/logit/prompter"
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
	prompterMock := prompter.NewMockPrompter()
	ctx := context.WithValue(context.Background(), prompterKey, prompterMock)
	cmd.SetContext(ctx)
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
	prompterMock := prompter.NewMockPrompter()
	ctx := context.WithValue(context.Background(), prompterKey, prompterMock)
	cmd.SetContext(ctx)
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
	prompterMock := prompter.NewMockPrompter()
	ctx := context.WithValue(context.Background(), prompterKey, prompterMock)
	cmd.SetContext(ctx)
	cmd.Flags().String("alias", "bugfix", "Task alias")

	task, err := determineTask(config, cmd)

	assert.NoError(t, err)
	assert.Equal(t, "BUG-456", task)
}

func TestDetermineTask_WithGitBranch(t *testing.T) {
	getGitBranch = mockGetGitBranch("FEAT-789", nil)
	config := Config{Aliases: make(map[string]string)}
	cmd := &cobra.Command{}
	prompterMock := prompter.NewMockPrompter()
	prompterMock.SetApproveResponses([]bool{true}, []error{nil})
	ctx := context.WithValue(context.Background(), prompterKey, prompterMock)
	cmd.SetContext(ctx)
	task, err := determineTask(config, cmd)

	assert.NoError(t, err)
	assert.Equal(t, "FEAT-789", task)
}

func TestDetermineTask_WithNotOnlyGitBranch(t *testing.T) {
	getGitBranch = mockGetGitBranch("feature/FEAT-789", nil)
	config := Config{Aliases: make(map[string]string)}
	cmd := &cobra.Command{}
	prompterMock := prompter.NewMockPrompter()
	prompterMock.SetApproveResponses([]bool{true}, []error{nil})
	ctx := context.WithValue(context.Background(), prompterKey, prompterMock)
	cmd.SetContext(ctx)
	task, err := determineTask(config, cmd)

	assert.NoError(t, err)
	assert.Equal(t, "FEAT-789", task)
}

func TestDetermineTask_WithInvalidGitBranch(t *testing.T) {
	getGitBranch = mockGetGitBranch("invalid-branch", nil)
	config := Config{Aliases: make(map[string]string)}
	cmd := &cobra.Command{}
	prompterMock := prompter.NewMockPrompter()
	prompterMock.SetApproveResponses([]bool{true}, []error{nil})
	ctx := context.WithValue(context.Background(), prompterKey, prompterMock)
	cmd.SetContext(ctx)
	task, err := determineTask(config, cmd)

	assert.Error(t, err)
	assert.Equal(t, "", task)
}

func TestDetermineTask_WithInvalidGitBranchAndPassedTask(t *testing.T) {
	getGitBranch = mockGetGitBranch("invalid-branch", nil)
	config := Config{Aliases: make(map[string]string)}
	cmd := &cobra.Command{}
	prompterMock := prompter.NewMockPrompter()
	prompterMock.SetStringResponses([]string{"PRO-123"}, []error{nil})
	ctx := context.WithValue(context.Background(), prompterKey, prompterMock)
	cmd.SetContext(ctx)
	task, err := determineTask(config, cmd)

	assert.NoError(t, err)
	assert.Equal(t, "PRO-123", task)
}

// Test when Git command fails
func TestDetermineTask_WithGitError(t *testing.T) {
	getGitBranch = mockGetGitBranch("", errors.New("Git error")) // Mock Git failure
	config := Config{Aliases: make(map[string]string)}

	cmd := &cobra.Command{}
	prompterMock := prompter.NewMockPrompter()
	ctx := context.WithValue(context.Background(), prompterKey, prompterMock)
	cmd.SetContext(ctx)
	task, err := determineTask(config, cmd)

	assert.Error(t, err)
	assert.Equal(t, "", task)
}
