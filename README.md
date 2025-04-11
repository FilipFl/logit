# Logit

A simple CLI tool to log work time into Jira straight from your terminal. Allows to seemlesly log time into tasks You're working on without interrupting Your workflow.

* integrated with Git - if Your current branch contains task key You don't need to pass task explicitly
* capable of measuring time - just use `logit start` when starting Your workday
* allows to set aliases for the tasks You frequently log time on
* provides one command to fetch all tasks You're assigned to

---

## Installation

    go build -o bin/logit
    sudo cp bin/logit /usr/local/bin/logit

Or using Make:

    make install

---

## Usage

    logit [command] [flags]

### Example

    logit alias set daily JIRA-111

    logit log --ticket JIRA-123 --H 2 --comment "Worked on backend bug" // will log 2 hours for task JIRA-123 for today

    logit log --a daily --m 20  // will log 20 minutes for task aliased by "daily"

    logit log --H 5 --y // will log 5 hours, for the task mentioned in current git branch, for the yesterday

    logit log --ticket JIRA-321 --f // will attempt to log time passed from latest snapshot for today
---

## Available Commands

### Root Level

| Command | Description                       |
| ------- | --------------------------------- |
| log     | Log time to a Jira ticket         |
| config  | Set of configuration commands     |
| alias   | Set of alias commands             |
| start   | Start time measure in this moment |
| tasks   | List tasks assigned to You        |
| help    | Show help for any command         |

### Config Level

| Command                          | Description                                                        |
| -------------------------------- | ------------------------------------------------------------------ |
| config set-origin [origin]       | Set Jira origin (schema + host)                                    |
| config set-token  [token]        | Set personal Jira token                                            |
| config set-token-env-name [name] | Set name of environmental variable where logit can find jira token |
| config set-email [email]         | Set Your Jira email                                                |
| config help                      | Show help for any command                                          |

### Alias Level

| Command                  | Description               |
| ------------------------ | ------------------------- |
| alias set [alias] [task] | Save new task alias       |
| alias remove [alias]     | Remove saved alias        |
| alias list               | List saved aliasses       |
| alias help               | Show help for any command |

---

## Log Flags

| Flag        | Flag shorthand | Description                                                                                               | Example                     |
| ----------- | -------------- | --------------------------------------------------------------------------------------------------------- | --------------------------- |
| --ticket    | --t            | Jira ticket key (if ommitted with alias git branch is inspected)                                          | --ticket JIRA-123           |
| --alias     | --a            | Jira ticket alias (if ommitted with ticket git branch is inspected)                                       | --alias myTask              |
| --hours     | --H            | Duration to log in hours (e.g. 1h) (if both hours and minutes flags are ommited time snapshot is used)    | --hours 1                   |
| --minutes   | --m            | Duration to log in minutes (e.g. 30m) (if both hours and minutes flags are ommited time snapshot is used) | --minutes 30                |
| --comment   | --c            | Worklog comment                                                                                           | --comment "Fixed login bug" |
| --yesterday | --y            | Log work for yesterday                                                                                    | --yesterday                 |
| --date      | --d            | Log work for date (dd.mm or dd-mm format required), current year is assumed                               | --date 12.03                |
| --reset     | --r            | If used with hours or minutes flags forces to reset snapshot on time log                                  | --reset                     |
| --force     | --f            | Forces all boolean prompts to pass                                                                        | --f                         |

---

## Configuration

Logit uses a configuration file to preserve all needed data.

It automatically creates a file at `~/.logit/config.yml` on first use.

Feel free to edit it's safe to use config commands but feel free to edit it by hand.

---

## Running Tests

    make test

---

## Contributing

Please open an issue if You'd like to see something added or fixed.

---

## License

MIT
