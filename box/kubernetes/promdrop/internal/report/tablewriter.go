package report

import (
	"fmt"
	"os"
	"younsl/promdrop/internal/metrics"
)

// PrintSummaryTable prints the summary data using the common table writer.
func PrintSummaryTable(summaryData []metrics.JobMetricSummary) {
	fmt.Println("[Summary of Metric Files to Process]")

	// Use the new common table writing function, writing to os.Stdout
	// No source file needed for console output
	err := writeJobSummaryTable(os.Stdout, summaryData, "")
	if err != nil {
		// Log error? Or handle differently for stdout?
		// For now, just print the error to stderr, but maybe log is better.
		fmt.Fprintf(os.Stderr, "[Error] Failed to print summary table: %v\n", err)
	}

	fmt.Printf("Processing a total of %d jobs.\n", len(summaryData))
}
