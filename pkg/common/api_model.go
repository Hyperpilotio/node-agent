package common

type TaskReport struct {
	Id            string `json:"Id,omitempty"`
	Plugin        string `json:"Plugin"`
	LastErrorMsg  string `json:"LastErrorMessage"`
	LastErrorTime int64  `json:"LastErrorTimestamp"`
	FailureCount  int64  `json:"FailureCount"`
}

type PublisherReport struct {
	Id            string `json:"Id,omitempty"`
	Plugin        string `json:"Plugin"`
	LastErrorMsg  string `json:"LastErrorMessage"`
	LastErrorTime int64  `json:"LastErrorTimestamp"`
	FailureCount  int64  `json:"FailureCount"`
}

type Report struct {
	Tasks     map[string]TaskReport      `json:"tasks"`
	Publisher map[string]PublisherReport `json:"publishers"`
}
