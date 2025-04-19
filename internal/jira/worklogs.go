package jira

import (
	"fmt"
	"sort"
	"time"
)

type TaskLog struct {
	Summary    string
	LoggedTime time.Duration
	TaskKey    string
}

func (d *TaskLog) StringLoggedTime() string {
	hours := int(d.LoggedTime.Hours())
	minutes := int(d.LoggedTime.Minutes()) % 60
	return fmt.Sprintf("%dh %dm", hours, minutes)
}

type Logs struct {
	Days []*Day
}

func (t *Logs) AddLog(worklog TaskLog, started time.Time) {
	addedNewDay := func(l TaskLog, s time.Time) bool {
		startedNormalized := started.Truncate(time.Hour * 24)
		for _, day := range t.Days {
			if day.Date.Equal(startedNormalized) {
				for _, taskWorklog := range day.Worklogs {
					if taskWorklog.TaskKey == worklog.TaskKey {
						taskWorklog.LoggedTime += worklog.LoggedTime
						day.TimeLogged += worklog.LoggedTime
						return false
					}
				}
				day.Worklogs = append(day.Worklogs, &worklog)
				day.TimeLogged += worklog.LoggedTime
				sort.Slice(day.Worklogs, func(i, j int) bool {
					return day.Worklogs[i].TaskKey < day.Worklogs[j].TaskKey
				})
				return false
			}
		}
		t.Days = append(t.Days, &Day{
			Date:       startedNormalized,
			Worklogs:   []*TaskLog{&worklog},
			TimeLogged: worklog.LoggedTime,
		})
		return true
	}(worklog, started)
	if addedNewDay {
		sort.Slice(t.Days, func(i, j int) bool {
			return t.Days[i].Date.Before(t.Days[j].Date)
		})
	}
}

type Day struct {
	Date       time.Time
	Worklogs   []*TaskLog
	TimeLogged time.Duration
}

func (d *Day) DateString() string {
	return d.Date.Format(time.DateOnly)
}
