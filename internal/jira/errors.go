package jira

import "errors"

var errorTokenNotConfigured = errors.New("before trying to connect to jira configure Jira token")
var errorOriginNotConfigured = errors.New("before trying to connect to jira configure Jira origin")
var errorFetchingAssignedIssues = errors.New("failed to fetch assigned issues")
var errorFailedToReadBody = errors.New("failed to read response body")
var errorNoProtocolInOrigin = errors.New("jira origin is not valid. Set proper protocol schema")
