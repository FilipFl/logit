package configuration

import (
	"fmt"

	"github.com/spf13/cobra"
)

func NewSetOriginCommand(cfgHandler ConfigurationHandler) *cobra.Command {
	return &cobra.Command{
		Use:   "set-origin [origin]",
		Short: "Set Jira origin (consists of schema + host)",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			cfg := cfgHandler.LoadConfig()
			cfg.JiraOrigin = args[0]
			cfgHandler.SaveConfig(cfg)
			fmt.Println("Jira origin updated.")
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
