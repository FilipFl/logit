package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"time"

	"github.com/FilipFl/logit/configuration"
	"github.com/FilipFl/logit/git"
	"github.com/FilipFl/logit/prompter"
	"github.com/spf13/cobra"
)

type Worklog struct {
	TimeSpent string `json:"timeSpent"`
	Started   string `json:"started"`
	Comment   string `json:"comment,omitempty"`
}

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

	return "", errorNoTargetToLogWork
}

func parseDuration(cmd *cobra.Command, cfgHandler configuration.ConfigurationHandler, prompter prompter.Prompter) (time.Duration, error) {
	result := time.Duration(0)
	hours, _ := cmd.Flags().GetInt("hours")
	minutes, _ := cmd.Flags().GetInt("minutes")
	if hours != 0 || minutes != 0 {
		result = time.Duration(hours)*time.Hour + time.Duration(minutes)*time.Minute
	} else {
		if cfgHandler.LoadConfig().Snapshot == nil {
			return time.Duration(0), errorNoSnapshot
		}
		now := time.Now()
		result = now.Sub(*cfgHandler.LoadConfig().Snapshot)
	}
	if int(result.Hours()) > 8 {
		proceed, err := prompter.PromptForApprove(fmt.Sprintf("Are You sure you want to log %d hours and %d minutes?", int(result.Hours()), int(result.Minutes())%60))
		if err != nil {
			return time.Duration(0), err
		}
		if proceed {
			return result, nil
		} else {
			return time.Duration(0), errorWrongDuration
		}
	}

	return result, nil
}

func logTimeToJira(ticket string, duration time.Duration, comment string, cfgHandler configuration.ConfigurationHandler) error {
	cfg := cfgHandler.LoadConfig()
	timeSpent := fmt.Sprintf("%dh %dm", int(duration.Hours()), int(duration.Minutes())%60)
	url := fmt.Sprintf("%s/rest/api/3/issue/%s/worklog", cfg.JiraHost, ticket)
	worklog := Worklog{
		TimeSpent: timeSpent,
		Started:   time.Now().Format("2006-01-02T15:04:05.000-0700"),
		Comment:   comment,
	}
	jsonData, _ := json.Marshal(worklog)

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	token := cfgHandler.GetToken()

	if cfg.JiraEmail == "" {
		return errorEmailNotConfigured
	}
	if token == "" {
		return errorTokenNotConfigured
	}

	req.SetBasicAuth(cfg.JiraEmail, token)
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to log time: %s", string(body))
	}

	return nil
}
