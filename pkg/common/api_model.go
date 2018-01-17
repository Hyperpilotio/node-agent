package common

type TaskReport struct {
	LastErrorMsg   string `json:"message"`
	LastErrorTime  int64  `json:"timestamp"`
}
