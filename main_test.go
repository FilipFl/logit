package main

import (
	"context"
	"errors"
	"testing"

	"github.com/FilipFl/logit/configuration"
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
	prompterMock := prompter.NewMockPrompter()
	configurationHandlerMock := configuration.NewMockConfigurationHandler()
	ctx := context.WithValue(context.Background(), prompterKey, prompterMock)
	ctx = context.WithValue(ctx, configKey, configurationHandlerMock)
	cmd := &cobra.Command{}
	cmd.SetContext(ctx)
	cmd.Flags().String("task", "PROJ-123", "Jira task ID")

	task, err := determineTask(cmd)

	assert.NoError(t, err)
	assert.Equal(t, "PROJ-123", task)
}

func TestDetermineTask_WithTaskFlagPassedBadAndPromptedProperly(t *testing.T) {
	prompterMock := prompter.NewMockPrompter()
	prompterMock.SetStringResponses([]string{"PROJ-123"}, []error{nil})
	configurationHandlerMock := configuration.NewMockConfigurationHandler()
	ctx := context.WithValue(context.Background(), prompterKey, prompterMock)
	ctx = context.WithValue(ctx, configKey, configurationHandlerMock)
	cmd := &cobra.Command{}
	cmd.SetContext(ctx)
	cmd.Flags().String("task", "not a task", "Jira task ID")

	task, err := determineTask(cmd)

	assert.NoError(t, err)
	assert.Equal(t, "PROJ-123", task)
}

func TestDetermineTask_WithTaskFlagPassedProperlyAndPromptedBad(t *testing.T) {
	prompterMock := prompter.NewMockPrompter()
	prompterMock.SetStringResponses([]string{"also not a task"}, []error{nil})
	configurationHandlerMock := configuration.NewMockConfigurationHandler()
	ctx := context.WithValue(context.Background(), prompterKey, prompterMock)
	ctx = context.WithValue(ctx, configKey, configurationHandlerMock)
	cmd := &cobra.Command{}
	cmd.SetContext(ctx)
	cmd.Flags().String("task", "not a task", "Jira task ID")

	task, err := determineTask(cmd)

	assert.Error(t, err)
	assert.Equal(t, "", task)
}

func TestDetermineTask_WithTaskFlagAndFullURL(t *testing.T) {
	prompterMock := prompter.NewMockPrompter()
	configurationHandlerMock := configuration.NewMockConfigurationHandler()
	ctx := context.WithValue(context.Background(), prompterKey, prompterMock)
	ctx = context.WithValue(ctx, configKey, configurationHandlerMock)
	cmd := &cobra.Command{}
	cmd.SetContext(ctx)
	cmd.Flags().String("task", "https://some-jira.host.com/PROJ-123", "Jira task ID")

	task, err := determineTask(cmd)

	assert.NoError(t, err)
	assert.Equal(t, "PROJ-123", task)
}

func TestDetermineTask_WithAlias(t *testing.T) {
	prompterMock := prompter.NewMockPrompter()
	configurationHandlerMock := configuration.NewMockConfigurationHandler()
	configurationHandlerMock.SetConfig(&configuration.Config{
		Aliases: map[string]string{"bugfix": "BUG-456"},
	})
	ctx := context.WithValue(context.Background(), prompterKey, prompterMock)
	ctx = context.WithValue(ctx, configKey, configurationHandlerMock)
	cmd := &cobra.Command{}
	cmd.SetContext(ctx)
	cmd.Flags().String("alias", "bugfix", "Task alias")

	task, err := determineTask(cmd)

	assert.NoError(t, err)
	assert.Equal(t, "BUG-456", task)
}

func TestDetermineTask_WithNotSetAliasButPromptedProperly(t *testing.T) {
	prompterMock := prompter.NewMockPrompter()
	prompterMock.SetStringResponses([]string{"bugfix"}, []error{nil})
	configurationHandlerMock := configuration.NewMockConfigurationHandler()
	configurationHandlerMock.SetConfig(&configuration.Config{
		Aliases: map[string]string{"bugfix": "BUG-456"},
	})
	ctx := context.WithValue(context.Background(), prompterKey, prompterMock)
	ctx = context.WithValue(ctx, configKey, configurationHandlerMock)
	cmd := &cobra.Command{}
	cmd.SetContext(ctx)
	cmd.Flags().String("alias", "notbugfix", "Task alias")

	task, err := determineTask(cmd)

	assert.NoError(t, err)
	assert.Equal(t, "BUG-456", task)
}

