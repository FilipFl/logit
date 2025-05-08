package commands

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"runtime"

	"text/tabwriter"

	"github.com/FilipFl/logit/internal/configuration"
	"github.com/FilipFl/logit/internal/git"
	"github.com/FilipFl/logit/internal/jira"
	"github.com/FilipFl/logit/internal/printer"
	"github.com/FilipFl/logit/internal/prompter"
	"github.com/FilipFl/logit/internal/timer"
	"github.com/spf13/cobra"
)

func NewStartTimerCommand(cfg configuration.Config, timer timer.Timer) *cobra.Command {
	return &cobra.Command{
		Use:   "start",
		Short: "Start measuring time from this moment",
		Args:  nil,
		Run: func(cmd *cobra.Command, args []string) {
			now := timer.Now()
			err := cfg.SetSnapshot(&now)
			if err != nil {
				fmt.Println("Failed starting to measure time:", err)
				return
			}
			fmt.Println("Started to measure time.")
		},
	}
}

func NewOpenCommand(cfg configuration.Config, prompter prompter.Prompter, gitHandler git.GitHandler) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "open [alias | taskKey]",
		Short: "Open task in browser",
		Args:  cobra.MaximumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			var task string
			var err error
			if len(args) > 0 {
				task, err = cfg.GetTaskFromAlias(args[0])
				if err != nil {
					task, _ = extractJiraTaskKey(args[0])
				}
			}
			if task == "" {
				task, err = determineTask(cmd, cfg, prompter, gitHandler, true)
				if err != nil {
					fmt.Println("Failed to open task:", err)
					return
				}
			}
			if cfg.GetJiraOrigin() == "" {
				fmt.Println("Before trying to open browser configure Jira origin")
				return
			}
			url := cfg.GetJiraOrigin() + "/browse/" + task
			var comm *exec.Cmd
			if runtime.GOOS == "darwin" {
				comm = exec.Command("open", url)
			} else {
				comm = exec.Command("xdg-open", url)
			}
			comm.Stderr = io.Discard
			comm.Stdout = io.Discard
			comm.Stdin = nil
			err = comm.Start()
			if err != nil {
				fmt.Println("Error opening browser:", err)
			}
		},
	}
	cmd.Flags().StringP("task", "t", "", "Jira task ID or URL")
	cmd.Flags().StringP("alias", "a", "", "Task by alias")
	cmd.RegisterFlagCompletionFunc("alias", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		aliases := []string{}
		for alias := range cfg.GetAliases() {
			aliases = append(aliases, alias)
		}
		return aliases, cobra.ShellCompDirectiveNoFileComp
	})
	return cmd
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
			w := tabwriter.NewWriter(os.Stdout, 0, 0, 1, ' ', tabwriter.Debug)
			for _, issue := range results {
				fmt.Fprintf(w, "%s\t%s\t%s\t\n", issue.Key, truncateString(issue.Summary, 37), issue.Status)
			}
			w.Flush()
		},
	}
}

func NewMyWorklogsCommand(client jira.Client) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "worklogs",
		Short: "List my worklogs",
		Args:  nil,
		Run: func(cmd *cobra.Command, args []string) {
			err := assertWorklogsFlagsAreValid(cmd)
			if err != nil {
				fmt.Println("Passed flags are invalid:", err)
				return
			}
			fromDays := worklogsFromHowManyDays(cmd)
			results, err := client.GetLoggedTime(fromDays)
			if err != nil {
				fmt.Println("Error fetching assigned tasks:", err)
				return
			}
			if len(results.Days) == 0 {
				fmt.Println("no time logged in specified time range")
				return
			}
			for _, day := range results.Days {
				w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', tabwriter.Debug)
				for _, log := range day.Worklogs {
					fmt.Fprintf(w, "%s\t%s\t%s\t\n", log.TaskKey, log.StringLoggedTime(), truncateString(log.Summary, 40))
				}
				printer.PrintGreen(fmt.Sprintf("%s (%s) - %dh %dm\n", day.DateString(), day.Date.Weekday(), int(day.TimeLogged.Hours()), int(day.TimeLogged.Minutes())%60))
				w.Flush()
			}
		},
	}
	cmd.Flags().BoolP("today", "t", false, "Return worklogs from today")
	cmd.Flags().BoolP("yesterday", "y", false, "Return worklogs from yesterday and today")
	cmd.Flags().BoolP("week", "w", false, "Return worklogs from last week")
	cmd.Flags().IntP("days", "d", 0, "Return worklogs from X days (X must be less or equal than 14)")
	return cmd
}

func NewLogCommand(cfg configuration.Config, prompter prompter.Prompter, gitHandler git.GitHandler, timer timer.Timer, client jira.Client) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "log",
		Short: "Log time to Jira",
		Run: func(cmd *cobra.Command, args []string) {
			err := assertFlagsAreValid(cmd, timer)
			if err != nil {
				fmt.Println("Error validating flags:", err)
				return
			}

			force, _ := cmd.Flags().GetBool("force")
			task, err := determineTask(cmd, cfg, prompter, gitHandler, force)
			if err != nil {
				fmt.Println("Error assessing task to log time:", err)
				return
			}
			if task == "" {
				fmt.Println("No target indicated for time logging.")
				return
			}
			duration, fromSnapshot, err := parseDuration(cmd, cfg, prompter, timer)
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
					now := timer.Now()
					err := cfg.SetSnapshot(&now)
					if err != nil {
						fmt.Println("Failed starting to measure time:", err)
						return
					}
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
	cmd.RegisterFlagCompletionFunc("alias", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		aliases := []string{}
		for alias := range cfg.GetAliases() {
			aliases = append(aliases, alias)
		}
		return aliases, cobra.ShellCompDirectiveNoFileComp
	})
	return cmd
}
