package main

import (
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

	setHostCmd := configuration.NewSetHostCommand(config)
	setTokenCmd := configuration.NewSetTokenCommand(config)
	setTokenEnvNameCmd := configuration.NewSetTokenEnvNameCommand(config)
	setEmailCmd := configuration.NewSetEmailCommand(config)

	setAliasCmd := NewSetAliasCommand(config)
	removeAliasCmd := NewRemoveAliasCommand(config)
	listAliasesCmd := NewListAliasesCommand(config)

	startTimerCmd := NewStartTimerCommand(config, timer)

	myTasksCmd := NewMyTasksCommand(jiraClient)
	logCmd := NewLogCommand(config, prompter, gitHandler, timer, jiraClient)

	configCmd.AddCommand(setHostCmd, setTokenCmd, setEmailCmd, setTokenEnvNameCmd)

	aliasCmd.AddCommand(setAliasCmd, listAliasesCmd, removeAliasCmd)

	rootCmd.AddCommand(configCmd, logCmd, startTimerCmd, aliasCmd, myTasksCmd)

	rootCmd.Execute()
}
