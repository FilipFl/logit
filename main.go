package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os/exec"
	"regexp"
	"strings"
	"time"

	"github.com/FilipFl/logit/configuration"
	"github.com/FilipFl/logit/prompter"
	"github.com/spf13/cobra"
)

const provideTaskMessage = "Provide task ID or task URL:"

var errorNoJiraTicket = errors.New("no Jira ticket found in passed string")
var errorNoTargetToLogWork = errors.New("no target to log work")
var errorNoSnapshot = errors.New("no start time saved")
var errorWrongDuration = errors.New("duration to log is invalid")

type contextKey string

const prompterKey contextKey = "prompter"
const configKey contextKey = "config"

type Worklog struct {
	TimeSpent string `json:"timeSpent"`
	Started   string `json:"started"`
	Comment   string `json:"comment,omitempty"`
}

var getGitBranch = func() (string, error) {
	cmd := exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD")
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(output)), nil
}

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

func logTimeToJira(ticket string, duration time.Duration, comment string, cfg *configuration.Config) error {
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
	req.SetBasicAuth(cfg.JiraEmail, cfg.JiraToken)
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

func determineTask(cmd *cobra.Command) (string, error) {
	prompter, _ := cmd.Context().Value(prompterKey).(prompter.Prompter)
	configurationHandler, _ := cmd.Context().Value(configKey).(configuration.ConfigurationHandler)
	aliases := configurationHandler.LoadConfig().Aliases
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
	gitBranch, err := getGitBranch()
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

func parseDuration(cmd *cobra.Command) (time.Duration, error) {
	prompter := cmd.Context().Value(prompterKey).(prompter.Prompter)
	configurationHandler, _ := cmd.Context().Value(configKey).(configuration.ConfigurationHandler)
	result := time.Duration(0)
	hours, _ := cmd.Flags().GetInt("hours")
	minutes, _ := cmd.Flags().GetInt("minutes")
	if hours != 0 || minutes != 0 {
		result = time.Duration(hours)*time.Hour + time.Duration(minutes)*time.Minute
	} else {
		if configurationHandler.LoadConfig().Snapshot == nil {
			return time.Duration(0), errorNoSnapshot
		}
		now := time.Now()
		result = now.Sub(*configurationHandler.LoadConfig().Snapshot)
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

func main() {
	var rootCmd = &cobra.Command{Use: "logit"}

	var configCmd = &cobra.Command{
		Use:   "config",
		Short: "Set Jira configuration",
	}

	var aliasCmd = &cobra.Command{
		Use:   "alias",
		Short: "Manage aliases",
	}

	var setHostCmd = &cobra.Command{
		Use:   "set-host [host]",
		Short: "Set Jira host",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			configHandler := cmd.Context().Value(configKey).(configuration.ConfigurationHandler)
			cfg := configHandler.LoadConfig()
			cfg.JiraHost = args[0]
			configHandler.SaveConfig(cfg)
			fmt.Println("Jira host updated.")
		},
	}

	var setTokenCmd = &cobra.Command{
		Use:   "set-token [token]",
		Short: "Set Jira token",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			configHandler := cmd.Context().Value(configKey).(configuration.ConfigurationHandler)
			cfg := configHandler.LoadConfig()
			cfg.JiraToken = args[0]
			configHandler.SaveConfig(cfg)
			fmt.Println("Jira token updated.")
		},
	}

	var setAliasCmd = &cobra.Command{
		Use:   "set [alias] [ticket]",
		Short: "Set an alias for a Jira ticket",
		Args:  cobra.ExactArgs(2),
		Run: func(cmd *cobra.Command, args []string) {
			configHandler := cmd.Context().Value(configKey).(configuration.ConfigurationHandler)
			cfg := configHandler.LoadConfig()
			ticket, err := extractJiraTicket(args[1])
			if err != nil {
				fmt.Println(err)
				return
			}
			cfg.Aliases[args[0]] = ticket
			configHandler.SaveConfig(cfg)

			fmt.Printf("Alias %s set for ticket %s\n", args[0], args[1])
		},
	}

	var removeAliasCmd = &cobra.Command{
		Use:   "remove [alias]",
		Short: "Remove an alias from aliases list",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			configHandler := cmd.Context().Value(configKey).(configuration.ConfigurationHandler)
			cfg := configHandler.LoadConfig()
			if _, exists := cfg.Aliases[args[0]]; exists {
				delete(cfg.Aliases, args[0])
				configHandler.SaveConfig(cfg)
				fmt.Printf("Removed alias %s from aliases list\n", args[0])
				return
			}

			fmt.Printf("Alias %s not found on aliases list\n", args[0])
		},
	}

	var listAliasesCmd = &cobra.Command{
		Use:   "list",
		Short: "Lists all aliases",
		Args:  nil,
		Run: func(cmd *cobra.Command, args []string) {
			configHandler := cmd.Context().Value(configKey).(configuration.ConfigurationHandler)
			cfg := configHandler.LoadConfig()
			for k, v := range cfg.Aliases {
				fmt.Printf("%s: %s \n", k, v)
			}
		},
	}

	var startTimerCmd = &cobra.Command{
		Use:   "start",
		Short: "Start measuring time from this moment",
		Args:  nil,
		Run: func(cmd *cobra.Command, args []string) {
			configHandler := cmd.Context().Value(configKey).(configuration.ConfigurationHandler)
			cfg := configHandler.LoadConfig()
			now := time.Now()
			cfg.Snapshot = &now
			configHandler.SaveConfig(cfg)
			fmt.Println("Started to measure time.")
		},
	}

	var logCmd = &cobra.Command{
		Use:   "log",
		Short: "Log time to Jira",
		Run: func(cmd *cobra.Command, args []string) {
			// comment, _ := cmd.Flags().GetString("comment")

			task, err := determineTask(cmd)
			if err != nil {
				fmt.Println("Error logging time: ", err)
			}
			if task == "" {
				fmt.Println("No target for time logging.")
				return
			}
			duration, err := parseDuration(cmd)
			if err != nil {
				fmt.Println("taki error polecial: ", err)
			}
			fmt.Println("task ", task)
			fmt.Println("duration ", fmt.Sprintf("%dh %dm", int(duration.Hours()), int(duration.Minutes())%60))
			// if err := logTimeToJira(task, duration, comment, config); err != nil {
			// 	fmt.Println("Error logging time:", err)
			// } else {
			// 	fmt.Printf("Successfully logged %dh %dm for ticket %s\n", hours, minutes, task)

			// }
			return

		},
	}

	logCmd.Flags().IntP("hours", "H", 0, "Hours spent")
	logCmd.Flags().IntP("minutes", "m", 0, "Minutes spent")
	logCmd.Flags().StringP("comment", "c", "", "Worklog comment")
	logCmd.Flags().StringP("task", "t", "", "Jira task ID")
	logCmd.Flags().StringP("alias", "a", "", "Task by alias")

	configCmd.AddCommand(setHostCmd, setTokenCmd)
	aliasCmd.AddCommand(setAliasCmd, listAliasesCmd, removeAliasCmd)
	rootCmd.AddCommand(configCmd, logCmd, startTimerCmd, aliasCmd)
	prompter := prompter.NewBasicPrompter()
	config := configuration.NewBasicConfigurationHandler()
	ctx := context.WithValue(context.Background(), prompterKey, prompter)
	ctx = context.WithValue(ctx, configKey, config)
	rootCmd.SetContext(ctx)
	rootCmd.Execute()
}
