package git

type MockGitHandler struct {
	Branch string
	Error  error
}

func NewMockGitHandler() *MockGitHandler {
	return &MockGitHandler{Branch: "", Error: nil}
}

func (h *MockGitHandler) GetGitBranch() (string, error) {
	return h.Branch, h.Error
}
