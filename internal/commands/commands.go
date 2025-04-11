package commands

import (
	"fmt"

	"github.com/FilipFl/logit/internal/configuration"
	"github.com/FilipFl/logit/internal/git"
	"github.com/FilipFl/logit/internal/jira"
	"github.com/FilipFl/logit/internal/prompter"
	"github.com/FilipFl/logit/internal/timer"
	"github.com/spf13/cobra"
)

func NewStartTimerCommand(cfgHandler configuration.ConfigurationHandler, timer timer.Timer) *cobra.Command {
	return &cobra.Command{
		Use:   "start",
		Short: "Start measuring time from this moment",
		Args:  nil,
		Run: func(cmd *cobra.Command, args []string) {
			cfg := cfgHandler.LoadConfig()
			now := timer.Now()
			cfg.Snapshot = &now
			cfgHandler.SaveConfig(cfg)
			fmt.Println("Started to measure time.")
		},
	}
}

func NewMyTasksCommand(client jira.Client) *cobra.Command {
	return &cobra.Command{
		Use:   "tasks",
		Short: "List tasks assigned to me",
		Args:  nil,
		Run: func(cmd *cobra.Command, args []string) {
			results, err := client.GetAssignedIssues()
			if err != nil {
				fmt.Println("Error fetching assigned tasks:", err)
				return
			}
			for _, issue := range results {
				fmt.Printf("Task: %s, Summary: %s, Status: %s\n", issue.Key, issue.Summary, issue.Status)
			}
		},
	}
}

func NewLogCommand(cfgHandler configuration.ConfigurationHandler, prompter prompter.Prompter, gitHandler git.GitHandler, timer timer.Timer, client jira.Client) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "log",
		Short: "Log time to Jira",
		Run: func(cmd *cobra.Command, args []string) {
			err := assertFlagsAreValid(cmd, timer)
			if err != nil {
				fmt.Println("Error validating flags:", err)
				return
			}

			task, err := determineTask(cmd, cfgHandler, prompter, gitHandler)
			if err != nil {
				fmt.Println("Error assessing task to log time:", err)
				return
			}
			if task == "" {
				fmt.Println("No target indicated for time logging.")
				return
			}
			duration, fromSnapshot, err := parseDuration(cmd, cfgHandler, prompter, timer)
			if err != nil {
				fmt.Println("Invalid log work duration:", err)
				return
			}
			dateStarted, err := determineStarted(cmd, timer)
			if err != nil {
				fmt.Println("Error assessing date to log time on:", err)
				return
			}

			comment, _ := cmd.Flags().GetString("comment")
			if err := client.LogTime(task, duration, dateStarted, comment); err != nil {
				fmt.Println("Error logging time:", err)
			} else {
				fmt.Printf("Successfully logged %dh %dm for task %s\n", int(duration.Hours()), int(duration.Minutes())%60, task)
				reset, _ := cmd.Flags().GetBool("reset")
				if fromSnapshot || reset {
					cfg := cfgHandler.LoadConfig()
					now := timer.Now()
					cfg.Snapshot = &now
					cfgHandler.SaveConfig(cfg)
				}
			}
		},
	}
	cmd.Flags().IntP("hours", "H", 0, "Hours spent")
	cmd.Flags().IntP("minutes", "m", 0, "Minutes spent")
	cmd.Flags().StringP("comment", "c", "", "Worklog comment")
	cmd.Flags().StringP("task", "t", "", "Jira task ID or URL")
	cmd.Flags().StringP("alias", "a", "", "Task by alias")
	cmd.Flags().BoolP("yesterday", "y", false, "Log time for yesterday")
	cmd.Flags().StringP("date", "d", "", "Date in format dd-mm, present year is assumed")
	cmd.Flags().BoolP("reset", "r", false, "Reset snapshot")
	cmd.Flags().BoolP("force", "f", false, "Force approve when prompted")
	return cmd
}
