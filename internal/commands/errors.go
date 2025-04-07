package commands

import "errors"

var errorNoJiraTicket = errors.New("no Jira ticket found in passed string")
var errorNoTargetToLogWork = errors.New("no target to log work")
var errorNoSnapshot = errors.New("no start time saved")
var errorCantUseSnapshotAndYesterday = errors.New("can't log time from snapshot for yesterday")
var errorWrongDuration = errors.New("duration to log is invalid")
var errorOperationAborted = errors.New("operation aborted by user")

var errorInvalidDateFormat = errors.New("invalid date format passed with date flag; accepted format either dd.mm or dd-mm")
var errorInvalidMonth = errors.New("invalid month")
var errorInvalidDay = errors.New("invalid day for called month")
var errorAliasAndTask = errors.New("alias and task flags are mutually exclusive")
var errorYesterdayAndDate = errors.New("yesterday and date flags are mutually exlusive")
var errorSnapshotNotToday = errors.New("unable to log time from snapshot for day other than today")
