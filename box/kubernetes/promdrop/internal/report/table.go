package report

import (
	"fmt"
	"io"
	"sort"
	"text/tabwriter"
	"younsl/promdrop/internal/metrics"
)

// writeJobSummaryTable formats and writes JobMetricSummary data to the writer.
// Optional sourceFile parameter adds a source column.
func writeJobSummaryTable(writer io.Writer, summaryData []metrics.JobMetricSummary, sourceFile string) error {
	// Sort data for consistent output
	sort.SliceStable(summaryData, func(i, j int) bool {
		return summaryData[i].JobName < summaryData[j].JobName
	})

	w := tabwriter.NewWriter(writer, 0, 8, 2, ' ', 0)

	totalMetricsOverall := 0

	// Print header
	if sourceFile != "" {
		fmt.Fprintln(w, "SOURCE FILE\tJOB NAME\tMETRIC COUNT\t")
	} else {
		fmt.Fprintln(w, "JOB NAME\tMETRIC COUNT\t")
	}

	// Print data rows
	for _, item := range summaryData {
		if sourceFile != "" {
			fmt.Fprintf(w, "%s\t%s\t%d\t\n", sourceFile, item.JobName, item.MetricCount)
		} else {
			fmt.Fprintf(w, "%s\t%d\t\n", item.JobName, item.MetricCount)
		}
		totalMetricsOverall += item.MetricCount
	}

	// Print footer/totals
	totalJobCount := len(summaryData)
	if sourceFile != "" {
		fmt.Fprintln(w, "--------\t--------\t---------\t")
		fmt.Fprintf(w, "%s\t%d jobs\t%d\t\n", "TOTAL", totalJobCount, totalMetricsOverall)
	} else {
		fmt.Fprintln(w, "--------\t---------\t")
		fmt.Fprintf(w, "%s (%d jobs)\t%d\t\n", "TOTAL", totalJobCount, totalMetricsOverall)
	}

	// Ensure all data is written to the underlying writer
	return w.Flush()
}
