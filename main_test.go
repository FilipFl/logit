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

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfgHandlerMock := configuration.NewMockConfigurationHandler()
			if tt.config != nil {
				cfgHandlerMock.SetConfig(tt.config)
			}
			prompterMock := prompter.NewMockPrompter()
			if tt.prompterResponses != nil {
				prompterMock.SetStringResponses(tt.prompterResponses, tt.prompterErrors)
			}
			if tt.prompterApproveResponses != nil {
				prompterMock.SetApproveResponses(tt.prompterApproveResponses, tt.prompterApproveErrors)
			}
			gitHandlerMock := git.NewMockGitHandler()
			if tt.gitBranch != "" {
				gitHandlerMock.Branch = tt.gitBranch
				gitHandlerMock.Error = nil
			} else if tt.gitError != nil {
				gitHandlerMock.Branch = ""
				gitHandlerMock.Error = tt.gitError
			}

			cmd := &cobra.Command{}

			if tt.taskFlag != "" {
				cmd.Flags().String("task", tt.taskFlag, "Jira task ID")
			}
			if tt.aliasFlag != "" {
				cmd.Flags().String("alias", tt.aliasFlag, "Task alias")
			}

			task, err := determineTask(cmd, cfgHandlerMock, prompterMock, gitHandlerMock)

			if tt.expectedError != nil {
				assert.Equal(t, tt.expectedError, err)
				assert.Equal(t, "", task)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedTask, task)
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

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfgHandlerMock := configuration.NewMockConfigurationHandler()
			if tt.config != nil {
				cfgHandlerMock.SetConfig(tt.config)
			}
			prompterMock := prompter.NewMockPrompter()
			if tt.prompterResponses != nil {
				prompterMock.SetStringResponses(tt.prompterResponses, tt.prompterErrors)
			}
			if tt.prompterApproveResponses != nil {
				prompterMock.SetApproveResponses(tt.prompterApproveResponses, tt.prompterApproveErrors)
			}
			timerMock := timer.NewMockTimer("2025-01-04T14:00:00.000Z")
			if tt.timer != nil {
				timerMock = tt.timer.(*timer.MockTimer)
			}

			cmd := &cobra.Command{}
			if tt.hours != 0 {
				cmd.Flags().Int("hours", tt.hours, "Hours spent")
			}
			if tt.minutes != 0 {
				cmd.Flags().Int("minutes", tt.minutes, "Minutes spent")
			}

			result, err := parseDuration(cmd, cfgHandlerMock, prompterMock, timerMock)

			if tt.expectedError != nil {
				assert.Equal(t, tt.expectedError, err)
				assert.Equal(t, time.Duration(0), result)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedDuration, result)
			}
		})
	}
}

func TestParseDateFromString(t *testing.T) {

	tests := []struct {
		name          string
		input         string
		expectedTime  time.Time
		expectedError error
		customTimer   timer.Timer
	}{
		{
			"valid date - dot",
			"12.05",
			time.Date(2025, 5, 12, 14, 0, 0, 0, time.UTC),
			nil,
			timer.NewMockTimer("2025-01-04T14:00:00.000Z"),
		},
		{
			"valid date - dash ",
			"04-01",
			time.Date(2025, 1, 4, 14, 0, 0, 0, time.UTC),
			nil,
			timer.NewMockTimer("2025-01-04T14:00:00.000Z"),
		},
		{
			"invalid format",
			"04/01",
			time.Time{},
			errorInvalidDateFormat,
			timer.NewMockTimer("2025-01-04T14:00:00.000Z"),
		},
		{
			"invalid month",
			"15.13",
			time.Time{},
			errorInvalidMonth,
			timer.NewMockTimer("2025-01-04T14:00:00.000Z"),
		},
		{
			"invalid day",
			"31.04",
			time.Time{},
			errorInvalidDay,
			timer.NewMockTimer("2025-01-04T14:00:00.000Z"),
		},
		{
			"leap year valid",
			"29.02",
			time.Date(2024, 2, 29, 14, 0, 0, 0, time.UTC),
			nil,
			timer.NewMockTimer("2024-01-04T14:00:00.000Z"),
		},
		{
			"leap year invalid year",
			"29.02",
			time.Time{},
			errorInvalidDay,
			timer.NewMockTimer("2025-01-04T14:00:00.000Z"),
		},
		{
			"leap year invalid day",
			"30.02",
			time.Time{},
			errorInvalidDay,
			timer.NewMockTimer("2025-01-04T14:00:00.000Z"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parseDateFromString(tt.input, tt.customTimer)

			if tt.expectedError != nil {
				assert.Equal(t, tt.expectedError, err)
				assert.Equal(t, tt.expectedTime, result)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedTime, result)
			}
		})
	}
}

func TestDetermineStarted(t *testing.T) {
	tests := []struct {
		name          string
		timer         timer.Timer
		yesterdayFlag bool
		dateFlag      string
		expectedStart time.Time
		expectedError error
	}{
		{
			name:          "date 01.05",
			timer:         timer.NewMockTimer("2025-01-04T14:00:00.000Z"),
			dateFlag:      "01.05",
			expectedStart: time.Date(2025, 5, 1, 14, 0, 0, 0, time.UTC),
		},
		{
			name:          "yestedar",
			timer:         timer.NewMockTimer("2025-01-04T14:00:00.000Z"),
			yesterdayFlag: true,
			expectedStart: time.Date(2025, 1, 3, 14, 0, 0, 0, time.UTC),
		},
		{
			name:          "When no flags are set, return current time",
			timer:         timer.NewMockTimer("2025-01-04T14:00:00.000Z"),
			expectedStart: time.Date(2025, 1, 4, 14, 0, 0, 0, time.UTC),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := &cobra.Command{}
			if tt.dateFlag != "" {
				cmd.Flags().String("date", tt.dateFlag, "Date in format dd-mm, present year is assumed")
			}
			if tt.yesterdayFlag {
				cmd.Flags().Bool("yesterday", tt.yesterdayFlag, "Log time for yesterday")
			}

			result, err := determineStarted(cmd, tt.timer)

			assert.Equal(t, tt.expectedStart, result)
			assert.NoError(t, err)
		})
	}
}
