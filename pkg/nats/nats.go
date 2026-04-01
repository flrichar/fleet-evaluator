package nats

import (
	"encoding/json"
	"fmt"
	"log"

	"fleet-evaluator/pkg/models"
	"github.com/nats-io/nats.go"
)

func PublishFindings(natsURL, subject string, report *models.Report) error {
	nc, err := nats.Connect(natsURL)
	if err != nil {
		return fmt.Errorf("failed to connect to NATS: %w", err)
	}
	defer nc.Close()

	for _, f := range report.Findings {
		data, err := json.Marshal(f)
		if err != nil {
			log.Printf("Error marshaling finding: %v", err)
			continue
		}

		if err := nc.Publish(subject, data); err != nil {
			log.Printf("Error publishing to NATS: %v", err)
		}
	}

	fmt.Printf("Published %d findings to NATS subject: %s\n", len(report.Findings), subject)
	return nil
}
