package git

import (
	"os/exec"
	"strings"
)

type BasicGitHandler struct{}

func NewBasicGitHandler() *BasicGitHandler {
	return &BasicGitHandler{}
}

func (h *BasicGitHandler) GetGitBranch() (string, error) {
	cmd := exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD")
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(output)), nil
}
