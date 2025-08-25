package report

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"younsl/promdrop/internal/metrics"
	"younsl/promdrop/internal/utils"
)

func GenerateTxtFile(outputDir, jobName string, metrics []string) {
	safeJobName := utils.SanitizeFilename(jobName)
	txtOutputFile := filepath.Join(outputDir, fmt.Sprintf("unused_metrics_%s.txt", safeJobName))
	txtFile, err := os.Create(txtOutputFile)
	if err != nil {
		log.Printf("[Warning] Failed to create .txt file for job '%s' ('%s'): %v. Continuing with YAML generation.", jobName, txtOutputFile, err)
		return
	}
	defer txtFile.Close()

	writer := bufio.NewWriter(txtFile)
	for _, metricName := range metrics {
		fmt.Fprintln(writer, metricName)
	}
	err = writer.Flush()
	if err != nil {
		log.Printf("[Warning] Failed to write to .txt file for job '%s' ('%s'): %v.", jobName, txtOutputFile, err)
		return
	}
	fmt.Printf("[Info] Created .txt file for job '%s' -> %s\n", jobName, txtOutputFile)
}

func GenerateSummaryFile(outputDir string, summaryData []metrics.JobMetricSummary, sourceJsonPath string) {
	sourceFileBase := filepath.Base(sourceJsonPath)
	summaryFilePath := filepath.Join(outputDir, "summary.txt")
	file, err := os.Create(summaryFilePath)
	if err != nil {
		log.Printf("[Warning] Failed to create summary file ('%s'): %v", summaryFilePath, err)
		return
	}
	defer file.Close()

	err = writeJobSummaryTable(file, summaryData, sourceFileBase)
	if err != nil {
		log.Printf("[Warning] Failed to write summary file content ('%s'): %v", summaryFilePath, err)
	}

	fmt.Printf("[Info] Summary file created -> %s\n", summaryFilePath)
}
