package scanner

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
	"sync"
	"sync/atomic"
	"text/tabwriter"
	"time"
)

// Scanner handles JDK version scanning operations
type Scanner struct {
	config *Config
}

// New creates a new Scanner instance
func New(config *Config) *Scanner {
	return &Scanner{
		config: config,
	}
}

// ScanPods scans all pods in the configured namespaces for Java versions
func (s *Scanner) ScanPods(ctx context.Context) (*ScanResult, error) {
	startTime := time.Now()

	results := make(map[string]map[string]string)
	var totalPods, jdkPods int64

	for _, namespace := range s.config.Namespaces {
		if s.config.Verbose {
			log.Printf("Checking pods in namespace: %s", namespace)
		}

		pods, err := s.getPods(ctx, namespace)
		if err != nil {
			log.Printf("Error getting pods in namespace %s: %v", namespace, err)
			continue
		}

		namespacePods, namespaceJDKPods, err := s.scanNamespace(ctx, namespace, pods)
		if err != nil {
			log.Printf("Error scanning namespace %s: %v", namespace, err)
			continue
		}

		atomic.AddInt64(&totalPods, int64(len(namespacePods)))
		atomic.AddInt64(&jdkPods, int64(namespaceJDKPods))

		// Merge results
		for pod, version := range namespacePods {
			if _, exists := results[namespace]; !exists {
				results[namespace] = make(map[string]string)
			}
			results[namespace][pod] = version
		}
	}

	return &ScanResult{
		TotalPods:   int(totalPods),
		JDKPods:     int(jdkPods),
		ElapsedTime: time.Since(startTime),
		PodVersions: results,
	}, nil
}

// getPods retrieves all pods from a namespace
func (s *Scanner) getPods(ctx context.Context, namespace string) ([]Pod, error) {
	cmd := exec.CommandContext(ctx, "kubectl", "get", "pods", "-n", namespace, "-o", "json")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to get pods: %w", err)
	}

	var podList PodList
	if err := json.Unmarshal(output, &podList); err != nil {
		return nil, fmt.Errorf("failed to parse pod list: %w", err)
	}

	return podList.Items, nil
}

// scanNamespace scans all pods in a specific namespace
func (s *Scanner) scanNamespace(ctx context.Context, namespace string, pods []Pod) (map[string]string, int, error) {
	results := make(map[string]string)
	var jdkCount int64
	var wg sync.WaitGroup
	sem := make(chan struct{}, s.config.MaxGoroutines)
	resultMux := sync.Mutex{}

	for i, pod := range pods {
		if s.config.Verbose {
			log.Printf("Scanning pod %d of %d: %s", i+1, len(pods), pod.Metadata.Name)
		}

		// Skip DaemonSet pods if configured
		if s.config.SkipDaemonSet && s.isDaemonSet(pod) {
			if s.config.Verbose {
				log.Printf("Skipping DaemonSet pod: %s", pod.Metadata.Name)
			}
			continue
		}

		wg.Add(1)
		sem <- struct{}{}

		go func(podName string) {
			defer wg.Done()
			defer func() { <-sem }()

			version, err := s.getJavaVersion(ctx, namespace, podName)
			if err != nil {
				if s.config.Verbose {
					log.Printf("Error getting Java version for pod %s: %v", podName, err)
				}
				version = "Unknown"
			}

			resultMux.Lock()
			results[podName] = version
			if version != "Unknown" {
				atomic.AddInt64(&jdkCount, 1)
			}
			resultMux.Unlock()
		}(pod.Metadata.Name)
	}

	wg.Wait()
	return results, int(jdkCount), nil
}

// getJavaVersion executes java -version command in a pod
func (s *Scanner) getJavaVersion(ctx context.Context, namespace, podName string) (string, error) {
	timeoutCtx, cancel := context.WithTimeout(ctx, s.config.Timeout)
	defer cancel()

	cmd := exec.CommandContext(timeoutCtx, "kubectl", "exec", "-n", namespace, podName, "--", "java", "-version")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("failed to execute java -version: %w", err)
	}

	return parseJavaVersion(string(output)), nil
}

// isDaemonSet checks if a pod is owned by a DaemonSet
func (s *Scanner) isDaemonSet(pod Pod) bool {
	for _, owner := range pod.OwnerReferences {
		if owner.Kind == "DaemonSet" {
			return true
		}
	}
	return false
}

// parseJavaVersion parses Java version from java -version output
func parseJavaVersion(output string) string {
	lines := strings.Split(output, "\n")
	for _, line := range lines {
		if strings.Contains(line, "version") {
			// Extract version from 'java version "1.8.0_292"'
			parts := strings.Split(line, "\"")
			if len(parts) > 1 {
				return parts[1]
			}
		}
	}
	return "Unknown"
}

// PrintResults prints the scan results in a formatted table
func (s *Scanner) PrintResults(result *ScanResult) error {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	defer w.Flush()

	// Print header
	fmt.Fprintln(w, "INDEX\tNAMESPACE\tPOD\tJAVA_VERSION")

	index := 1
	for namespace, pods := range result.PodVersions {
		for podName, version := range pods {
			fmt.Fprintf(w, "%d\t%s\t%s\t%s\n", index, namespace, podName, version)
			index++
		}
	}

	// Print summary
	fmt.Printf("\nScan Summary:\n")
	fmt.Printf("Total pods scanned: %d\n", result.TotalPods)
	fmt.Printf("Pods using JDK: %d\n", result.JDKPods)
	fmt.Printf("Time taken: %dm %ds\n",
		int(result.ElapsedTime.Minutes()),
		int(result.ElapsedTime.Seconds())%60)

	return nil
}
