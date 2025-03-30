package prompter

type MockPrompter struct {
	approveResponses []bool
	approveErrors    []error
	stringResponses  []string
	stringErrors     []error
	approveIndex     int
	stringIndex      int
}

func NewMockPrompter() *MockPrompter {
	return &MockPrompter{
		approveResponses: []bool{},
		approveErrors:    []error{},
		stringResponses:  []string{},
		stringErrors:     []error{},
		approveIndex:     0,
		stringIndex:      0,
	}
}

func (m *MockPrompter) SetApproveResponses(responses []bool, errors []error) {
	m.approveResponses = responses
	m.approveErrors = errors
	m.approveIndex = 0
}

func (m *MockPrompter) SetStringResponses(responses []string, errors []error) {
	m.stringResponses = responses
	m.stringErrors = errors
	m.stringIndex = 0
}

func (m *MockPrompter) PromptForApprove(msg string) (bool, error) {
	if m.approveIndex >= len(m.approveResponses) {
		return false, ErrorWrongApproveInput
	}
	resp := m.approveResponses[m.approveIndex]
	err := error(nil)
	if m.approveErrors != nil && m.approveIndex < len(m.approveErrors) {
		err = m.approveErrors[m.approveIndex]
	}
	m.approveIndex++
	return resp, err
}

func (m *MockPrompter) PromptForString(infoMsg, promptMsg string) (string, error) {
	if m.stringIndex >= len(m.stringResponses) {
		return "", ErrorScanningUserInput
	}
	resp := m.stringResponses[m.stringIndex]
	err := error(nil)
	if m.stringErrors != nil && m.stringIndex < len(m.stringErrors) {
		err = m.stringErrors[m.stringIndex]
	}
	m.stringIndex++
	return resp, err
}
