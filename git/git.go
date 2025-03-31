package git

type GitHandler interface {
	GetGitBranch() (string, error)
}
