package cleanup

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"cloud.google.com/go/firestore"
	"cloud.google.com/go/storage"
	"github.com/GoogleCloudPlatform/functions-framework-go/functions"
	"google.golang.org/api/iterator"
)

func init() {
	functions.HTTP("Cleanup", Cleanup)
}

// CleanupRequest represents the cleanup configuration
type CleanupRequest struct {
	CleanupTypes  []string `json:"cleanup_types"`
	RetentionDays int      `json:"retention_days"`
	DryRun        bool     `json:"dry_run"`
}

// CleanupResponse represents the cleanup results
type CleanupResponse struct {
	Status       string                 `json:"status"`
	Timestamp    string                 `json:"timestamp"`
	CleanupTypes []string               `json:"cleanup_types"`
	Results      map[string]interface{} `json:"results"`
	Errors       []string               `json:"errors"`
}

// Cleanup performs cleanup of old data
func Cleanup(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()

	// Parse request
	var req CleanupRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, fmt.Sprintf("Invalid request: %v", err), http.StatusBadRequest)
		return
	}

	// Set defaults
	if req.RetentionDays == 0 {
		retentionStr := os.Getenv("RETENTION_DAYS")
		if retentionStr != "" {
			req.RetentionDays, _ = strconv.Atoi(retentionStr)
		} else {
			req.RetentionDays = 30
		}
	}

	if len(req.CleanupTypes) == 0 {
		req.CleanupTypes = []string{"sessions", "temp_files", "old_logs"}
	}

	// Initialize response
	resp := CleanupResponse{
		Status:       "started",
		Timestamp:    time.Now().UTC().Format(time.RFC3339),
		CleanupTypes: req.CleanupTypes,
		Results:      make(map[string]interface{}),
		Errors:       []string{},
	}

	// Perform cleanup for each type
	for _, cleanupType := range req.CleanupTypes {
		switch cleanupType {
		case "sessions":
			if err := cleanupSessions(ctx, req.RetentionDays, req.DryRun, &resp); err != nil {
				resp.Errors = append(resp.Errors, fmt.Sprintf("Session cleanup failed: %v", err))
			}
		case "temp_files":
			if err := cleanupTempFiles(ctx, req.RetentionDays, req.DryRun, &resp); err != nil {
				resp.Errors = append(resp.Errors, fmt.Sprintf("Temp file cleanup failed: %v", err))
			}
		case "old_logs":
			if err := cleanupOldLogs(ctx, req.RetentionDays, req.DryRun, &resp); err != nil {
				resp.Errors = append(resp.Errors, fmt.Sprintf("Log cleanup failed: %v", err))
			}
		default:
			resp.Errors = append(resp.Errors, fmt.Sprintf("Unknown cleanup type: %s", cleanupType))
		}
	}

	// Set final status
	if len(resp.Errors) > 0 {
		resp.Status = "completed_with_errors"
	} else {
		resp.Status = "completed"
	}

	// Return response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// cleanupSessions removes expired sessions from Firestore
func cleanupSessions(ctx context.Context, retentionDays int, dryRun bool, resp *CleanupResponse) error {
	projectID := os.Getenv("FIRESTORE_PROJECT")
	if projectID == "" {
		projectID = os.Getenv("PROJECT_ID")
	}

	client, err := firestore.NewClient(ctx, projectID)
	if err != nil {
		return fmt.Errorf("failed to create Firestore client: %v", err)
	}
	defer client.Close()

	cutoffTime := time.Now().AddDate(0, 0, -retentionDays)
	sessionsDeleted := 0

	// Query old sessions
	iter := client.Collection("sessions").
		Where("updated_at", "<", cutoffTime).
		Documents(ctx)

	batch := client.Batch()
	batchSize := 0

	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return fmt.Errorf("failed to iterate sessions: %v", err)
		}

		if !dryRun {
			batch.Delete(doc.Ref)
			batchSize++

			// Commit batch every 500 documents
			if batchSize >= 500 {
				if _, err := batch.Commit(ctx); err != nil {
					return fmt.Errorf("failed to commit batch: %v", err)
				}
				batch = client.Batch()
				batchSize = 0
			}
		}

		sessionsDeleted++
	}

	// Commit final batch
	if batchSize > 0 && !dryRun {
		if _, err := batch.Commit(ctx); err != nil {
			return fmt.Errorf("failed to commit final batch: %v", err)
		}
	}

	resp.Results["sessions"] = map[string]interface{}{
		"deleted":     sessionsDeleted,
		"dry_run":     dryRun,
		"cutoff_date": cutoffTime.Format(time.RFC3339),
	}

	log.Printf("Session cleanup: deleted %d sessions older than %s (dry_run: %v)",
		sessionsDeleted, cutoffTime.Format(time.RFC3339), dryRun)

	return nil
}

