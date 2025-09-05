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

// performCheck performs the check and retry logic for a single domain/URL
func performCheck(domainOrURL string, httpClient *http.Client) CheckResult {
	checkedURL := domainOrURL
	if !strings.HasPrefix(domainOrURL, "http://") && !strings.HasPrefix(domainOrURL, "https://") {
		checkedURL = "https://" + domainOrURL
	}

	var resp *http.Response
	var duration time.Duration
	status := "FAILED"
	statusCode := "-"
	attempts := 0

	for i := 0; i < maxRetries; i++ {
		attempts = i + 1
		startTime := time.Now()
		tempResp, tempErr := httpClient.Get(checkedURL)
		duration = time.Since(startTime)

		if tempErr == nil {
			resp = tempResp
			statusCode = fmt.Sprintf("%d", resp.StatusCode)
			if resp.StatusCode >= 200 && resp.StatusCode < 300 {
				status = "OK"
				resp.Body.Close()
				break
			} else {
				status = "UNEXPECTED_CODE"
			}
			resp.Body.Close()
		} else {
			statusCode = "-"
		}

		if i < maxRetries-1 && status != "OK" {
			time.Sleep(retryInterval)
		}
	}

	return CheckResult{
		InputItem:  domainOrURL,
		CheckedURL: checkedURL,
		Duration:   duration,
		Status:     status,
		StatusCode: statusCode,
		Attempts:   attempts,
	}
}

// RunChecks checks the given list of domains/URLs in parallel and prints results as a table
func RunChecks(domainsOrURLs []string) {
	totalStartTime := time.Now()

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "URL\tTIME\tSTATUS\tCODE\tATTEMPTS")

	httpClient := &http.Client{
		Timeout: requestTimeout,
	}

	var wg sync.WaitGroup
	resultsChan := make(chan CheckResult, len(domainsOrURLs))
	successCount := 0
	completedCount := 0 // Counter for completed checks
	totalChecks := len(domainsOrURLs)

	// Spinner setup
	s := spinner.New(spinner.CharSets[9], 100*time.Millisecond)
	s.Suffix = " Knock knock... Who's there?" // Initial message
	s.HideCursor = true                       // Hide cursor while spinner is active
	s.Start()

	for _, item := range domainsOrURLs {
		wg.Add(1)
		go func(checkItem string) {
			defer wg.Done()
			result := performCheck(checkItem, httpClient)
			resultsChan <- result
		}(item)
	}

	go func() {
		wg.Wait()
		close(resultsChan)
	}()

	// Loop to receive results and update spinner
	for result := range resultsChan {
		completedCount++ // Increment completed counter
		if result.Status == "OK" {
			successCount++
		}
		s.Suffix = fmt.Sprintf(" Knock knock... Who's there? (%d/%d door is open!)", completedCount, totalChecks)

		var attemptsStr string
		if result.Status == "OK" {
			attemptsStr = fmt.Sprintf("%d", result.Attempts)
		} else {
			attemptsStr = fmt.Sprintf("%d (failed)", result.Attempts)
		}

		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n",
			result.CheckedURL,
			result.Duration.Round(time.Millisecond),
			result.Status,
			result.StatusCode,
			attemptsStr,
		)
	}

	s.Stop() // Stop spinner after processing all results
	w.Flush()

	totalDuration := time.Since(totalStartTime)
	// Print summary information
	fmt.Printf("\nSummary: %d/%d successful checks in %.1fs.\n",
		successCount,
		totalChecks,
		totalDuration.Seconds(),
	)
}
