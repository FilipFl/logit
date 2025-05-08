package commands

import (
	"fmt"

	"github.com/FilipFl/logit/internal/configuration"
	"github.com/FilipFl/logit/internal/prompter"
	"github.com/spf13/cobra"
)

func NewSetAliasCommand(config configuration.Config, prompter prompter.Prompter) *cobra.Command {
	return &cobra.Command{
		Use:   "set [alias] [taskKey]",
		Short: "Set an alias for a Jira task",
		Args:  cobra.ExactArgs(2),
		Run: func(cmd *cobra.Command, args []string) {
			taskKey, err := extractJiraTaskKey(args[1])
			if err != nil {
				fmt.Println(err)
				return
			}
			oldTaskKey, _ := config.GetTaskFromAlias(args[0])
			if oldTaskKey != "" {
				approve, err := prompter.PromptForApprove(fmt.Sprintf("Are You sure You want to overwrite alias %s: %s with task %s", args[0], oldTaskKey, taskKey))
				if err != nil {
					fmt.Println("Error setting an alias:", err)
					return
				}
				if !approve {
					return
				}
			}
			err = config.AddAlias(args[0], taskKey)
			if err != nil {
				fmt.Println("Failed setting alias:", err)
				return
			}
			fmt.Printf("Alias %s set for task %s\n", args[0], taskKey)
		},
	}
}

func NewRemoveAliasCommand(config configuration.Config) *cobra.Command {
	return &cobra.Command{
		Use:   "remove [alias]",
		Short: "Remove an alias from aliases list",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			err := config.RemoveAlias(args[0])
			if err != nil {
				fmt.Printf("Alias %s not found on aliases list\n", args[0])
				return
			}
			fmt.Printf("Removed alias %s from aliases list\n", args[0])
		},
	}
}

func NewListAliasesCommand(config configuration.Config) *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "Lists all aliases",
		Args:  nil,
		Run: func(cmd *cobra.Command, args []string) {
			for k, v := range config.GetAliases() {
				fmt.Printf("%s: %s \n", k, v)
			}
		},
	}
}