func TestDetermineTask_WithNotSetAliasAndPromptedBadly(t *testing.T) {
	prompterMock := prompter.NewMockPrompter()
	prompterMock.SetStringResponses([]string{"BUG-456"}, []error{nil})
	configurationHandlerMock := configuration.NewMockConfigurationHandler()
	configurationHandlerMock.SetConfig(&configuration.Config{
		Aliases: map[string]string{"bugfix": "BUG-456"},
	})
	ctx := context.WithValue(context.Background(), prompterKey, prompterMock)
	ctx = context.WithValue(ctx, configKey, configurationHandlerMock)
	cmd := &cobra.Command{}
	cmd.SetContext(ctx)
	cmd.Flags().String("alias", "notbugfix", "Task alias")

	task, err := determineTask(cmd)

	assert.Error(t, err)
	assert.Equal(t, "", task)
}

func TestDetermineTask_WithGitBranch(t *testing.T) {
	getGitBranch = mockGetGitBranch("FEAT-789", nil)
	prompterMock := prompter.NewMockPrompter()
	prompterMock.SetApproveResponses([]bool{true}, []error{nil})
	configurationHandlerMock := configuration.NewMockConfigurationHandler()
	ctx := context.WithValue(context.Background(), prompterKey, prompterMock)
	ctx = context.WithValue(ctx, configKey, configurationHandlerMock)
	cmd := &cobra.Command{}
	cmd.SetContext(ctx)
	task, err := determineTask(cmd)

	assert.NoError(t, err)
	assert.Equal(t, "FEAT-789", task)
}

func TestDetermineTask_WithNotOnlyGitBranch(t *testing.T) {
	getGitBranch = mockGetGitBranch("feature/FEAT-789", nil)
	prompterMock := prompter.NewMockPrompter()
	prompterMock.SetApproveResponses([]bool{true}, []error{nil})
	configurationHandlerMock := configuration.NewMockConfigurationHandler()
	ctx := context.WithValue(context.Background(), prompterKey, prompterMock)
	ctx = context.WithValue(ctx, configKey, configurationHandlerMock)
	cmd := &cobra.Command{}
	cmd.SetContext(ctx)
	task, err := determineTask(cmd)

	assert.NoError(t, err)
	assert.Equal(t, "FEAT-789", task)
}

func TestDetermineTask_WithInvalidGitBranch(t *testing.T) {
	getGitBranch = mockGetGitBranch("invalid-branch", nil)
	prompterMock := prompter.NewMockPrompter()
	prompterMock.SetApproveResponses([]bool{true}, []error{nil})
	configurationHandlerMock := configuration.NewMockConfigurationHandler()
	ctx := context.WithValue(context.Background(), prompterKey, prompterMock)
	ctx = context.WithValue(ctx, configKey, configurationHandlerMock)
	cmd := &cobra.Command{}
	cmd.SetContext(ctx)
	task, err := determineTask(cmd)

	assert.Error(t, err)
	assert.Equal(t, "", task)
}

func TestDetermineTask_WithInvalidGitBranchAndPassedTask(t *testing.T) {
	getGitBranch = mockGetGitBranch("invalid-branch", nil)
	prompterMock := prompter.NewMockPrompter()
	prompterMock.SetStringResponses([]string{"PRO-123"}, []error{nil})
	configurationHandlerMock := configuration.NewMockConfigurationHandler()
	ctx := context.WithValue(context.Background(), prompterKey, prompterMock)
	ctx = context.WithValue(ctx, configKey, configurationHandlerMock)
	cmd := &cobra.Command{}
	cmd.SetContext(ctx)
	task, err := determineTask(cmd)

	assert.NoError(t, err)
	assert.Equal(t, "PRO-123", task)
}

// Test when Git command fails
func TestDetermineTask_WithGitError(t *testing.T) {
	getGitBranch = mockGetGitBranch("", errors.New("Git error")) // Mock Git failure
	prompterMock := prompter.NewMockPrompter()
	configurationHandlerMock := configuration.NewMockConfigurationHandler()
	ctx := context.WithValue(context.Background(), prompterKey, prompterMock)
	ctx = context.WithValue(ctx, configKey, configurationHandlerMock)
	cmd := &cobra.Command{}
	cmd.SetContext(ctx)
	task, err := determineTask(cmd)

	assert.Error(t, err)
	assert.Equal(t, "", task)
}
