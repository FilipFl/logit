package jira

import (
	"time"
)

type Client interface {
	LogTime(taskKey string, duration time.Duration, started time.Time, comment string) error
	GetAssignedIssues() ([]Issue, error)
	GetLoggedTime(fromDays int) (Logs, error)
}

type Result struct {
	Issues []JiraIssue `json:"issues"`
}

type Issue struct {
	Summary string `json:"summary"`
	Status  string `json:"status"`
	Key     string `json:"key"`
}

type JiraIssue struct {
	Key    string          `json:"key"`
	Fields JiraIssueFields `json:"fields"`
}

type JiraStatus struct {
	Name string `json:"name"`
}

type JiraIssueFields struct {
	Worklog JiraWorklogs `json:"worklog"`
	Summary string       `json:"summary"`
	Status  JiraStatus   `json:"status"`
}

type JiraWorklogs struct {
	Worklogs []JiraIssueWorklog `json:"worklogs"`
}

type JiraIssueWorklog struct {
	Author           JiraAuthor `json:"author"`
	Started          string     `json:"started"`
	TimeSpent        string     `json:"timeSpent"`
	TimeSpentSeconds int        `json:"timeSpentSeconds"`
}

type JiraAuthor struct {
	DisplayName string `json:"displayName"`
	Email       string `json:"emailAddress"`
}
