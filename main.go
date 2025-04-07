package main

import (
	"github.com/FilipFl/logit/commands"
	"github.com/FilipFl/logit/configuration"
	"github.com/FilipFl/logit/git"
	"github.com/FilipFl/logit/jira"
	"github.com/FilipFl/logit/prompter"
	"github.com/FilipFl/logit/timer"
	"github.com/spf13/cobra"
)

func main() {
	prompter := prompter.NewBasicPrompter()
	config := configuration.NewBasicConfigurationHandler()
	gitHandler := git.NewBasicGitHandler()
	timer := timer.NewBasicTimer()
	jiraClient := jira.NewJiraClient(config)

	var rootCmd = &cobra.Command{Use: "logit"}

	var configCmd = &cobra.Command{
		Use:   "config",
		Short: "Set Jira configuration",
	}

	var aliasCmd = &cobra.Command{
		Use:   "alias",
		Short: "Manage aliases",
	}

	setHostCmd := configuration.NewSetOriginCommand(config)
	setTokenCmd := configuration.NewSetTokenCommand(config)
	setTokenEnvNameCmd := configuration.NewSetTokenEnvNameCommand(config)
	setEmailCmd := configuration.NewSetEmailCommand(config)

	setAliasCmd := commands.NewSetAliasCommand(config)
	removeAliasCmd := commands.NewRemoveAliasCommand(config)
	listAliasesCmd := commands.NewListAliasesCommand(config)

	startTimerCmd := commands.NewStartTimerCommand(config, timer)

	myTasksCmd := commands.NewMyTasksCommand(jiraClient)
	logCmd := commands.NewLogCommand(config, prompter, gitHandler, timer, jiraClient)

	configCmd.AddCommand(setHostCmd, setTokenCmd, setEmailCmd, setTokenEnvNameCmd)

	aliasCmd.AddCommand(setAliasCmd, listAliasesCmd, removeAliasCmd)

	rootCmd.AddCommand(configCmd, logCmd, startTimerCmd, aliasCmd, myTasksCmd)

	rootCmd.Execute()
}
