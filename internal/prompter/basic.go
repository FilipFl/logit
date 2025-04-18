package prompter

import (
	"fmt"
)

type BasicPrompter struct {
}

func NewBasicPrompter() *BasicPrompter {
	return &BasicPrompter{}
}

func (p *BasicPrompter) PromptForApprove(msg string) (bool, error) {
	fmt.Println(msg)
	fmt.Println(ApprovePrompt)
	promptedApprove := ""
	fmt.Scanln(&promptedApprove)
	if promptedApprove != "" {
		switch promptedApprove {
		case "y":
			return true, nil
		case "Y":
			return true, nil
		case "n":
			return false, nil
		case "N":
			return false, nil
		default:
			return false, ErrorWrongApproveInput
		}
	}
	return false, ErrorWrongApproveInput
}

func (p *BasicPrompter) PromptForString(infoMsg, promptMsg string) (string, error) {
	fmt.Println(infoMsg)
	fmt.Println(promptMsg)
	promptedString := ""
	fmt.Scanln(&promptedString)
	if promptedString != "" {
		return promptedString, nil
	}
	return "", ErrorScanningUserInput
}
