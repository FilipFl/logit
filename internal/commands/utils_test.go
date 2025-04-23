package commands

import (
	"errors"
	"testing"
	"time"

	"github.com/FilipFl/logit/internal/configuration"
	"github.com/FilipFl/logit/internal/git"
	"github.com/FilipFl/logit/internal/prompter"
	"github.com/FilipFl/logit/internal/timer"

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
		config                   *configuration.Cfg
		configError              error
		expectedTask             string
		expectedError            error
		forceFlag                bool
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
			expectedError:     errorNoJiraTask,
		},
		{
			name:         "WithTaskFlagAndFullURL",
			taskFlag:     "https://some-jira.host.com/PROJ-123",
			expectedTask: "PROJ-123",
		},
		{
			name:      "WithAlias",
			aliasFlag: "bugfix",
			config: &configuration.Cfg{
				Aliases: map[string]string{"bugfix": "BUG-456"},
			},
			expectedTask: "BUG-456",
		},
		{
			name:              "WithNotSetAliasButPromptedProperly",
			aliasFlag:         "notbugfix",
			prompterResponses: []string{"bugfix"},
			config: &configuration.Cfg{
				Aliases: map[string]string{"bugfix": "BUG-456"},
			},
			configError:  configuration.ErrorAliasDontExists,
			expectedTask: "BUG-456",
		},
		{
			name:              "WithNotSetAliasAndPromptedBadly",
			aliasFlag:         "notbugfix",
			prompterResponses: []string{"BUG-456"},
			config: &configuration.Cfg{
				Aliases: map[string]string{"bugfix": "BUG-456"},
			},
			configError:   configuration.ErrorAliasDontExists,
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
			name:         "WithGitBranchAndTrustGitBranchTrue",
			gitBranch:    "FEAT-789",
			gitError:     nil,
			expectedTask: "FEAT-789",
			config: &configuration.Cfg{
				TrustGitBranch: true,
			},
		},
		{
			name:         "WithGitBranchAndForceFlag",
			gitBranch:    "FEAT-789",
			gitError:     nil,
			expectedTask: "FEAT-789",
			forceFlag:    true,
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
			cfgHandlerMock := configuration.NewMockConfig(nil)
			if tt.config != nil {
				cfgHandlerMock.SetConfig(tt.config)
			}
			if tt.configError != nil {
				cfgHandlerMock.SetError(tt.configError)
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
			cmd.Flags().String("task", tt.taskFlag, "")
			cmd.Flags().String("alias", tt.aliasFlag, "")

			task, err := determineTask(cmd, cfgHandlerMock, prompterMock, gitHandlerMock, tt.forceFlag)

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
		config                   *configuration.Cfg
		timer                    timer.Timer
		expectedDuration         time.Duration
		expectedFromSnapshot     bool
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
			name:                 "WithoutAnyFlagAndNoSnapshot",
			expectedDuration:     time.Duration(0),
			expectedError:        errorNoSnapshot,
			expectedFromSnapshot: true,
		},
		{
			name:                 "WithoutAnyFlagWithSnapshot",
			expectedDuration:     time.Duration(1) * time.Hour,
			timer:                timer.NewMockTimer("2025-01-04T14:00:00.000Z"),
			config:               &configuration.Cfg{Snapshot: timer.ParseStringToTime("2025-01-04T13:00:00.000Z")},
			expectedFromSnapshot: true,
		},
		{
			name:                     "WithoutAnyFlagWith9hSnapshotAndApprove",
			expectedDuration:         time.Duration(9) * time.Hour,
			timer:                    timer.NewMockTimer("2025-01-04T14:00:00.000Z"),
			config:                   &configuration.Cfg{Snapshot: timer.ParseStringToTime("2025-01-04T05:00:00.000Z")},
			prompterApproveResponses: []bool{true},
			prompterApproveErrors:    []error{nil},
			expectedFromSnapshot:     true,
		},
		{
			name:                     "WithoutAnyFlagWith9hSnapshotAndDecline",
			expectedDuration:         time.Duration(0),
			expectedError:            errorOperationAborted,
			timer:                    timer.NewMockTimer("2025-01-04T14:00:00.000Z"),
			config:                   &configuration.Cfg{Snapshot: timer.ParseStringToTime("2025-01-04T05:00:00.000Z")},
			prompterApproveResponses: []bool{false},
			prompterApproveErrors:    []error{nil},
			expectedFromSnapshot:     true,
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
			cfgHandlerMock := configuration.NewMockConfig(nil)
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
			cmd.Flags().Int("hours", tt.hours, "")
			cmd.Flags().Int("minutes", tt.minutes, "")

			result, fromSnapshot, err := parseDuration(cmd, cfgHandlerMock, prompterMock, timerMock)

			if tt.expectedError != nil {
				assert.Equal(t, tt.expectedError, err)
				assert.Equal(t, time.Duration(0), result)
				assert.Equal(t, tt.expectedFromSnapshot, fromSnapshot)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedDuration, result)
				assert.Equal(t, tt.expectedFromSnapshot, fromSnapshot)
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
			cmd.Flags().String("date", tt.dateFlag, "")
			cmd.Flags().Bool("yesterday", tt.yesterdayFlag, "")

			result, err := determineStarted(cmd, tt.timer)

			assert.Equal(t, tt.expectedStart, result)
			assert.NoError(t, err)
		})
	}
}

