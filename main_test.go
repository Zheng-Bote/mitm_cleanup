package main

import (
	"encoding/json"
	"testing"
)

func TestParseCleanupArgs(t *testing.T) {
	jsonStr := `{"target_fragments_retention_days": 10, "raw_ingestion_orphan_days": 20}`
	
	args := CleanupArgs{
		TargetFragmentsRetentionDays:      7,
		RawIngestionOrphanDays:            14,
		AdminAuditLogsRetentionDays:       90,
		JobAuditLogsRetentionDays:         30,
		SystemLogsRetentionDays:           30,
		JobStatusEventsRetentionDays:      14,
		TransformationErrorsRetentionDays: 30,
	}
	
	if err := json.Unmarshal([]byte(jsonStr), &args); err != nil {
		t.Fatalf("Failed to parse args: %v", err)
	}
	
	if args.TargetFragmentsRetentionDays != 10 {
		t.Errorf("Expected 10, got %d", args.TargetFragmentsRetentionDays)
	}
	
	if args.RawIngestionOrphanDays != 20 {
		t.Errorf("Expected 20, got %d", args.RawIngestionOrphanDays)
	}
	
	if args.AdminAuditLogsRetentionDays != 90 {
		t.Errorf("Expected defaults to be kept, got %d", args.AdminAuditLogsRetentionDays)
	}
}
