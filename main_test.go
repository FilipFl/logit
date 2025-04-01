package main

import (
	"errors"
	"testing"
	"time"

	"github.com/FilipFl/logit/configuration"
	"github.com/FilipFl/logit/git"
	"github.com/FilipFl/logit/prompter"
	"github.com/FilipFl/logit/timer"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

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
		expectedError            error
	}{
		{
			name:          "WithTaskFlag",
			taskFlag:      "PROJ-123",
			expectedTask:  "PROJ-123",
			expectedError: nil,
		},
		{
			name:              "WithTaskFlagPassedBadAndPromptedProperly",
			taskFlag:          "not a task",
			prompterResponses: []string{"PROJ-123"},
			expectedTask:      "PROJ-123",
			expectedError:     nil,
		},
		{
			name:              "WithTaskFlagPassedProperlyAndPromptedBad",
			taskFlag:          "not a task",
			prompterResponses: []string{"also not a task"},
			expectedError:     errorNoJiraTicket,
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
			expectedError: errorNoTargetToLogWork,
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
			name:          "WithInvalidGitBranchAndErrorPrompt",
			gitBranch:     "invalid-branch",
			expectedError: prompter.ErrorScanningUserInput,
		},
		{
			name:              "WithInvalidGitBranchAndPassedTask",
			gitBranch:         "invalid-branch",
			prompterResponses: []string{"PRO-123"},
			prompterErrors:    []error{nil},
			expectedTask:      "PRO-123",
		},
		{
			name:              "WithGitError",
			gitError:          errors.New("Git error"),
			prompterResponses: []string{"PRO-123"},
			prompterErrors:    []error{nil},
			expectedTask:      "PRO-123",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			cfgHandlerMock := configuration.NewMockConfigurationHandler()
			if tc.config != nil {
				cfgHandlerMock.SetConfig(tc.config)
			}
			prompterMock := prompter.NewMockPrompter()
			if tc.prompterResponses != nil {
				prompterMock.SetStringResponses(tc.prompterResponses, tc.prompterErrors)
			}
			if tc.prompterApproveResponses != nil {
				prompterMock.SetApproveResponses(tc.prompterApproveResponses, tc.prompterApproveErrors)
			}
			gitHandlerMock := git.NewMockGitHandler()
			if tc.gitBranch != "" {
				gitHandlerMock.Branch = tc.gitBranch
				gitHandlerMock.Error = nil
			} else if tc.gitError != nil {
				gitHandlerMock.Branch = ""
				gitHandlerMock.Error = tc.gitError
			}

			cmd := &cobra.Command{}

			if tc.taskFlag != "" {
				cmd.Flags().String("task", tc.taskFlag, "Jira task ID")
			}
			if tc.aliasFlag != "" {
				cmd.Flags().String("alias", tc.aliasFlag, "Task alias")
			}

			task, err := determineTask(cmd, cfgHandlerMock, prompterMock, gitHandlerMock)

			if tc.expectedError != nil {
				assert.Equal(t, tc.expectedError, err)
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
		timer                    timer.Timer
		expectedDuration         time.Duration
		expectedError            error
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
			expectedError:    errorNoSnapshot,
		},
		{
			name:             "WithoutAnyFlagWithSnapshot",
			expectedDuration: time.Duration(1) * time.Hour,
			timer:            timer.NewMockTimer("2025-01-04T14:00:00.000Z"),
			config:           &configuration.Config{Snapshot: timer.ParseStringToTime("2025-01-04T13:00:00.000Z")},
		},
		{
			name:                     "WithoutAnyFlagWith9hSnapshotAndApprove",
			expectedDuration:         time.Duration(9) * time.Hour,
			timer:                    timer.NewMockTimer("2025-01-04T14:00:00.000Z"),
			config:                   &configuration.Config{Snapshot: timer.ParseStringToTime("2025-01-04T05:00:00.000Z")},
			prompterApproveResponses: []bool{true},
			prompterApproveErrors:    []error{nil},
		},
		{
			name:                     "WithoutAnyFlagWith9hSnapshotAndDecline",
			expectedDuration:         time.Duration(0),
			expectedError:            errorOperationAborted,
			timer:                    timer.NewMockTimer("2025-01-04T14:00:00.000Z"),
			config:                   &configuration.Config{Snapshot: timer.ParseStringToTime("2025-01-04T05:00:00.000Z")},
			prompterApproveResponses: []bool{false},
			prompterApproveErrors:    []error{nil},
		},
		{
			name:             "With120MinutesFlag",
			minutes:          120,
			expectedDuration: time.Duration(2) * time.Hour,
		},
		{
			name:                     "With9HoursFlagAndApprove",
			hours:                    9,
			expectedDuration:         time.Duration(9) * time.Hour,
			prompterApproveResponses: []bool{true},
			prompterApproveErrors:    []error{nil},
		},
		{
			name:                     "With9HoursFlagAndDecline",
			hours:                    9,
			expectedDuration:         time.Duration(0),
			prompterApproveResponses: []bool{false},
			prompterApproveErrors:    []error{nil},
			expectedError:            errorOperationAborted,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			cfgHandlerMock := configuration.NewMockConfigurationHandler()
			if tc.config != nil {
				cfgHandlerMock.SetConfig(tc.config)
			}
			prompterMock := prompter.NewMockPrompter()
			if tc.prompterResponses != nil {
				prompterMock.SetStringResponses(tc.prompterResponses, tc.prompterErrors)
			}
			if tc.prompterApproveResponses != nil {
				prompterMock.SetApproveResponses(tc.prompterApproveResponses, tc.prompterApproveErrors)
			}
			timerMock := timer.NewMockTimer("2025-01-04T14:00:00.000Z")
			if tc.timer != nil {
				timerMock = tc.timer.(*timer.MockTimer)
			}

			cmd := &cobra.Command{}
			if tc.hours != 0 {
				cmd.Flags().Int("hours", tc.hours, "Hours spent")
			}
			if tc.minutes != 0 {
				cmd.Flags().Int("minutes", tc.minutes, "Minutes spent")
			}

			result, err := parseDuration(cmd, cfgHandlerMock, prompterMock, timerMock)

			if tc.expectedError != nil {
				assert.Equal(t, tc.expectedError, err)
				assert.Equal(t, time.Duration(0), result)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.expectedDuration, result)
			}
		})
	}
}
