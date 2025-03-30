package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

const configPath = "~/.logit/config.json"

var errorNoJiraTicket = errors.New("no Jira ticket found in passed string")
var errorNoJiraTicketInFlagValue = errors.New("no Jira ticket found in passed value passed with task flag")
var errorScanningUserInput = errors.New("error scanning user input")
var errorWrongApproveInput = errors.New("user input wrong approve command")
var errorNoTargetToLogWork = errors.New("no target to log work")
var errorNoSnapshot = errors.New("no start time saved")
var errorWrongDuration = errors.New("duration to log is invalid")

type Config struct {
	JiraHost  string            `json:"jira_host"`
	JiraToken string            `json:"jira_token"`
	Aliases   map[string]string `json:"aliases"`
	Snapshot  *time.Time        `json:"snapshot"`
}

type Worklog struct {
	TimeSpent string `json:"timeSpent"`
	Started   string `json:"started"`
	Comment   string `json:"comment,omitempty"`
}

func loadConfig() (*Config, error) {
	file, err := os.Open(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			return &Config{Aliases: make(map[string]string)}, nil
		}
		return nil, err
	}
	defer file.Close()

	config := &Config{}
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(config); err != nil {
		return nil, err
	}
	return config, nil
}

func saveConfig(config *Config) error {
	file, err := os.Create(configPath)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	return encoder.Encode(config)
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

func parseDuration(config Config, cmd *cobra.Command) (time.Duration, error) {
	result := time.Duration(0)
	hours, _ := cmd.Flags().GetInt("hours")
	minutes, _ := cmd.Flags().GetInt("minutes")
	if hours != 0 || minutes != 0 {
		result = time.Duration(hours)*time.Hour + time.Duration(minutes)*time.Minute
	} else {
		if config.Snapshot == nil {
			return time.Duration(0), errorNoSnapshot
		}
		now := time.Now()
		result = now.Sub(*config.Snapshot)
	}
	if int(result.Hours()) > 8 {
		proceed, err := promptUserForApprove(fmt.Sprintf("Are You sure you want to log %d hours and %d minutes?", int(result.Hours()), int(result.Minutes())%60))
		if err != nil {
			return time.Duration(0), err
		}
		if proceed {
			return result, nil
		}
	}

	return time.Duration(0), errorWrongDuration
}

func logTimeToJira(ticket string, duration time.Duration, comment string, config *Config) error {
	timeSpent := fmt.Sprintf("%dh %dm", int(duration.Hours()), int(duration.Minutes())%60)
	url := fmt.Sprintf("%s/rest/api/2/issue/%s/worklog", config.JiraHost, ticket)
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
	req.SetBasicAuth("", config.JiraToken)
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

func promptUserForTask(msg string) (string, error) {
	fmt.Println(msg)
	fmt.Print("Provide task ID or task URL:")
	promptedTask := ""
	fmt.Scanln(&promptedTask)
	if promptedTask != "" {
		return extractJiraTicket(promptedTask)
	}
	return "", errorScanningUserInput
}

func promptUserForApprove(msg string) (bool, error) {
	fmt.Println(msg)
	fmt.Print("Proceed? y/n (and hit enter)")
	promptedApprove := ""
	fmt.Scanln(&promptedApprove)
	if promptedApprove != "" {
		switch promptedApprove {
		case "y":
			return true, nil
		case "Y":
			return true, nil
		case "n":
			return false, nil
		case "N":
			return false, nil
		default:
			return false, errorWrongApproveInput
		}
	}
	return false, errorWrongApproveInput

}

func determineTask(config Config, cmd *cobra.Command) (string, error) {
	resultTask := ""
	task, _ := cmd.Flags().GetString("task")
	alias, _ := cmd.Flags().GetString("alias")
	if resultTask, exists := config.Aliases[alias]; exists {
		return resultTask, nil
	}
	if task != "" {
		resultTask, err := extractJiraTicket(task)
		if err != nil {
			return "", errorNoJiraTicketInFlagValue
		}
		if resultTask != "" {
			return resultTask, nil
		}
		userPromptedTask, err := promptUserForTask("There is no jira task in value passed to task flag.")
		return userPromptedTask, err

	}
	gitBranch, err := getGitBranch()
	if err != nil {
		userPromptedTask, err := promptUserForTask("Current directory is not a git repository or something failed during branch name extraction.")
		return userPromptedTask, err
	}

	resultTask, err = extractJiraTicket(gitBranch)
	if err != nil {
		userPromptedTask, err := promptUserForTask("Current branch name does not contain task ID.")
		return userPromptedTask, err
	}

	proceed, err := promptUserForApprove(fmt.Sprintf("Detected task ID %s in current branch name.", resultTask))

	if err != nil {
		userPromptedTask, err := promptUserForTask("Error scanning proceed approve.")
		return userPromptedTask, err
	}

	if proceed {
		return resultTask, nil
	}

	return "", errorNoTargetToLogWork
}

func main() {
	var rootCmd = &cobra.Command{Use: "logit"}

	var configCmd = &cobra.Command{
		Use:   "config",
		Short: "Set Jira configuration",
	}

	var setHostCmd = &cobra.Command{
		Use:   "set-host [host]",
		Short: "Set Jira host",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			config, _ := loadConfig()
			config.JiraHost = args[0]
			saveConfig(config)
			fmt.Println("Jira host updated.")
		},
	}

	var setTokenCmd = &cobra.Command{
		Use:   "set-token [token]",
		Short: "Set Jira token",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			config, _ := loadConfig()
			config.JiraToken = args[0]
			saveConfig(config)
			fmt.Println("Jira token updated.")
		},
	}

	var setAliasCmd = &cobra.Command{
		Use:   "set-alias [alias] [ticket]",
		Short: "Set an alias for a Jira ticket",
		Args:  cobra.ExactArgs(2),
		Run: func(cmd *cobra.Command, args []string) {
			config, _ := loadConfig()
			config.Aliases[args[0]] = args[1]
			saveConfig(config)

			fmt.Printf("Alias %s set for ticket %s\n", args[0], args[1])
		},
	}

	var startTimerCmd = &cobra.Command{
		Use:   "start",
		Short: "Start measuring time from this moment",
		Args:  nil,
		Run: func(cmd *cobra.Command, args []string) {
			config, _ := loadConfig()
			now := time.Now()
			config.Snapshot = &now
			saveConfig(config)
			fmt.Println("Started to measure time.")
		},
	}

	var logCmd = &cobra.Command{
		Use:   "log",
		Short: "Log time to Jira",
		Run: func(cmd *cobra.Command, args []string) {
			config, _ := loadConfig()

			comment, _ := cmd.Flags().GetString("comment")

			task, err := determineTask(*config, cmd)
			if err != nil {
				fmt.Println("Error logging time: ", err)
			}
			if task == "" {
				fmt.Println("No target for time logging.")
				return
			}
			duration := parseDuration(*config, cmd)
			fmt.Println(task)
			return

			if err := logTimeToJira(task, duration, comment, config); err != nil {
				fmt.Println("Error logging time:", err)
			} else {
				fmt.Printf("Successfully logged %dh %dm for ticket %s\n", hours, minutes, task)
			}
		},
	}

	logCmd.Flags().IntP("hours", "H", 0, "Hours spent")
	logCmd.Flags().IntP("minutes", "m", 0, "Minutes spent")
	logCmd.Flags().StringP("comment", "c", "", "Worklog comment")
	logCmd.Flags().StringP("task", "t", "", "Jira task ID")
	logCmd.Flags().StringP("alias", "a", "", "Task by alias")

	configCmd.AddCommand(setHostCmd, setTokenCmd, setAliasCmd)
	rootCmd.AddCommand(configCmd, logCmd, startTimerCmd)
	rootCmd.Execute()
}
