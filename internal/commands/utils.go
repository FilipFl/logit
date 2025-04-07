package commands

import (
	"fmt"
	"regexp"
	"time"

	"github.com/FilipFl/logit/internal/configuration"
	"github.com/FilipFl/logit/internal/git"
	"github.com/FilipFl/logit/internal/prompter"
	"github.com/FilipFl/logit/internal/timer"
	"github.com/spf13/cobra"
)

const provideTaskMessage = "Provide task ID or task URL:"

func extractJiraTicket(arg string) (string, error) {
	re := regexp.MustCompile(`([A-Z]+-\d+)`)
	matches := re.FindStringSubmatch(arg)
	if len(matches) > 1 {
		return matches[1], nil
	}
	return "", errorNoJiraTicket
}

func promptForTask(prompter prompter.Prompter, msg string) (string, error) {
	userPromptedMessage, err := prompter.PromptForString(msg, provideTaskMessage)
	if err != nil {
		return "", err
	}
	return extractJiraTicket(userPromptedMessage)
}

func determineTask(cmd *cobra.Command, cfgHandler configuration.ConfigurationHandler, prompter prompter.Prompter, gitHandler git.GitHandler) (string, error) {
	aliases := cfgHandler.LoadConfig().Aliases
	resultTask := ""
	task, _ := cmd.Flags().GetString("task")
	alias, _ := cmd.Flags().GetString("alias")
	if alias != "" {
		if resultTask, exists := aliases[alias]; exists {
			return resultTask, nil
		} else {
			alias, err := prompter.PromptForString("Passed alias was not found.", "Please pass proper alias this time")
			if err != nil {
				return "", err
			}
			if resultTask, exists := aliases[alias]; exists {
				return resultTask, nil
			} else {
				fmt.Println("Passed alias was not found.")
				return "", errorNoTargetToLogWork
			}
		}
	}

	if task != "" {
		resultTask, err := extractJiraTicket(task)
		if err == nil {
			return resultTask, nil
		}
		return promptForTask(prompter, "There is no jira task in value passed to task flag.")
	}
	gitBranch, err := gitHandler.GetGitBranch()
	if err != nil {
		return promptForTask(prompter, "Current directory is not a git repository or something failed during branch name extraction.")
	}

	resultTask, err = extractJiraTicket(gitBranch)
	if err != nil {
		return promptForTask(prompter, "Current branch name does not contain task ID.")
	}

	proceed, err := prompter.PromptForApprove((fmt.Sprintf("Detected task ID %s in current branch name.", resultTask)))

	if err != nil {
		return promptForTask(prompter, "Error scanning proceed approve.")
	}

	if proceed {
		return resultTask, nil
	}

	return "", errorOperationAborted
}

func assertFlagsAreValid(cmd *cobra.Command, timer timer.Timer) error {
	task, _ := cmd.Flags().GetString("task")
	alias, _ := cmd.Flags().GetString("alias")
	yesterday, _ := cmd.Flags().GetBool("yesterday")
	date, _ := cmd.Flags().GetString("date")
	hours, _ := cmd.Flags().GetInt("hours")
	minutes, _ := cmd.Flags().GetInt("minutes")

	if task != "" && alias != "" {
		return errorAliasAndTask
	}
	if yesterday && date != "" {
		return errorYesterdayAndDate
	}
	if hours == 0 && minutes == 0 && (yesterday || date != "") {
		return errorSnapshotNotToday
	}
	if hours < 0 || minutes < 0 {
		return errorWrongDuration
	}
	if date != "" {
		_, _, err := extractNewDayAndMonth(date, timer)
		if err != nil {
			return err
		}
	}

	return nil
}

func parseDuration(cmd *cobra.Command, cfgHandler configuration.ConfigurationHandler, prompter prompter.Prompter, timer timer.Timer) (time.Duration, bool, error) {
	result := time.Duration(0)
	fromSnapshot := false
	hours, _ := cmd.Flags().GetInt("hours")
	minutes, _ := cmd.Flags().GetInt("minutes")
	if hours != 0 || minutes != 0 {
		result = time.Duration(hours)*time.Hour + time.Duration(minutes)*time.Minute
	} else {
		fromSnapshot = true
		if cfgHandler.LoadConfig().Snapshot == nil {
			return time.Duration(0), fromSnapshot, errorNoSnapshot
		}
		now := timer.Now()
		result = now.Sub(*cfgHandler.LoadConfig().Snapshot)
	}
	if int(result.Hours()) > 8 {
		proceed, err := prompter.PromptForApprove(fmt.Sprintf("Are You sure you want to log %d hours and %d minutes?", int(result.Hours()), int(result.Minutes())%60))
		if err != nil {
			return time.Duration(0), fromSnapshot, err
		}
		if proceed {
			return result, fromSnapshot, nil
		} else {
			return time.Duration(0), fromSnapshot, errorOperationAborted
		}
	}

	return result, fromSnapshot, nil
}

func parseDateFromString(s string, timer timer.Timer) (time.Time, error) {
	t := timer.Now()
	newDay, newMonth, err := extractNewDayAndMonth(s, timer)
	if err != nil {
		return time.Time{}, err
	}

	return time.Date(t.Year(), time.Month(newMonth), newDay, t.Hour(), t.Minute(), t.Second(), t.Nanosecond(), t.Location()), nil
}

func extractNewDayAndMonth(s string, tt timer.Timer) (int, int, error) {
	t := tt.Now()
	dotFormat := regexp.MustCompile(`^(\d{2})\.(\d{2})$`)
	dashFormat := regexp.MustCompile(`^(\d{2})-(\d{2})$`)

	var newDay, newMonth int

	switch {
	case dotFormat.MatchString(s):
		fmt.Sscanf(s, "%02d.%02d", &newDay, &newMonth)
	case dashFormat.MatchString(s):
		fmt.Sscanf(s, "%02d-%02d", &newDay, &newMonth)
	default:
		return 0, 0, errorInvalidDateFormat
	}

	if newMonth < 1 || newMonth > 12 {
		return 0, 0, errorInvalidMonth
	}

	newTime := time.Date(t.Year(), time.Month(newMonth), 1, t.Hour(), t.Minute(), t.Second(), t.Nanosecond(), t.Location())
	lastDayOfMonth := newTime.AddDate(0, 1, -1).Day()

	if newDay < 1 || newDay > lastDayOfMonth {
		return 0, 0, errorInvalidDay
	}
	return newDay, newMonth, nil
}

func safeSubtractDay(t time.Time) time.Time {
	hour, min, sec, nsec := t.Hour(), t.Minute(), t.Second(), t.Nanosecond()
	yesterday := t.AddDate(0, 0, -1)
	return time.Date(yesterday.Year(), yesterday.Month(), yesterday.Day(), hour, min, sec, nsec, t.Location())
}

func determineStarted(cmd *cobra.Command, timer timer.Timer) (time.Time, error) {
	yesterday, _ := cmd.Flags().GetBool("yesterday")
	date, _ := cmd.Flags().GetString("date")

	if date != "" {
		return parseDateFromString(date, timer)
	}
	if yesterday {
		return safeSubtractDay(timer.Now()), nil
	}

	return timer.Now(), nil
}
