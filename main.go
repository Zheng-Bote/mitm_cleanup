/**
 * SPDX-FileComment: DB Cleanup Maintenance Job
 * SPDX-FileType: SOURCE
 * SPDX-FileContributor: ZHENG Robert
 * SPDX-FileCopyrightText: 2026 ZHENG Robert
 * SPDX-License-Identifier: Apache-2.0
 */

package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"os"
	"strconv"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	appName        = "MitM Cleanup Job"
	appDescription = "Maintains database health by pruning old records."
	version        = "1.0.0"
)

// TargetDBConfig defines parameters for the MitM target database
type TargetDBConfig struct {
	Host     string `json:"host"`
	Port     int    `json:"port"`
	User     string `json:"user"`
	Password string `json:"password"`
	Database string `json:"database"`
	DSN      string `json:"dsn"`
}

// CleanupArgs defines retention policies
type CleanupArgs struct {
	TargetFragmentsRetentionDays int `json:"target_fragments_retention_days"`
	RawIngestionOrphanDays       int `json:"raw_ingestion_orphan_days"`
	AdminAuditLogsRetentionDays  int `json:"admin_audit_logs_retention_days"`
	JobAuditLogsRetentionDays    int `json:"job_audit_logs_retention_days"`
	SystemLogsRetentionDays      int `json:"system_logs_retention_days"`
	JobStatusEventsRetentionDays int `json:"job_status_events_retention_days"`
	TransformationErrorsRetentionDays int `json:"transformation_errors_retention_days"`
}

// StatusEvent is sent to the scheduler Unix socket
type StatusEvent struct {
	RunID     int    `json:"run_id"`
	Type      string `json:"type"`
	Component string `json:"component"`
	Status    string `json:"status"`
	Message   string `json:"message"`
	Progress  int    `json:"progress"`
}

type IPCClient struct {
	SocketPath string
	RunID      int
	Component  string
}

func (c *IPCClient) SendEvent(status, message string, progress int) {
	if c == nil || c.SocketPath == "" {
		return
	}
	conn, err := net.Dial("unix", c.SocketPath)
	if err != nil {
		log.Printf("[IPC ERROR] Failed to connect to scheduler socket: %v", err)
		return
	}
	defer conn.Close()

	event := StatusEvent{
		RunID:    c.RunID,
		Type:     "status",
		Status:   status,
		Message:  message,
		Progress: progress,
	}
	data, _ := json.Marshal(event)
	_, _ = conn.Write(append(data, '\n'))
}

func (c *IPCClient) SendAudit(message string) {
	if c == nil || c.SocketPath == "" {
		return
	}
	conn, err := net.Dial("unix", c.SocketPath)
	if err != nil {
		log.Printf("[IPC ERROR] Failed to connect to scheduler socket: %v", err)
		return
	}
	defer conn.Close()

	event := StatusEvent{
		RunID:     c.RunID,
		Type:      "audit",
		Component: c.Component,
		Message:   message,
	}
	data, _ := json.Marshal(event)
	_, _ = conn.Write(append(data, '\n'))
}

