package jira

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/FilipFl/logit/internal/configuration"
)

type JiraClient struct {
	cfgHandler configuration.ConfigurationHandler
}

type Worklog struct {
	TimeSpent string `json:"timeSpent"`
	Started   string `json:"started"`
	Comment   string `json:"comment,omitempty"`
}

type SearchJql struct {
	Fields     []string `json:"fields"`
	MaxResults int      `json:"maxResults"`
	JQL        string   `json:"jql"`
	StartAt    int      `json:"startAt"`
}

func NewJiraClient(cfgHandler configuration.ConfigurationHandler) *JiraClient {
	return &JiraClient{
		cfgHandler,
	}
}

func (c *JiraClient) LogTime(taskKey string, duration time.Duration, started time.Time, comment string) error {
	endpoint := fmt.Sprintf("/rest/api/2/issue/%s/worklog", taskKey)
	timeSpent := fmt.Sprintf("%dh %dm", int(duration.Hours()), int(duration.Minutes())%60)
	worklog := Worklog{
		TimeSpent: timeSpent,
		Started:   started.Format("2006-01-02T15:04:05.000-0700"),
		Comment:   comment,
	}
	jsonData, err := json.Marshal(worklog)
	if err != nil {
		return err
	}
	resp, err := c.callPost(endpoint, jsonData, c.assertConfigurationIsValid)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to log time: %s", string(body))
	}
	return nil
}

func (c *JiraClient) GetAssignedIssues() ([]Issue, error) {
	endpoint := "/rest/api/2/search"
	data := SearchJql{
		Fields:     []string{"key", "summary", "status", "assignee"},
		JQL:        "assignee = currentUser() AND status not in (Done, Closed)",
		MaxResults: 100,
		StartAt:    0,
	}
	jsonData, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}
	resp, err := c.callPost(endpoint, jsonData, c.assertConfigurationIsValid)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, errorFailedToReadBody
	}
	if resp.StatusCode == http.StatusOK {
		issuesResults := make([]Issue, 0)
		var result Result
		err := json.Unmarshal(body, &result)
		if err != nil {
			return nil, err
		}
		for _, issue := range result.Issues {
			issueResult := Issue{}
			issueResult.Key = issue.Key
			issueResult.Summary = issue.Fields.Summary
			issueResult.Status = issue.Fields.Status.Name
			issuesResults = append(issuesResults, issueResult)
		}
		return issuesResults, nil

	}

	return nil, errorFetchingAssignedIssues
}

func (c *JiraClient) GetLoggedTime(fromDays int) (Logs, error) {
	resultLogs := Logs{}
	endpoint := "/rest/api/2/search"
	data := SearchJql{
		Fields:     []string{"key", "summary"},
		JQL:        fmt.Sprintf("worklogAuthor = currentUser() AND worklogDate > -%dd", fromDays),
		MaxResults: 100,
		StartAt:    0,
	}
	jsonData, err := json.Marshal(data)
	if err != nil {
		return resultLogs, err
	}
	resp, err := c.callPost(endpoint, jsonData, c.assertConfigurationForFetchingWorklogsIsValid)
	if err != nil {
		return resultLogs, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return resultLogs, errorFailedToReadBody
	}
	if resp.StatusCode != http.StatusOK {
		return resultLogs, errorFetchingAssignedIssues
	}

	var result Result
	err = json.Unmarshal(body, &result)
	if err != nil {
		return resultLogs, err
	}

	var wg sync.WaitGroup
	fromDaysDuration := time.Hour * time.Duration(fromDays) * 24
	fromDaysBoundaryTime := time.Now().AddDate(0, 0, -(fromDays))
	i := 0
	for _, issue := range result.Issues {
		wg.Add(1)
		go func(issue JiraIssue) {
			defer wg.Done()

			logs, err := c.getAllWorklogs(issue.Key, fromDaysDuration)
			if err != nil {
				fmt.Printf("Error fetching logs for %s: %v\n", issue.Key, err)
				return
			}

			for _, log := range logs {
				if strings.ToLower(log.Author.Email) != strings.ToLower(c.cfgHandler.LoadConfig().JiraEmail) {
					continue
				}
				startTime, err := time.Parse("2006-01-02T15:04:05.000-0700", log.Started)
				if err != nil {
					continue
				}
				if startTime.Truncate(24 * time.Hour).Before(fromDaysBoundaryTime) {
					continue
				}
				worklog := TaskLog{
					Summary:    issue.Fields.Summary,
					LoggedTime: time.Duration(log.TimeSpentSeconds) * time.Second,
					TaskKey:    issue.Key,
				}
				resultLogs.AddLog(worklog, startTime)
			}
		}(issue)
		i++
		fmt.Printf("Completed fetching %d/%d tasks.\n", i, len(result.Issues))
	}

	wg.Wait()
	return resultLogs, nil
}

func (c *JiraClient) getAllWorklogs(issueKey string, days time.Duration) ([]JiraIssueWorklog, error) {
	startAt := 0
	pageSize := 5000
	allWorklogs := []JiraIssueWorklog{}
	startedAfter := time.Now().Add(-1 * days).Unix()
	for {
		endpoint := fmt.Sprintf("/rest/api/2/issue/%s/worklog?startAt=%d&maxResults=%d&startedAfter=%d", issueKey, startAt, pageSize, startedAfter)
		resp, err := c.callGet(endpoint)
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}

		var container JiraWorklogs
		if err := json.Unmarshal(body, &container); err != nil {
			return nil, err
		}

		allWorklogs = append(allWorklogs, container.Worklogs...)

		if len(container.Worklogs) < pageSize {
			break
		}
		startAt += pageSize
	}

	return allWorklogs, nil
}

func (c *JiraClient) callPost(endpoint string, jsonData []byte, validateFunc func() error) (*http.Response, error) {
	err := validateFunc()
	if err != nil {
		return nil, err
	}
	url := fmt.Sprintf("%s%s", c.cfgHandler.LoadConfig().JiraOrigin, endpoint)

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.cfgHandler.GetToken()))
	client := &http.Client{}
	return client.Do(req)
}

func (c *JiraClient) callGet(endpoint string) (*http.Response, error) {
	err := c.assertConfigurationIsValid()
	if err != nil {
		return nil, err
	}
	url := fmt.Sprintf("%s%s", c.cfgHandler.LoadConfig().JiraOrigin, endpoint)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.cfgHandler.GetToken()))
	client := &http.Client{}
	return client.Do(req)
}

func (c *JiraClient) assertConfigurationIsValid() error {
	token := c.cfgHandler.GetToken()
	if token == "" {
		if c.cfgHandler.LoadConfig().JiraTokenEnvName != "" {
			return errorTokenEnvNameSetButEmpty
		}
		return errorTokenNotConfigured
	}
	origin := c.cfgHandler.LoadConfig().JiraOrigin
	if origin == "" {
		return errorOriginNotConfigured
	}
	if !strings.HasPrefix(c.cfgHandler.LoadConfig().JiraOrigin, "https://") && !strings.HasPrefix(c.cfgHandler.LoadConfig().JiraOrigin, "http://") {
		return errorNoProtocolInOrigin
	}
	return nil
}

func (c *JiraClient) assertConfigurationForFetchingWorklogsIsValid() error {
	err := c.assertConfigurationIsValid()
	if err != nil {
		return err
	}
	email := c.cfgHandler.LoadConfig().JiraEmail
	if email == "" {
		return errorEmailNotConfigured
	}
	return nil
}
