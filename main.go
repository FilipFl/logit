package main

import (
	"github.com/FilipFl/logit/internal/commands"
	"github.com/FilipFl/logit/internal/configuration"
	"github.com/FilipFl/logit/internal/git"
	"github.com/FilipFl/logit/internal/jira"
	"github.com/FilipFl/logit/internal/prompter"
	"github.com/FilipFl/logit/internal/timer"
	"github.com/spf13/cobra"
)

func main() {
	prompter := prompter.NewBasicPrompter()
	config := configuration.NewBasicConfig()
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
	initCmd := configuration.NewInitCommand(config, prompter)
	trustGitBranchCmd := configuration.NewSwitchTrustGitBranchCommand(config)

	setAliasCmd := commands.NewSetAliasCommand(config, prompter)
	removeAliasCmd := commands.NewRemoveAliasCommand(config)
	listAliasesCmd := commands.NewListAliasesCommand(config)

	startTimerCmd := commands.NewStartTimerCommand(config, timer)
	openCmd := commands.NewOpenCommand(config, prompter, gitHandler)

	myTasksCmd := commands.NewMyTasksCommand(jiraClient)
	myWorklogsCmd := commands.NewMyWorklogsCommand(jiraClient)
	logCmd := commands.NewLogCommand(config, prompter, gitHandler, timer, jiraClient)

	configCmd.AddCommand(setHostCmd, setTokenCmd, setTokenEnvNameCmd, setEmailCmd, initCmd, trustGitBranchCmd)

	aliasCmd.AddCommand(setAliasCmd, listAliasesCmd, removeAliasCmd)

	rootCmd.AddCommand(configCmd, logCmd, startTimerCmd, aliasCmd, myTasksCmd, myWorklogsCmd, openCmd)

	rootCmd.Execute()
}
