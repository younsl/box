package checker

import (
	"fmt"
	"net/http"
	"os"
	"strings"
	"sync"
	"text/tabwriter"
	"time"

	"github.com/briandowns/spinner"
	"github.com/younsl/kk/internal/logger"
)

const (
	maxRetries     = 3               // Maximum number of retries
	retryInterval  = 2 * time.Second // Interval between retries
	requestTimeout = 2 * time.Second // Request timeout
)

// CheckResult holds the result of a single domain/URL check
type CheckResult struct {
	InputItem  string // Original input string (URL or domain)
	CheckedURL string // The actual URL used for the check
	Duration   time.Duration
	Status     string
	StatusCode string
	Attempts   int
}

// normalizeURL ensures the URL has a scheme (http:// or https://)
func normalizeURL(domainOrURL string) string {
	if !strings.HasPrefix(domainOrURL, "http://") && !strings.HasPrefix(domainOrURL, "https://") {
		return "https://" + domainOrURL
	}
	return domainOrURL
}

// isSuccessfulStatus checks if the HTTP status code indicates success (2xx)
func isSuccessfulStatus(statusCode int) bool {
	return statusCode >= 200 && statusCode < 300
}

// performHTTPRequest performs a single HTTP request and returns the result
func performHTTPRequest(url string, httpClient *http.Client) (statusCode int, duration time.Duration, err error) {
	logger.Log.Debugf("Performing HTTP request to %s", url)
	startTime := time.Now()
	resp, err := httpClient.Get(url)
	duration = time.Since(startTime)

	if err != nil {
		logger.Log.Debugf("HTTP request failed for %s: %v", url, err)
		return 0, duration, err
	}
	defer resp.Body.Close()

	logger.Log.Debugf("HTTP request completed for %s: status=%d, duration=%v", url, resp.StatusCode, duration)
	return resp.StatusCode, duration, nil
}

// performCheck performs the check and retry logic for a single domain/URL
func performCheck(domainOrURL string, httpClient *http.Client) CheckResult {
	checkedURL := normalizeURL(domainOrURL)

	var lastDuration time.Duration
	status := "FAILED"
	statusCodeStr := "-"

	for attempt := 1; attempt <= maxRetries; attempt++ {
		statusCode, duration, err := performHTTPRequest(checkedURL, httpClient)
		lastDuration = duration

		if err == nil {
			statusCodeStr = fmt.Sprintf("%d", statusCode)
			if isSuccessfulStatus(statusCode) {
				status = "OK"
				return newCheckResult(domainOrURL, checkedURL, lastDuration, status, statusCodeStr, attempt)
			}
			status = "UNEXPECTED_CODE"
		}

		// Retry if not the last attempt and not successful
		if attempt < maxRetries {
			time.Sleep(retryInterval)
		}
	}

	return newCheckResult(domainOrURL, checkedURL, lastDuration, status, statusCodeStr, maxRetries)
}

// newCheckResult creates a new CheckResult instance
func newCheckResult(inputItem, checkedURL string, duration time.Duration, status, statusCode string, attempts int) CheckResult {
	return CheckResult{
		InputItem:  inputItem,
		CheckedURL: checkedURL,
		Duration:   duration,
		Status:     status,
		StatusCode: statusCode,
		Attempts:   attempts,
	}
}

// formatAttempts formats the attempts string based on the check status
func formatAttempts(status string, attempts int) string {
	if status == "OK" {
		return fmt.Sprintf("%d", attempts)
	}
	return fmt.Sprintf("%d (failed)", attempts)
}

// printResult prints a single check result to the table writer
func printResult(w *tabwriter.Writer, result CheckResult) {
	fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n",
		result.CheckedURL,
		result.Duration.Round(time.Millisecond),
		result.Status,
		result.StatusCode,
		formatAttempts(result.Status, result.Attempts),
	)
}

// createSpinner creates and starts a new spinner with default settings
func createSpinner() *spinner.Spinner {
	s := spinner.New(spinner.CharSets[9], 100*time.Millisecond)
	s.Suffix = " Knock knock... Who's there?"
	s.HideCursor = true
	s.Start()
	return s
}

// processResults processes check results from the channel and updates the spinner
func processResults(resultsChan chan CheckResult, w *tabwriter.Writer, s *spinner.Spinner, totalChecks int) int {
	successCount := 0
	completedCount := 0

	for result := range resultsChan {
		completedCount++
		if result.Status == "OK" {
			successCount++
		}
		s.Suffix = fmt.Sprintf(" Knock knock... Who's there? (%d/%d door is open!)", completedCount, totalChecks)
		printResult(w, result)
	}

	return successCount
}

// RunChecks checks the given list of domains/URLs in parallel and prints results as a table
func RunChecks(domainsOrURLs []string) {
	logger.Log.Infof("Starting domain checks for %d domains", len(domainsOrURLs))
	totalStartTime := time.Now()
	totalChecks := len(domainsOrURLs)

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "URL\tTIME\tSTATUS\tCODE\tATTEMPTS")

	httpClient := &http.Client{Timeout: requestTimeout}
	resultsChan := make(chan CheckResult, totalChecks)

	s := createSpinner()

	// Launch parallel checks
	logger.Log.Debugf("Launching parallel checks for %d domains", totalChecks)
	var wg sync.WaitGroup
	for _, item := range domainsOrURLs {
		wg.Add(1)
		go func(checkItem string) {
			defer wg.Done()
			resultsChan <- performCheck(checkItem, httpClient)
		}(item)
	}

	// Close results channel when all checks complete
	go func() {
		wg.Wait()
		close(resultsChan)
		logger.Log.Debug("All checks completed")
	}()

	// Process results and update spinner
	successCount := processResults(resultsChan, w, s, totalChecks)

	s.Stop()
	w.Flush()

	totalDuration := time.Since(totalStartTime)

	// Print summary
	fmt.Printf("\nSummary: %d/%d successful checks in %.1fs.\n",
		successCount,
		totalChecks,
		totalDuration.Seconds(),
	)

	logger.Log.Infof("Checks completed: %d/%d successful in %.1fs", successCount, totalChecks, totalDuration.Seconds())
}
