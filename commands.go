package main

import (
	"fmt"

	"github.com/FilipFl/logit/configuration"
	"github.com/FilipFl/logit/git"
	"github.com/FilipFl/logit/prompter"
	"github.com/FilipFl/logit/timer"
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

func NewLogCommand(cfgHandler configuration.ConfigurationHandler, prompter prompter.Prompter, gitHandler git.GitHandler, timer timer.Timer) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "log",
		Short: "Log time to Jira",
		Run: func(cmd *cobra.Command, args []string) {
			// comment, _ := cmd.Flags().GetString("comment")

			task, err := determineTask(cmd, cfgHandler, prompter, gitHandler)
			if err != nil {
				fmt.Println("Error logging time: ", err)
				return
			}
			if task == "" {
				fmt.Println("No target for time logging.")
				return
			}
			duration, err := parseDuration(cmd, cfgHandler, prompter, timer)
			if err != nil {
				fmt.Println("Invalid log work duration: ", err)
				return
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
	cmd.Flags().IntP("hours", "H", 0, "Hours spent")
	cmd.Flags().IntP("minutes", "m", 0, "Minutes spent")
	cmd.Flags().StringP("comment", "c", "", "Worklog comment")
	cmd.Flags().StringP("task", "t", "", "Jira task ID")
	cmd.Flags().StringP("alias", "a", "", "Task by alias")
	return cmd
}
