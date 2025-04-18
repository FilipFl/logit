package jira

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
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
	resp, err := c.callPost(endpoint, jsonData)
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
	resp, err := c.callPost(endpoint, jsonData)
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
		result := make(map[string]interface{})
		err := json.Unmarshal(body, &result)
		if err != nil {
			return nil, err
		}
		issues := result["issues"].([]interface{})
		for _, issue := range issues {
			issueResult := Issue{}
			issueDetails := issue.(map[string]interface{})
			issueResult.Key = issueDetails["key"].(string)
			fields := issueDetails["fields"].(map[string]interface{})
			issueResult.Summary = fields["summary"].(string)
			issueResult.Status = fields["status"].(map[string]interface{})["name"].(string)
			issuesResults = append(issuesResults, issueResult)
		}
		return issuesResults, nil

	}

	return nil, errorFetchingAssignedIssues
}

func (c *JiraClient) callPost(endpoint string, jsonData []byte) (*http.Response, error) {
	err := c.assertConfigurationIsValid()
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

func (c *JiraClient) assertConfigurationIsValid() error {
	token := c.cfgHandler.GetToken()
	if token == "" {
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
