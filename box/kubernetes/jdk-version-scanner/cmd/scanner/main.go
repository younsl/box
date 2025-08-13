package main

import (
	"context"
	"flag"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/younsl/box/tools/jdk-version-scanner/internal/scanner"
)

func main() {
	var (
		namespaces    = flag.String("namespaces", "default", "Comma-separated list of namespaces to scan")
		maxGoroutines = flag.Int("max-goroutines", 20, "Maximum number of concurrent goroutines")
		timeout       = flag.Duration("timeout", 30*time.Second, "Timeout for kubectl commands")
		skipDaemonSet = flag.Bool("skip-daemonset", true, "Skip DaemonSet pods")
		verbose       = flag.Bool("verbose", false, "Enable verbose logging")
	)
	flag.Parse()

	if *verbose {
		log.SetFlags(log.LstdFlags | log.Lshortfile)
	}

	// Create context with timeout and signal handling
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()

	// Handle graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigChan
		log.Println("Received shutdown signal, canceling operations...")
		cancel()
	}()

	config := &scanner.Config{
		Namespaces:    parseNamespaces(*namespaces),
		MaxGoroutines: *maxGoroutines,
		Timeout:       *timeout,
		SkipDaemonSet: *skipDaemonSet,
		Verbose:       *verbose,
	}

	s := scanner.New(config)
	results, err := s.ScanPods(ctx)
	if err != nil {
		log.Fatalf("Failed to scan pods: %v", err)
	}

	if err := s.PrintResults(results); err != nil {
		log.Fatalf("Failed to print results: %v", err)
	}
}

func parseNamespaces(input string) []string {
	namespaces := strings.Split(input, ",")
	for i := range namespaces {
		namespaces[i] = strings.TrimSpace(namespaces[i])
	}
	return namespaces
}
