package main

import "errors"

var errorNoJiraTicket = errors.New("no Jira ticket found in passed string")
var errorNoTargetToLogWork = errors.New("no target to log work")
var errorNoSnapshot = errors.New("no start time saved")
var errorWrongDuration = errors.New("duration to log is invalid")
var errorOperationAborted = errors.New("operation aborted by user")
var errorEmailNotConfigured = errors.New("before trying to log work configure email")
var errorTokenNotConfigured = errors.New("before trying to log work configure Jira token")