// cleanupTempFiles removes old temporary files from Cloud Storage
func cleanupTempFiles(ctx context.Context, retentionDays int, dryRun bool, resp *CleanupResponse) error {
	bucketName := os.Getenv("STORAGE_BUCKET")
	if bucketName == "" {
		return fmt.Errorf("STORAGE_BUCKET not configured")
	}

	client, err := storage.NewClient(ctx)
	if err != nil {
		return fmt.Errorf("failed to create Storage client: %v", err)
	}
	defer client.Close()

	bucket := client.Bucket(bucketName)
	cutoffTime := time.Now().AddDate(0, 0, -retentionDays)
	filesDeleted := 0

	// List and delete old temp files
	iter := bucket.Objects(ctx, &storage.Query{
		Prefix: "temp/",
	})

	for {
		attrs, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return fmt.Errorf("failed to iterate objects: %v", err)
		}

		if attrs.Created.Before(cutoffTime) {
			if !dryRun {
				if err := bucket.Object(attrs.Name).Delete(ctx); err != nil {
					log.Printf("Failed to delete object %s: %v", attrs.Name, err)
					continue
				}
			}
			filesDeleted++
		}
	}

	resp.Results["temp_files"] = map[string]interface{}{
		"deleted":     filesDeleted,
		"dry_run":     dryRun,
		"cutoff_date": cutoffTime.Format(time.RFC3339),
	}

	log.Printf("Temp file cleanup: deleted %d files older than %s (dry_run: %v)",
		filesDeleted, cutoffTime.Format(time.RFC3339), dryRun)

	return nil
}

// cleanupOldLogs removes old log entries from Cloud Storage
func cleanupOldLogs(ctx context.Context, retentionDays int, dryRun bool, resp *CleanupResponse) error {
	bucketName := os.Getenv("STORAGE_BUCKET")
	if bucketName == "" {
		return fmt.Errorf("STORAGE_BUCKET not configured")
	}

	client, err := storage.NewClient(ctx)
	if err != nil {
		return fmt.Errorf("failed to create Storage client: %v", err)
	}
	defer client.Close()

	bucket := client.Bucket(bucketName)
	cutoffTime := time.Now().AddDate(0, 0, -retentionDays)
	logsDeleted := 0

	// List and delete old logs
	iter := bucket.Objects(ctx, &storage.Query{
		Prefix: "logs/",
	})

	for {
		attrs, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return fmt.Errorf("failed to iterate objects: %v", err)
		}

		if attrs.Created.Before(cutoffTime) {
			if !dryRun {
				if err := bucket.Object(attrs.Name).Delete(ctx); err != nil {
					log.Printf("Failed to delete log %s: %v", attrs.Name, err)
					continue
				}
			}
			logsDeleted++
		}
	}

	resp.Results["old_logs"] = map[string]interface{}{
		"deleted":     logsDeleted,
		"dry_run":     dryRun,
		"cutoff_date": cutoffTime.Format(time.RFC3339),
	}

	log.Printf("Log cleanup: deleted %d logs older than %s (dry_run: %v)",
		logsDeleted, cutoffTime.Format(time.RFC3339), dryRun)

	return nil
}
