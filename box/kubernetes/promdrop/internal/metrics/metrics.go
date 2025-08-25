package metrics

type JobMetricSummary struct {
	JobName     string
	MetricCount int
}

type GroupInfo struct {
	Prefix  string
	Pattern string
	Part    string
	Count   int
}

type RawMetricCount struct {
	Metric    string        `json:"metric"`
	JobCounts []RawJobCount `json:"job_counts"`
}

type RawJobCount struct {
	Job string `json:"job"`
}
