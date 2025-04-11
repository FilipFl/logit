package jira

import "time"

type Client interface {
	LogTime(taskKey string, duration time.Duration, started time.Time, comment string) error
	GetAssignedIssues() ([]Issue, error)
}

type Issue struct {
	Summary string `json:"summary"`
	Status  string `json:"status"`
	Key     string `json:"key"`
}
