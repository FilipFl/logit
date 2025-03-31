package main

import (
	"github.com/FilipFl/logit/configuration"
	"github.com/FilipFl/logit/git"
	"github.com/FilipFl/logit/prompter"
	"github.com/spf13/cobra"
)

func main() {
	prompter := prompter.NewBasicPrompter()
	config := configuration.NewBasicConfigurationHandler()
	gitHandler := git.NewBasicGitHandler()

	var rootCmd = &cobra.Command{Use: "logit"}

	var configCmd = &cobra.Command{
		Use:   "config",
		Short: "Set Jira configuration",
	}

	var aliasCmd = &cobra.Command{
		Use:   "alias",
		Short: "Manage aliases",
	}

	setHostCmd := configuration.NewSetHostCommand(config)
	setTokenCmd := configuration.NewSetTokenCommand(config)
	setTokenEnvNameCmd := configuration.NewSetTokenEnvNameCommand(config)
	setEmailCmd := configuration.NewSetEmailCommand(config)

	setAliasCmd := NewSetAliasCommand(config)
	removeAliasCmd := NewRemoveAliasCommand(config)
	listAliasesCmd := NewListAliasesCommand(config)

	startTimerCmd := NewStartTimerCommand(config)

	logCmd := NewLogCommand(config, prompter, gitHandler)

	configCmd.AddCommand(setHostCmd, setTokenCmd, setEmailCmd, setTokenEnvNameCmd)

	aliasCmd.AddCommand(setAliasCmd, listAliasesCmd, removeAliasCmd)

	rootCmd.AddCommand(configCmd, logCmd, startTimerCmd, aliasCmd)

	rootCmd.Execute()
}
