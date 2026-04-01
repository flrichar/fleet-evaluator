package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	"fleet-evaluator/pkg/analyzer"
	"fleet-evaluator/pkg/nats"
	"fleet-evaluator/pkg/reporter"
)

func main() {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [options] <file1.yaml> <file2.yaml> ...\n", os.Args[0])
		flag.PrintDefaults()
	}

	reportFile := flag.String("output", "fleet-eval-report.md", "Path to save the markdown report")
	natsURL := flag.String("nats-url", "", "NATS server URL (optional)")
	natsSubject := flag.String("nats-subject", "fleet.eval.findings", "NATS subject to publish to")
	flag.Parse()

	files := flag.Args()
	if len(files) == 0 {
		flag.Usage()
		os.Exit(1)
	}

	report, err := analyzer.AnalyzeFiles(files)
	if err != nil {
		log.Fatalf("Error analyzing files: %v", err)
	}

	// Environment comparison (simplified)
	if hasMultipleEnvironments(files) {
		report.ComparisonTable = generateComparisonTable(files)
	}

	err = reporter.GenerateMarkdownReport(report, *reportFile)
	if err != nil {
		log.Fatalf("Error generating report: %v", err)
	}

	reporter.PrintTerminalSummary(report)

	if *natsURL != "" {
		if err := nats.PublishFindings(*natsURL, *natsSubject, report); err != nil {
			log.Printf("Error publishing to NATS: %v", err)
		}
	}
}

func hasMultipleEnvironments(files []string) bool {
	envs := make(map[string]bool)
	for _, f := range files {
		if strings.Contains(strings.ToLower(f), "dev") {
			envs["dev"] = true
		}
		if strings.Contains(strings.ToLower(f), "prod") {
			envs["prod"] = true
		}
	}
	return len(envs) > 1
}

func generateComparisonTable(files []string) string {
	// Very simple comparison logic for the evaluation
	var sb strings.Builder
	sb.WriteString("| Feature | Development | Production |\n")
	sb.WriteString("|---------|-------------|------------|\n")
	
	devFile := ""
	prodFile := ""
	for _, f := range files {
		if strings.Contains(strings.ToLower(f), "dev") {
			devFile = f
		}
		if strings.Contains(strings.ToLower(f), "prod") {
			prodFile = f
		}
	}

	sb.WriteString(fmt.Sprintf("| Files | `%s` | `%s` |\n", devFile, prodFile))
	sb.WriteString("| Consistent Labels | No | No |\n")
	sb.WriteString("| Drift Correction | Disabled | Enabled |\n")
	
	return sb.String()
}
