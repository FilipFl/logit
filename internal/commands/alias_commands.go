package commands

import (
	"fmt"

	"github.com/FilipFl/logit/internal/configuration"
	"github.com/FilipFl/logit/internal/prompter"
	"github.com/spf13/cobra"
)

func NewSetAliasCommand(cfgHandler configuration.ConfigurationHandler, prompter prompter.Prompter) *cobra.Command {
	return &cobra.Command{
		Use:   "set [alias] [ticket]",
		Short: "Set an alias for a Jira ticket",
		Args:  cobra.ExactArgs(2),
		Run: func(cmd *cobra.Command, args []string) {
			cfg := cfgHandler.LoadConfig()
			ticket, err := extractJiraTicket(args[1])
			if err != nil {
				fmt.Println(err)
				return
			}
			if val, exists := cfg.Aliases[args[0]]; exists {
				approve, err := prompter.PromptForApprove(fmt.Sprintf("Are You sure You want to overwrite alias %s: %s with task %s", args[0], val, ticket))
				if err != nil {
					fmt.Println("Error setting an alias:", err)
					return
				}
				if !approve {
					return
				}
			}
			cfg.Aliases[args[0]] = ticket
			cfgHandler.SaveConfig(cfg)

			fmt.Printf("Alias %s set for ticket %s\n", args[0], args[1])
		},
	}
}

func NewRemoveAliasCommand(cfgHandler configuration.ConfigurationHandler) *cobra.Command {
	return &cobra.Command{
		Use:   "remove [alias]",
		Short: "Remove an alias from aliases list",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			cfg := cfgHandler.LoadConfig()
			if _, exists := cfg.Aliases[args[0]]; exists {
				delete(cfg.Aliases, args[0])
				cfgHandler.SaveConfig(cfg)
				fmt.Printf("Removed alias %s from aliases list\n", args[0])
				return
			}

			fmt.Printf("Alias %s not found on aliases list\n", args[0])
		},
	}
}

func NewListAliasesCommand(cfgHandler configuration.ConfigurationHandler) *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "Lists all aliases",
		Args:  nil,
		Run: func(cmd *cobra.Command, args []string) {
			cfg := cfgHandler.LoadConfig()
			for k, v := range cfg.Aliases {
				fmt.Printf("%s: %s \n", k, v)
			}
		},
	}
}
