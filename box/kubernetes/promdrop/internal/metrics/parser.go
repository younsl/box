package metrics

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
)

// PrometheusMetricsFile holds the top-level structure of the input JSON.
type PrometheusMetricsFile struct {
	AdditionalMetricCounts []RawMetricCount `json:"additional_metric_counts"`
}

// ParseAndExtractMetrics reads the JSON file, parses it,
// and extracts unused metrics per job.
// It returns a map of jobName to its metric list and a slice of job summaries.
func ParseAndExtractMetrics(jsonPath string) (map[string][]string, []JobMetricSummary, error) {
	byteValue, err := os.ReadFile(jsonPath)
	if err != nil {
		return nil, nil, fmt.Errorf("error reading JSON file: %w", err)
	}

	var data PrometheusMetricsFile
	err = json.Unmarshal(byteValue, &data)
	if err != nil {
		return nil, nil, fmt.Errorf("error unmarshalling JSON: %w", err)
	}

	jobMetricsMap := make(map[string][]string)
	jobMetricCounts := make(map[string]int)

	for _, metricCount := range data.AdditionalMetricCounts {
		metricName := metricCount.Metric
		if metricName == "" {
			continue
		}
		for _, jobCount := range metricCount.JobCounts {
			jobName := jobCount.Job
			if jobName == "" {
				continue
			}

			found := false
			for _, existingMetric := range jobMetricsMap[jobName] {
				if existingMetric == metricName {
					found = true
					break
				}
			}
			if !found {
				jobMetricsMap[jobName] = append(jobMetricsMap[jobName], metricName)
				jobMetricCounts[jobName]++
			}
		}
	}

	var summaryData []JobMetricSummary
	for jobName, metrics := range jobMetricsMap {
		sort.Strings(metrics)
		summaryData = append(summaryData, JobMetricSummary{
			JobName:     jobName,
			MetricCount: jobMetricCounts[jobName],
		})
	}

	sort.Slice(summaryData, func(i, j int) bool {
		return summaryData[i].JobName < summaryData[j].JobName
	})

	return jobMetricsMap, summaryData, nil
}
