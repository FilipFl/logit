package jira

import "errors"

var errorTokenNotConfigured = errors.New("before trying to connect to Jira configure Jira token")
var errorOriginNotConfigured = errors.New("before trying to connect to Jira configure Jira origin")
var errorFetchingAssignedIssues = errors.New("failed to fetch assigned issues")
var errorFailedToReadBody = errors.New("failed to read response body")
var errorNoProtocolInOrigin = errors.New("jira origin is not valid. Set proper protocol schema")
var errorEmailNotConfigured = errors.New("before trying this operation configure Jira email")
var errorTokenEnvNameSetButEmpty = errors.New("env token name is configured but it's not set properly in Your system")
