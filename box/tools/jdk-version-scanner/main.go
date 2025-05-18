package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"sync"
	"text/tabwriter"
	"time"
)

type PodList struct {
	Items []struct {
		Metadata struct {
			Name string `json:"name"`
		} `json:"metadata"`
		OwnerReferences []struct {
			Kind string `json:"kind"`
		} `json:"ownerReferences"`
	} `json:"items"`
}

const (
	maxGoroutines    = 20
	targetNamespaces = "default" // Define namespaces as a comma-separated string constant
)

func main() {
	startTime := time.Now()

	// Split the constant string into a slice of namespaces
	namespaces := strings.Split(targetNamespaces, ",")
	// Trim whitespace from each namespace name
	for i := range namespaces {
		namespaces[i] = strings.TrimSpace(namespaces[i])
	}

	results := make(map[string]map[string]string) // Map to store results

	totalPods := 0
	jdkPods := 0

	for _, namespace := range namespaces {
		fmt.Printf("Checking pods in namespace: %s\n", namespace)

		// Get the list of all pods in the namespace
		cmd := exec.Command("kubectl", "get", "pods", "-n", namespace, "-o", "json")
		output, err := cmd.Output()
		if err != nil {
			fmt.Printf("Error getting pods in namespace %s: %v\n", namespace, err)
			continue
		}

		var podList PodList
		if err := json.Unmarshal(output, &podList); err != nil {
			fmt.Printf("Error parsing JSON for namespace %s: %v\n", namespace, err)
			continue
		}

		var wg sync.WaitGroup
		sem := make(chan struct{}, maxGoroutines) // Use constant to manage goroutine count

		for index, item := range podList.Items {
			totalPods++ // Increase total pod count
			fmt.Printf("Scanning pod %d of %d: Checking Java version for pod %s\n", index+1, len(podList.Items), item.Metadata.Name)

			// Filter out DaemonSet owners
			isDaemonSet := false
			for _, owner := range item.OwnerReferences {
				if owner.Kind == "DaemonSet" {
					isDaemonSet = true
					break
				}
			}
			if isDaemonSet {
				fmt.Printf("Skipping DaemonSet pod: %s\n", item.Metadata.Name)
				continue
			}

			wg.Add(1)
			sem <- struct{}{} // Add value to semaphore channel

			go func(podName string) {
				defer wg.Done()
				defer func() { <-sem }() // Remove value from semaphore channel after work is done

				cmd := exec.Command("kubectl", "exec", "-n", namespace, podName, "--", "java", "-version")
				output, err := cmd.CombinedOutput()
				if err != nil {
					fmt.Printf("Error executing command for pod %s in namespace %s: %v\n", podName, namespace, err)
					return
				}

				// Parse JDK version
				version := parseJavaVersion(string(output))
				if _, exists := results[namespace]; !exists {
					results[namespace] = make(map[string]string)
				}
				results[namespace][podName] = version

				if version != "Unknown" {
					jdkPods++ // Increase count of pods using JDK
				}
			}(item.Metadata.Name)
		}

		wg.Wait() // Wait for all goroutines to complete
	}

	// Print aggregated results using tabwriter for kubectl-like formatting
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0) // minwidth, tabwidth, padding=2, padchar, flags
	// Print header in uppercase
	fmt.Fprintln(w, "INDEX\tNAMESPACE\tPOD\tJAVA_VERSION")
	index := 1 // Initialize index
	for namespace, pods := range results {
		for podName, version := range pods {
			// Use Fprintf with tabs (\t) to align columns
			fmt.Fprintf(w, "%d\t%s\t%s\t%s\n", index, namespace, podName, version)
			index++ // Increase index
		}
	}
	w.Flush() // Ensure all buffered output is written

	// Record completion time and calculate elapsed time
	elapsedTime := time.Since(startTime)
	fmt.Printf("Total pods scanned: %d\n", totalPods)
	fmt.Printf("Pods using JDK: %d\n", jdkPods)
	fmt.Printf("Time taken: %dm %ds\n", int(elapsedTime.Minutes()), int(elapsedTime.Seconds())%60)
}

// Function to parse JDK version string
func parseJavaVersion(output string) string {
	lines := strings.Split(output, "\n")
	for _, line := range lines {
		if strings.Contains(line, "version") {
			// Extract only the version part from e.g., 'java version "1.8.0_292"'
			parts := strings.Split(line, "\"")
			if len(parts) > 1 {
				return parts[1]
			}
		}
	}
	return "Unknown"
}