func main() {
	var ipc *IPCClient
	runIDStr := os.Getenv("RUN_ID")
	socketPath := os.Getenv("SCHEDULER_SOCKET_PATH")
	if runIDStr != "" && socketPath != "" {
		runID, err := strconv.Atoi(runIDStr)
		if err == nil {
			ipc = &IPCClient{
				SocketPath: socketPath,
				RunID:      runID,
				Component:  "mitm_cleanup",
			}
		}
	}

	ipc.SendEvent("started", fmt.Sprintf("%s (%s) started", appName, version), 0)
	ipc.SendAudit(fmt.Sprintf("%s (%s) started", appName, version))

	var targetCfg TargetDBConfig
	targetCfg.Host = os.Getenv("MITM_DB_HOST")
	if portStr := os.Getenv("MITM_DB_PORT"); portStr != "" {
		targetCfg.Port, _ = strconv.Atoi(portStr)
	}
	targetCfg.User = os.Getenv("MITM_DB_USER")
	targetCfg.Password = os.Getenv("MITM_DB_PASSWORD")
	targetCfg.Database = os.Getenv("MITM_DB_NAME")

	if targetCfg.Host == "" {
		jsonConfig := os.Getenv("MITM_DB_CONFIG_JSON")
		if jsonConfig != "" {
			if err := json.Unmarshal([]byte(jsonConfig), &targetCfg); err != nil {
				ipc.SendEvent("failed", fmt.Sprintf("Failed to parse MitM database JSON config: %v", err), 0)
				log.Fatalf("Failed to parse config: %v", err)
			}
		} else {
			ipc.SendEvent("failed", "MitM database configuration missing in ENV", 0)
			log.Fatal("MitM database credentials not found")
		}
	}

	args := CleanupArgs{
		TargetFragmentsRetentionDays:      7,
		RawIngestionOrphanDays:            14,
		AdminAuditLogsRetentionDays:       90,
		JobAuditLogsRetentionDays:         30,
		SystemLogsRetentionDays:           30,
		JobStatusEventsRetentionDays:      14,
		TransformationErrorsRetentionDays: 30,
	}

	if len(os.Args) >= 2 {
		if err := json.Unmarshal([]byte(os.Args[1]), &args); err != nil {
			log.Printf("Warning: Failed to parse collector arguments from os.Args[1]: %v. Using defaults.", err)
		}
	}

	var mitmDSN string
	if targetCfg.DSN != "" {
		mitmDSN = targetCfg.DSN
	} else {
		mitmDSN = fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=disable",
			targetCfg.User, targetCfg.Password, targetCfg.Host, targetCfg.Port, targetCfg.Database)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	pool, err := pgxpool.New(ctx, mitmDSN)
	if err != nil {
		ipc.SendEvent("failed", fmt.Sprintf("Failed to connect to MitM database: %v", err), 0)
		log.Fatalf("Failed to connect: %v", err)
	}
	defer pool.Close()

	ipc.SendEvent("processing", "Connected to MitM database. Starting cleanup...", 10)

	totalDeleted := 0

	// 1. Clean Target Fragments
	res, err := pool.Exec(ctx, "DELETE FROM target_fragments WHERE status = 'delivered' AND created_at < NOW() - INTERVAL '1 day' * $1", args.TargetFragmentsRetentionDays)
	if err != nil {
		log.Printf("Error cleaning target_fragments: %v", err)
	} else {
		count := res.RowsAffected()
		totalDeleted += int(count)
		ipc.SendAudit(fmt.Sprintf("Deleted %d delivered target fragments older than %d days.", count, args.TargetFragmentsRetentionDays))
	}

	ipc.SendEvent("processing", "Cleaned target fragments.", 30)

	// 2. Clean Orphaned Raw Ingestion
	// E.g., status pending and no update in 'x' days
	res, err = pool.Exec(ctx, "DELETE FROM raw_ingestion WHERE status IN ('pending', 'delivered') AND created_at < NOW() - INTERVAL '1 day' * $1", args.RawIngestionOrphanDays)
	if err != nil {
		log.Printf("Error cleaning raw_ingestion: %v", err)
	} else {
		count := res.RowsAffected()
		totalDeleted += int(count)
		ipc.SendAudit(fmt.Sprintf("Deleted %d raw ingestion fragments older than %d days.", count, args.RawIngestionOrphanDays))
	}

	ipc.SendEvent("processing", "Cleaned raw fragments.", 50)

	// 3. Clean Audit Logs
	res, err = pool.Exec(ctx, "DELETE FROM job_audit_logs WHERE created_at < NOW() - INTERVAL '1 day' * $1", args.JobAuditLogsRetentionDays)
	if err == nil {
		totalDeleted += int(res.RowsAffected())
	}
	res, err = pool.Exec(ctx, "DELETE FROM admin_audit_logs WHERE created_at < NOW() - INTERVAL '1 day' * $1", args.AdminAuditLogsRetentionDays)
	if err == nil {
		totalDeleted += int(res.RowsAffected())
	}

	ipc.SendEvent("processing", "Cleaned audit logs.", 60)

	// 4. Clean System Logs
	res, err = pool.Exec(ctx, "DELETE FROM system_logs WHERE created_at < NOW() - INTERVAL '1 day' * $1", args.SystemLogsRetentionDays)
	if err == nil {
		totalDeleted += int(res.RowsAffected())
	}
	
	// 5. Clean Job Status Events
	res, err = pool.Exec(ctx, "DELETE FROM job_status_events WHERE created_at < NOW() - INTERVAL '1 day' * $1", args.JobStatusEventsRetentionDays)
	if err == nil {
		totalDeleted += int(res.RowsAffected())
	}
	
	// 6. Clean Transformation Errors
	res, err = pool.Exec(ctx, "DELETE FROM transformation_errors WHERE created_at < NOW() - INTERVAL '1 day' * $1", args.TransformationErrorsRetentionDays)
	if err == nil {
		totalDeleted += int(res.RowsAffected())
	}

	ipc.SendAudit(fmt.Sprintf("%s (%s) finished", appName, version))
	ipc.SendEvent("finished", fmt.Sprintf("Cleanup complete. Removed %d outdated records in total.", totalDeleted), 100)
	log.Printf("Cleanup complete. Deleted %d rows.", totalDeleted)
}
