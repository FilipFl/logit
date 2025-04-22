package configuration

import (
	"fmt"

	"github.com/FilipFl/logit/internal/prompter"
	"github.com/spf13/cobra"
)

func NewSetOriginCommand(config Config) *cobra.Command {
	return &cobra.Command{
		Use:   "set-origin [origin]",
		Short: "Set Jira origin (consists of schema + host)",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			err := config.SetJiraOrigin(args[0])
			if err != nil {
				fmt.Println("Failed setting Jira Origin:", err)
				return
			}
			fmt.Println("Jira origin updated.")
		},
	}
}

func NewSetTokenCommand(config Config) *cobra.Command {
	return &cobra.Command{
		Use:   "set-token [token]",
		Short: "Set Jira token",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			err := config.SetJiraToken(args[0])
			if err != nil {
				fmt.Println("Failed setting token:", err)
				return
			}
			fmt.Println("Jira token updated.")
		},
	}
}

func NewSetTokenEnvNameCommand(config Config) *cobra.Command {
	return &cobra.Command{
		Use:   "set-token-env-name [name]",
		Short: "Set name of environmental variable which holds Jira token",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			err := config.SetJiraTokenEnvName(args[0])
			if err != nil {
				fmt.Println("Failed setting name of environmental variable holding Jira token:", err)
				return
			}
			fmt.Println("Name of environmental variable holding Jira token updated.")
		},
	}
}

func NewSetEmailCommand(config Config) *cobra.Command {
	return &cobra.Command{
		Use:   "set-email [email]",
		Short: "Set Your jira email",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			err := config.SetJiraEmail(args[0])
			if err != nil {
				fmt.Println("Failed setting email:", err)
				return
			}
			fmt.Println("Jira email updated.")
		},
	}
}

func NewInitCommand(config Config, prompter prompter.Prompter) *cobra.Command {
	return &cobra.Command{
		Use:   "init",
		Short: "Initialize config. Logit will prompt for Jira origin, and Your Jira PAT Access",
		Args:  cobra.ExactArgs(0),
		Run: func(cmd *cobra.Command, args []string) {
			origin, err := prompter.PromptForString("", "Please enter Jira origin (schema + host): ")
			if err != nil {
				fmt.Println("operation aborted ", err)
				return
			}
			err = config.SetJiraOrigin(origin)
			if err != nil {
				fmt.Println("Failed setting origin:", err)
				return
			}
			email, err := prompter.PromptForString("", "Please enter Jira email: ")
			if err != nil {
				fmt.Println("operation aborted ", err)
				return
			}
			err = config.SetJiraEmail(email)
			if err != nil {
				fmt.Println("Failed setting email:", err)
				return
			}
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
				err = config.SetJiraToken(token)
				if err != nil {
					fmt.Println("Failed setting token:", err)
					return
				}
			} else {
				tokenEnvName, err := prompter.PromptForString("", "Please enter token environmental variable name: ")
				if err != nil {
					fmt.Println("operation aborted ", err)
					return
				}
				err = config.SetJiraTokenEnvName(tokenEnvName)
				if err != nil {
					fmt.Println("Failed setting name of environmental variable holding Jira token:", err)
					return
				}
			}
			fmt.Println("Jira config saved")
		},
	}
}

func NewSwitchTrustGitBranchCommand(config Config) *cobra.Command {
	return &cobra.Command{
		Use:   "trustGitBranch",
		Short: "Switch value of trustGitBranch variable - if true logit will not prompt for confirmation of a task extracted from git branch",
		Args:  cobra.ExactArgs(0),
		Run: func(cmd *cobra.Command, args []string) {
			err := config.SwapTrustGitBranch()
			if err != nil {
				fmt.Println("Error setting new value of TrustGitBranch variable:", err)
			}
			if config.GetTrustGitBranch() {
				fmt.Println("logit will automatically extract task from git branch and wont ask for confirmation")
			} else {
				fmt.Println("logit will prompt for confirmation of a task extracted from git branch")
			}
		},
	}
}
