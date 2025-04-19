package configuration

import (
	"fmt"

	"github.com/FilipFl/logit/internal/prompter"
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
		Short: "Set Your jira email",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			cfg := cfgHandler.LoadConfig()
			cfg.JiraEmail = args[0]
			cfgHandler.SaveConfig(cfg)
			fmt.Println("Jira email updated.")
		},
	}
}

func NewInitCommand(cfgHandler ConfigurationHandler, prompter prompter.Prompter) *cobra.Command {
	return &cobra.Command{
		Use:   "init",
		Short: "Initialize config. Logit will prompt for Jira origin, and Your Jira PAT Access",
		Args:  cobra.ExactArgs(0),
		Run: func(cmd *cobra.Command, args []string) {
			cfg := cfgHandler.LoadConfig()
			origin, err := prompter.PromptForString("", "Please enter Jira origin (schema + host): ")
			if err != nil {
				fmt.Println("operation aborted ", err)
				return
			}
			cfg.JiraOrigin = origin
			email, err := prompter.PromptForString("", "Please enter Jira email: ")
			if err != nil {
				fmt.Println("operation aborted ", err)
				return
			}
			cfg.JiraEmail = email
			directToken, err := prompter.PromptForApprove("Do You want to provide Personal Access Token directly to be stored in config file?")
			if err != nil {
				fmt.Println("operation aborted ", err)
			}
			if directToken {
				token, err := prompter.PromptForString("", "Please enter Your Personal Access Token: ")
				if err != nil {
					fmt.Println("operation aborted ", err)
					return
				}
				cfg.JiraToken = token
			} else {
				tokenEnvName, err := prompter.PromptForString("", "Please enter token environmental variable name: ")
				if err != nil {
					fmt.Println("operation aborted ", err)
					return
				}
				cfg.JiraTokenEnvName = tokenEnvName
			}
			cfgHandler.SaveConfig(cfg)
			fmt.Println("Jira config saved")
		},
	}
}

func NewSwitchTrustGitBranchCommand(cfgHandler ConfigurationHandler) *cobra.Command {
	return &cobra.Command{
		Use:   "trustGitBranch",
		Short: "Switch value of trustGitBranch variable - if true logit will not prompt for confirmation of a task extracted from git branch",
		Args:  cobra.ExactArgs(0),
		Run: func(cmd *cobra.Command, args []string) {
			cfg := cfgHandler.LoadConfig()
			cfg.TrustGitBranch = !cfg.TrustGitBranch
			cfgHandler.SaveConfig(cfg)
			if cfg.TrustGitBranch {
				fmt.Println("logit will automatically extract task from git branch and wont ask for confirmation")
			} else {
				fmt.Println("logit will prompt for confirmation of a task extracted from git branch")
			}
		},
	}
}
