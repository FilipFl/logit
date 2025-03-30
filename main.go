package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

const configPath = "~/.logit/config.json"

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

func getGitBranch() (string, error) {
	cmd := exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD")
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(output)), nil
}

func extractJiraTicket(branch string) (string, error) {
	re := regexp.MustCompile(`([A-Z]+-\d+)`)
	matches := re.FindStringSubmatch(branch)
	if len(matches) > 1 {
		return matches[1], nil
	}
	return "", fmt.Errorf("no Jira ticket found in branch")
}

func parseDuration(hours, minutes int) time.Duration {
	return time.Duration(hours)*time.Hour + time.Duration(minutes)*time.Minute
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
		body, _ := ioutil.ReadAll(resp.Body)
		return fmt.Errorf("failed to log time: %s", string(body))
	}

	return nil
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

	var logCmd = &cobra.Command{
		Use:   "log",
		Short: "Log time to Jira",
		Run: func(cmd *cobra.Command, args []string) {
			config, _ := loadConfig()
			hours, _ := cmd.Flags().GetInt("hours")
			minutes, _ := cmd.Flags().GetInt("minutes")
			comment, _ := cmd.Flags().GetString("comment")
			ticket, _ := cmd.Flags().GetString("ticket")
			alias, _ := cmd.Flags().GetString("alias")

			if alias, exists := config.Aliases[alias]; exists {
				ticket = alias
			}

			duration := parseDuration(hours, minutes)
			if err := logTimeToJira(ticket, duration, comment, config); err != nil {
				fmt.Println("Error logging time:", err)
			} else {
				fmt.Printf("Successfully logged %dh %dm for ticket %s\n", hours, minutes, ticket)
			}
		},
	}

	logCmd.Flags().IntP("hours", "h", 0, "Hours spent")
	logCmd.Flags().IntP("minutes", "m", 0, "Minutes spent")
	logCmd.Flags().StringP("comment", "c", "", "Worklog comment")
	logCmd.Flags().StringP("ticket", "t", "", "Jira ticket ID")
	logCmd.Flags().StringP("alias", "a", "", "Task by alias")

	configCmd.AddCommand(setHostCmd, setTokenCmd, setAliasCmd)
	rootCmd.AddCommand(configCmd, logCmd)
	rootCmd.Execute()
}
