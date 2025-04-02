package jira

import "errors"

var errorEmailNotConfigured = errors.New("before trying to log work configure Jira user email")
var errorTokenNotConfigured = errors.New("before trying to log work configure Jira token")
var errorHostNotConfigured = errors.New("before trying to log work configure Jira host")
var errorFetchingAssignedIssues = errors.New("failed to fetch assigned issues")
var errorFailedToReadBody = errors.New("failed to read response body")
