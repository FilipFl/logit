package prompter

type Prompter interface {
	PromptForApprove(string) (bool, error)
	PromptForString(string, string) (string, error)
}