func TestAssertFlagsAreValid(t *testing.T) {
	tests := []struct {
		name          string
		task          string
		alias         string
		yesterday     bool
		date          string
		hours         int
		minutes       int
		expectedError error
		customTimer   timer.Timer
	}{
		{
			name:          "task and alias",
			task:          "PRO-123",
			alias:         "alias",
			expectedError: errorAliasAndTask,
			customTimer:   timer.NewMockTimer("2025-01-04T14:00:00.000Z"),
		},
		{
			name:          "yesterday and date",
			yesterday:     true,
			date:          "01.01",
			expectedError: errorYesterdayAndDate,
			customTimer:   timer.NewMockTimer("2025-01-04T14:00:00.000Z"),
		},
		{
			name:          "yesterday and snapshot",
			yesterday:     true,
			expectedError: errorSnapshotNotToday,
			customTimer:   timer.NewMockTimer("2025-01-04T14:00:00.000Z"),
		},
		{
			name:          "date and snapshot",
			date:          "03.03",
			expectedError: errorSnapshotNotToday,
			customTimer:   timer.NewMockTimer("2025-01-04T14:00:00.000Z"),
		},
		{
			name:          "wrong day",
			date:          "32.01",
			hours:         1,
			expectedError: errorInvalidDay,
			customTimer:   timer.NewMockTimer("2025-01-04T14:00:00.000Z"),
		},
		{
			name:          "wrong month",
			date:          "01.13",
			hours:         1,
			expectedError: errorInvalidMonth,
			customTimer:   timer.NewMockTimer("2025-01-04T14:00:00.000Z"),
		},
		{
			name:          "wrong date format",
			date:          "01:12",
			hours:         1,
			expectedError: errorInvalidDateFormat,
			customTimer:   timer.NewMockTimer("2025-01-04T14:00:00.000Z"),
		},
		{
			name:          "negative hours",
			hours:         -1,
			expectedError: errorWrongDuration,
			customTimer:   timer.NewMockTimer("2025-01-04T14:00:00.000Z"),
		},
		{
			name:          "negative minutes",
			minutes:       -30,
			expectedError: errorWrongDuration,
			customTimer:   timer.NewMockTimer("2025-01-04T14:00:00.000Z"),
		},
		{
			name:          "valid",
			task:          "TASK123",
			alias:         "",
			hours:         1,
			minutes:       30,
			expectedError: nil,
			customTimer:   timer.NewMockTimer("2025-01-04T14:00:00.000Z"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := &cobra.Command{}
			cmd.Flags().String("task", tt.task, "")
			cmd.Flags().String("alias", tt.alias, "")
			cmd.Flags().String("date", tt.date, "")
			cmd.Flags().Bool("yesterday", tt.yesterday, "")
			cmd.Flags().Int("hours", tt.hours, "")
			cmd.Flags().Int("minutes", tt.minutes, "")

			err := assertFlagsAreValid(cmd, tt.customTimer)

			if tt.expectedError == nil {
				assert.NoError(t, err)
			} else {
				assert.Equal(t, tt.expectedError, err)
			}
		})
	}
}
