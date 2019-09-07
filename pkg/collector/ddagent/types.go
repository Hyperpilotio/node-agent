package ddagent

type Metric struct {
	MetricName string            `json:"metric"`
	Value      string            `json:"value"`
	Timestamp  int64             `json:"time_stamp"`
	Host       string            `json:"host"`
	Tags       map[string]string `json:"tags,omitempty"`
}
