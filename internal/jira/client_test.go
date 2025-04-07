package jira

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/FilipFl/logit/internal/configuration"
	"github.com/stretchr/testify/assert"
)

func TestLogTime_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/rest/api/3/issue/TEST-123/worklog", r.URL.Path)
		assert.Equal(t, "POST", r.Method)

		body, _ := io.ReadAll(r.Body)
		assert.Contains(t, string(body), "Working on task")

		w.WriteHeader(http.StatusCreated)
	}))
	defer server.Close()

	mockCfg := configuration.NewMockConfigurationHandler()
	mockCfg.SetConfig(&configuration.Config{
		JiraEmail:  "user@example.com",
		JiraOrigin: server.URL,
		JiraToken:  "token123",
	})

	client := NewJiraClient(mockCfg)
	err := client.LogTime("TEST-123", 90*time.Minute, time.Now(), "Working on task")
	assert.NoError(t, err)
}

func TestLogTime_FailureStatus(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Bad Request"))
	}))
	defer server.Close()

	mockCfg := configuration.NewMockConfigurationHandler()
	mockCfg.SetConfig(&configuration.Config{
		JiraEmail:  "user@example.com",
		JiraOrigin: server.URL,
		JiraToken:  "token123",
	})

	client := NewJiraClient(mockCfg)
	err := client.LogTime("TEST-123", 1*time.Hour, time.Now(), "Logging failed task")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to log time")
}

func TestGetAssignedIssues_Success(t *testing.T) {
	responseJSON := `{
		"issues": [
			{
				"key": "ISSUE-1",
				"fields": {
					"summary": "Fix bug",
					"status": { "name": "In Progress" }
				}
			},
			{
				"key": "ISSUE-2",
				"fields": {
					"summary": "Add feature",
					"status": { "name": "To Do" }
				}
			}
		]
	}`

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/rest/api/3/search/jql", r.URL.Path)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(responseJSON))
	}))
	defer server.Close()

	mockCfg := configuration.NewMockConfigurationHandler()
	mockCfg.SetConfig(&configuration.Config{
		JiraEmail:  "user@example.com",
		JiraOrigin: server.URL,
		JiraToken:  "token123",
	})

	client := NewJiraClient(mockCfg)
	issues, err := client.GetAssignedIssues()
	assert.NoError(t, err)
	assert.Len(t, issues, 2)
	assert.Equal(t, "ISSUE-1", issues[0].Key)
	assert.Equal(t, "Fix bug", issues[0].Summary)
	assert.Equal(t, "In Progress", issues[0].Status)
}

func TestGetAssignedIssues_HTTPError(t *testing.T) {
	mockCfg := configuration.NewMockConfigurationHandler()
	mockCfg.SetConfig(&configuration.Config{
		JiraEmail:  "user@example.com",
		JiraOrigin: "nonexistent.invalid", // force HTTP error
		JiraToken:  "token123",
	})

	client := NewJiraClient(mockCfg)
	issues, err := client.GetAssignedIssues()
	assert.Error(t, err)
	assert.Nil(t, issues)
}
