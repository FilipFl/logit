package configuration

import (
	"fmt"

	"github.com/spf13/cobra"
)

func NewSetHostCommand(cfgHandler ConfigurationHandler) *cobra.Command {
	return &cobra.Command{
		Use:   "set-host [host]",
		Short: "Set Jira host",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			cfg := cfgHandler.LoadConfig()
			cfg.JiraHost = args[0]
			cfgHandler.SaveConfig(cfg)
			fmt.Println("Jira host updated.")
		},
	}
}

func NewSetTokenCommand(cfgHandler ConfigurationHandler) *cobra.Command {
	return &cobra.Command{
		Use:   "set-token [token]",
		Short: "Set Jira token",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			cfg := cfgHandler.LoadConfig()
			cfg.JiraToken = args[0]
			cfgHandler.SaveConfig(cfg)
			fmt.Println("Jira token updated.")
		},
	}
}

func NewSetTokenEnvNameCommand(cfgHandler ConfigurationHandler) *cobra.Command {
	return &cobra.Command{
		Use:   "set-token-env-name [name]",
		Short: "Set name of environmental variable which holds Jira token",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			cfg := cfgHandler.LoadConfig()
			cfg.JiraTokenEnvName = args[0]
			cfgHandler.SaveConfig(cfg)
			fmt.Println("Name of environmental variable holding Jira token updated.")
		},
	}
}

func NewSetEmailCommand(cfgHandler ConfigurationHandler) *cobra.Command {
	return &cobra.Command{
		Use:   "set-email [email]",
		Short: "Set Jira user email",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			cfg := cfgHandler.LoadConfig()
			cfg.JiraEmail = args[0]
			cfgHandler.SaveConfig(cfg)
			fmt.Println("Jira user email updated.")
		},
	}
}
