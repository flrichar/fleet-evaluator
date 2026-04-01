package reporter

import (
	"fmt"
	"os"
	"sort"
	"strings"

	"fleet-evaluator/pkg/models"
)

func GenerateMarkdownReport(report *models.Report, outputPath string) error {
	var sb strings.Builder

	sb.WriteString("# Fleet Configuration Evaluation Report\n\n")
	sb.WriteString(fmt.Sprintf("**Date:** %s\n", report.Date.Format("2006-01-02")) )
	sb.WriteString("**Files evaluated:**\n")
	for _, f := range report.FilesEvaluated {
		sb.WriteString(fmt.Sprintf("- `%s`\n", f))
	}
	sb.WriteString("\n")

	sb.WriteString("## Summary\n\n")
	sb.WriteString("| Severity | Count |\n")
	sb.WriteString("|----------|-------|\n")
	sb.WriteString(fmt.Sprintf("| CRITICAL | %d |\n", report.Summary[models.Critical]))
	sb.WriteString(fmt.Sprintf("| WARNING  | %d |\n", report.Summary[models.Warning]))
	sb.WriteString(fmt.Sprintf("| INFO     | %d |\n", report.Summary[models.Info]))
	sb.WriteString("\n")

	if report.Summary[models.Critical] > 0 {
		sb.WriteString("Overall assessment: Several critical issues were found that need immediate attention.\n")
	} else if report.Summary[models.Warning] > 0 {
		sb.WriteString("Overall assessment: No critical issues found, but some warnings should be reviewed.\n")
	} else {
		sb.WriteString("Overall assessment: Configuration looks solid with no major issues found.\n")
	}
	sb.WriteString("\n")

	sb.WriteString("## Findings\n\n")

	severities := []models.Severity{models.Critical, models.Warning, models.Info}
	for _, sev := range severities {
		if report.Summary[sev] == 0 {
			continue
		}

		sb.WriteString(fmt.Sprintf("### %s\n\n", sev))

		for _, f := range report.Findings {
			if f.Severity != sev {
				continue
			}

			sb.WriteString(fmt.Sprintf("#### [%s] %s\n", f.ID, f.Title))
			sb.WriteString(fmt.Sprintf("- **File:** `%s`\n", f.File))
			sb.WriteString(fmt.Sprintf("- **Location:** `%s`\n", f.Location))
			sb.WriteString(fmt.Sprintf("- **Issue:** %s\n", f.Issue))
			sb.WriteString(fmt.Sprintf("- **Impact:** %s\n", f.Impact))
			sb.WriteString(fmt.Sprintf("- **Recommendation:** %s\n\n", f.Recommendation))
		}
	}

	if report.ComparisonTable != "" {
		sb.WriteString("## Environment Comparison\n\n")
		sb.WriteString(report.ComparisonTable)
		sb.WriteString("\n")
	}

	return os.WriteFile(outputPath, []byte(sb.String()), 0644)
}

func PrintTerminalSummary(report *models.Report) {
	fmt.Println("Fleet Evaluation Complete")
	fmt.Println("========================")
	fmt.Printf("Files analyzed: %d\n", len(report.FilesEvaluated))
	fmt.Printf("CRITICAL: %d | WARNING: %d | INFO: %d\n\n", 
		report.Summary[models.Critical], 
		report.Summary[models.Warning], 
		report.Summary[models.Info])

	if len(report.Findings) > 0 {
		fmt.Println("Top issues:")
		// Sort findings by severity
		sortedFindings := make([]models.Finding, len(report.Findings))
		copy(sortedFindings, report.Findings)
		sort.Slice(sortedFindings, func(i, j int) bool {
			sevMap := map[models.Severity]int{models.Critical: 0, models.Warning: 1, models.Info: 2}
			return sevMap[sortedFindings[i].Severity] < sevMap[sortedFindings[j].Severity]
		})

		count := 0
		for _, f := range sortedFindings {
			fmt.Printf("  [%s] %s\n", f.ID, f.Title)
			count++
			if count >= 5 {
				break
			}
		}
		fmt.Println()
	}

	fmt.Printf("Full report saved to: fleet-eval-report.md\n")
}
