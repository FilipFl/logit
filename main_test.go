package main

import (
	"context"
	"errors"
	"testing"
	"time"

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

func setupTestContext(prompterMock *prompter.MockPrompter, configMock *configuration.MockConfigurationHandler) context.Context {
	ctx := context.WithValue(context.Background(), prompterKey, prompterMock)
	return context.WithValue(ctx, configKey, configMock)
}

func setupTestCommand(ctx context.Context) *cobra.Command {
	cmd := &cobra.Command{}
	cmd.SetContext(ctx)
	return cmd
}

func setupTestWithMocks(prompterResponses []string, prompterErrors []error, prompterApproveResponses []bool, prompterApproveErrors []error, config *configuration.Config) (*cobra.Command, *prompter.MockPrompter) {
	prompterMock := prompter.NewMockPrompter()
	prompterMock.SetStringResponses(prompterResponses, prompterErrors)
	prompterMock.SetApproveResponses(prompterApproveResponses, prompterApproveErrors)

	configMock := configuration.NewMockConfigurationHandler()
	if config != nil {
		configMock.SetConfig(config)
	}

	ctx := setupTestContext(prompterMock, configMock)
	cmd := setupTestCommand(ctx)

	return cmd, prompterMock
}

func TestDetermineTask(t *testing.T) {
	tests := []struct {
		name                     string
		taskFlag                 string
		aliasFlag                string
		gitBranch                string
		gitError                 error
		prompterResponses        []string
		prompterErrors           []error
		prompterApproveResponses []bool
		prompterApproveErrors    []error
		config                   *configuration.Config
		expectedTask             string
		expectError              bool
	}{
		{
			name:         "WithTaskFlag",
			taskFlag:     "PROJ-123",
			expectedTask: "PROJ-123",
		},
		{
			name:              "WithTaskFlagPassedBadAndPromptedProperly",
			taskFlag:          "not a task",
			prompterResponses: []string{"PROJ-123"},
			expectedTask:      "PROJ-123",
		},
		{
			name:              "WithTaskFlagPassedProperlyAndPromptedBad",
			taskFlag:          "not a task",
			prompterResponses: []string{"also not a task"},
			expectError:       true,
		},
		{
			name:         "WithTaskFlagAndFullURL",
			taskFlag:     "https://some-jira.host.com/PROJ-123",
			expectedTask: "PROJ-123",
		},
		{
			name:      "WithAlias",
			aliasFlag: "bugfix",
			config: &configuration.Config{
				Aliases: map[string]string{"bugfix": "BUG-456"},
			},
			expectedTask: "BUG-456",
		},
		{
			name:              "WithNotSetAliasButPromptedProperly",
			aliasFlag:         "notbugfix",
			prompterResponses: []string{"bugfix"},
			config: &configuration.Config{
				Aliases: map[string]string{"bugfix": "BUG-456"},
			},
			expectedTask: "BUG-456",
		},
		{
			name:              "WithNotSetAliasAndPromptedBadly",
			aliasFlag:         "notbugfix",
			prompterResponses: []string{"BUG-456"},
			config: &configuration.Config{
				Aliases: map[string]string{"bugfix": "BUG-456"},
			},
			expectError: true,
		},
		{
			name:                     "WithGitBranch",
			gitBranch:                "FEAT-789",
			gitError:                 nil,
			expectedTask:             "FEAT-789",
			prompterApproveResponses: []bool{true},
			prompterApproveErrors:    []error{nil},
		},
		{
			name:                     "WithNotOnlyGitBranch",
			gitBranch:                "feature/FEAT-789",
			expectedTask:             "FEAT-789",
			prompterApproveResponses: []bool{true},
			prompterApproveErrors:    []error{nil},
		},
		{
			name:        "WithInvalidGitBranch",
			gitBranch:   "invalid-branch",
			expectError: true,
		},
		{
			name:              "WithInvalidGitBranchAndPassedTask",
			gitBranch:         "invalid-branch",
			prompterResponses: []string{"PRO-123"},
			prompterErrors:    []error{nil},
			expectedTask:      "PRO-123",
		},
		{
			name:        "WithGitError",
			gitError:    errors.New("Git error"),
			expectError: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			getGitBranch = mockGetGitBranch(tc.gitBranch, tc.gitError)

			cmd, _ := setupTestWithMocks(tc.prompterResponses, tc.prompterErrors, tc.prompterApproveResponses, tc.prompterApproveErrors, tc.config)

			if tc.taskFlag != "" {
				cmd.Flags().String("task", tc.taskFlag, "Jira task ID")
			}
			if tc.aliasFlag != "" {
				cmd.Flags().String("alias", tc.aliasFlag, "Task alias")
			}

			task, err := determineTask(cmd)

			if tc.expectError {
				assert.Error(t, err)
				assert.Equal(t, "", task)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.expectedTask, task)
			}
		})
	}
}

func TestParseDuration(t *testing.T) {
	tests := []struct {
		name                     string
		hours                    int
		minutes                  int
		prompterResponses        []string
		prompterErrors           []error
		prompterApproveResponses []bool
		prompterApproveErrors    []error
		config                   *configuration.Config
		expectedDuration         time.Duration
		expectError              bool
	}{
		{
			name:             "WithHoursFlag",
			hours:            1,
			expectedDuration: time.Duration(1) * time.Hour,
		},
		{
			name:             "WithMinutesFlag",
			minutes:          45,
			expectedDuration: time.Duration(45) * time.Minute,
		},
		{
			name:             "WithHoursAndMinutesFlag",
			hours:            2,
			minutes:          45,
			expectedDuration: time.Duration(45)*time.Minute + time.Duration(2)*time.Hour,
		},
		{
			name:             "WithoutAnyFlagAndNoSnapshot",
			expectedDuration: time.Duration(0),
			expectError:      true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {

			cmd, _ := setupTestWithMocks(tc.prompterResponses, tc.prompterErrors, tc.prompterApproveResponses, tc.prompterApproveErrors, tc.config)

			if tc.hours != 0 {
				cmd.Flags().Int("hours", tc.hours, "Hours spent")
			}
			if tc.minutes != 0 {
				cmd.Flags().Int("minutes", tc.minutes, "Minutes spent")
			}

			result, err := parseDuration(cmd)

			if tc.expectError {
				assert.Error(t, err)
				assert.Equal(t, time.Duration(0), result)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.expectedDuration, result)
			}
		})
	}
}
