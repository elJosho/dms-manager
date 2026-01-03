package main

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"strings"
	"time"
)

// Mock DMS data structures
type MockTask struct {
	ReplicationTaskArn          string     `json:"ReplicationTaskArn"`
	ReplicationTaskIdentifier   string     `json:"ReplicationTaskIdentifier"`
	Status                      string     `json:"Status"`
	ReplicationInstanceArn      string     `json:"ReplicationInstanceArn"`
	SourceEndpointArn           string     `json:"SourceEndpointArn"`
	TargetEndpointArn           string     `json:"TargetEndpointArn"`
	MigrationType               string     `json:"MigrationType"`
	TableMappings               string     `json:"TableMappings"`
	ReplicationTaskCreationDate int64      `json:"ReplicationTaskCreationDate,omitempty"`
	ReplicationTaskStartDate    int64      `json:"ReplicationTaskStartDate,omitempty"`
	LastFailureMessage          string     `json:"LastFailureMessage,omitempty"`
	ReplicationTaskStats        *TaskStats `json:"ReplicationTaskStats,omitempty"`
}

type TaskStats struct {
	FullLoadProgressPercent int32 `json:"FullLoadProgressPercent"`
	ElapsedTimeMillis       int64 `json:"ElapsedTimeMillis"`
	TablesLoaded            int32 `json:"TablesLoaded"`
	TablesLoading           int32 `json:"TablesLoading"`
	TablesQueued            int32 `json:"TablesQueued"`
	TablesErrored           int32 `json:"TablesErrored"`
}

// Global mock data
var tasks = map[string]*MockTask{
	"mock-task-1": {
		ReplicationTaskArn:          "arn:aws:dms:us-east-1:123456789012:task:mock-task-1",
		ReplicationTaskIdentifier:   "mock-task-1",
		Status:                      "running",
		ReplicationInstanceArn:      "arn:aws:dms:us-east-1:123456789012:rep:mock-instance",
		SourceEndpointArn:           "arn:aws:dms:us-east-1:123456789012:endpoint:mock-source",
		TargetEndpointArn:           "arn:aws:dms:us-east-1:123456789012:endpoint:mock-target",
		MigrationType:               "full-load",
		TableMappings:               `{"rules":[{"rule-type":"selection","rule-id":"1","rule-name":"1","object-locator":{"schema-name":"%","table-name":"%"},"rule-action":"include"}]}`,
		ReplicationTaskCreationDate: epoch(time.Now().Add(-24 * time.Hour)),
		ReplicationTaskStartDate:    epoch(time.Now().Add(-2 * time.Hour)),
		ReplicationTaskStats: &TaskStats{
			FullLoadProgressPercent: 75,
			ElapsedTimeMillis:       7200000,
			TablesLoaded:            15,
			TablesLoading:           3,
			TablesQueued:            2,
			TablesErrored:           0,
		},
	},
	"mock-task-2": {
		ReplicationTaskArn:          "arn:aws:dms:us-east-1:123456789012:task:mock-task-2",
		ReplicationTaskIdentifier:   "mock-task-2",
		Status:                      "stopped",
		ReplicationInstanceArn:      "arn:aws:dms:us-east-1:123456789012:rep:mock-instance",
		SourceEndpointArn:           "arn:aws:dms:us-east-1:123456789012:endpoint:mock-source",
		TargetEndpointArn:           "arn:aws:dms:us-east-1:123456789012:endpoint:mock-target",
		MigrationType:               "full-load-and-cdc",
		TableMappings:               `{"rules":[{"rule-type":"selection","rule-id":"1","rule-name":"1","object-locator":{"schema-name":"public","table-name":"%"},"rule-action":"include"}]}`,
		ReplicationTaskCreationDate: epoch(time.Now().Add(-48 * time.Hour)),
		ReplicationTaskStats: &TaskStats{
			FullLoadProgressPercent: 100,
			TablesLoaded:            25,
			TablesLoading:           0,
			TablesQueued:            0,
			TablesErrored:           0,
		},
	},
	"mock-task-3": {
		ReplicationTaskArn:          "arn:aws:dms:us-east-1:123456789012:task:mock-task-3",
		ReplicationTaskIdentifier:   "mock-task-3",
		Status:                      "failed",
		ReplicationInstanceArn:      "arn:aws:dms:us-east-1:123456789012:rep:mock-instance",
		SourceEndpointArn:           "arn:aws:dms:us-east-1:123456789012:endpoint:mock-source",
		TargetEndpointArn:           "arn:aws:dms:us-east-1:123456789012:endpoint:mock-target",
		MigrationType:               "cdc",
		TableMappings:               `{"rules":[{"rule-type":"selection","rule-id":"1","rule-name":"1","object-locator":{"schema-name":"app","table-name":"%"},"rule-action":"include"}]}`,
		ReplicationTaskCreationDate: epoch(time.Now().Add(-72 * time.Hour)),
		LastFailureMessage:          "Connection timeout to source database",
		ReplicationTaskStats: &TaskStats{
			FullLoadProgressPercent: 45,
			TablesLoaded:            10,
			TablesLoading:           0,
			TablesQueued:            5,
			TablesErrored:           1,
		},
	},
}

// Stats for table statistics mock
type MockTableStat struct {
	SchemaName      string `json:"SchemaName"`
	TableName       string `json:"TableName"`
	Inserts         int64  `json:"Inserts"`
	Deletes         int64  `json:"Deletes"`
	Updates         int64  `json:"Updates"`
	Ddls            int64  `json:"Ddls"`
	FullLoadRows    int64  `json:"FullLoadRows"`
	LastUpdateTime  int64  `json:"LastUpdateTime"`
	ValidationState string `json:"ValidationState"`
}

var tableStats = map[string][]MockTableStat{
	"arn:aws:dms:us-east-1:123456789012:task:mock-task-1": {
		{SchemaName: "public", TableName: "users", Inserts: 100, Deletes: 5, Updates: 20, FullLoadRows: 1000, LastUpdateTime: epoch(time.Now()), ValidationState: "Validated"},
		{SchemaName: "public", TableName: "orders", Inserts: 500, Deletes: 10, Updates: 50, FullLoadRows: 5000, LastUpdateTime: epoch(time.Now()), ValidationState: "Validated"},
		{SchemaName: "public", TableName: "products", Inserts: 50, Deletes: 0, Updates: 10, FullLoadRows: 500, LastUpdateTime: epoch(time.Now()), ValidationState: "Validated"},
	},
	"arn:aws:dms:us-east-1:123456789012:task:mock-task-2": {
		{SchemaName: "public", TableName: "inventory", Inserts: 10, Deletes: 0, Updates: 5, FullLoadRows: 100, LastUpdateTime: epoch(time.Now()), ValidationState: "Pending"},
	},
}

func main() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// Log request
		body, _ := io.ReadAll(r.Body)
		r.Body = io.NopCloser(bytes.NewBuffer(body)) // Restore body for subsequent reads if needed
		log.Printf("Request: %s %s", r.Method, r.URL.Path)

		// Determine action
		var req map[string]interface{}
		json.Unmarshal(body, &req)

		// AWS SDK sends action in header
		action := r.Header.Get("X-Amz-Target")

		switch {
		case strings.Contains(action, "DescribeReplicationTasks"):
			handleDescribeReplicationTasks(w, r)
		case strings.Contains(action, "StartReplicationTask"):
			handleStartReplicationTask(w, r)
		case strings.Contains(action, "StopReplicationTask"):
			handleStopReplicationTask(w, r)
		case strings.Contains(action, "DescribeTableStatistics"):
			handleDescribeTableStatistics(w, r)
		default:
			log.Printf("Unknown action: %s", action)
			http.Error(w, "Unknown action", http.StatusBadRequest)
		}
	})

	port := ":4566"
	log.Printf("🚀 Mock DMS API Server starting on http://localhost%s", port)
	log.Printf("✅ Ready to handle DMS requests")
	log.Printf("   Available mock tasks: mock-task-1, mock-task-2, mock-task-3")
	log.Fatal(http.ListenAndServe(port, nil))
}

func handleDescribeReplicationTasks(w http.ResponseWriter, r *http.Request) {
	// Parse request body
	var req map[string]interface{}
	json.NewDecoder(r.Body).Decode(&req)

	// Get filter if present
	filters := req["Filters"]

	var tasksToReturn []MockTask

	if filters != nil {
		// Filter by ARN if specified
		filterList := filters.([]interface{})
		if len(filterList) > 0 {
			filter := filterList[0].(map[string]interface{})
			if filter["Name"] == "replication-task-arn" {
				values := filter["Values"].([]interface{})
				arn := values[0].(string)

				// Find task by ARN
				for _, task := range tasks {
					if task.ReplicationTaskArn == arn {
						tasksToReturn = append(tasksToReturn, *task)
						break
					}
				}
			}
		}
	} else {
		// Return all tasks
		for _, task := range tasks {
			tasksToReturn = append(tasksToReturn, *task)
		}
	}

	response := map[string]interface{}{
		"ReplicationTasks": tasksToReturn,
	}

	w.Header().Set("Content-Type", "application/x-amz-json-1.1")
	json.NewEncoder(w).Encode(response)
}

func handleStartReplicationTask(w http.ResponseWriter, r *http.Request) {
	var req map[string]interface{}
	json.NewDecoder(r.Body).Decode(&req)

	arn := req["ReplicationTaskArn"].(string)

	// Find and update task status
	for _, task := range tasks {
		if task.ReplicationTaskArn == arn {
			task.Status = "starting"
			task.ReplicationTaskStartDate = epoch(time.Now())

			response := map[string]interface{}{
				"ReplicationTask": task,
			}

			w.Header().Set("Content-Type", "application/x-amz-json-1.1")
			json.NewEncoder(w).Encode(response)

			log.Printf("✅ Started task: %s", task.ReplicationTaskIdentifier)
			return
		}
	}

	// Task not found
	http.Error(w, "Task not found", http.StatusNotFound)
}

func handleStopReplicationTask(w http.ResponseWriter, r *http.Request) {
	var req map[string]interface{}
	json.NewDecoder(r.Body).Decode(&req)

	arn := req["ReplicationTaskArn"].(string)

	// Find and update task status
	for _, task := range tasks {
		if task.ReplicationTaskArn == arn {
			task.Status = "stopping"

			response := map[string]interface{}{
				"ReplicationTask": task,
			}

			w.Header().Set("Content-Type", "application/x-amz-json-1.1")
			json.NewEncoder(w).Encode(response)

			log.Printf("⏹️  Stopped task: %s", task.ReplicationTaskIdentifier)
			return
		}
	}

	// Task not found
	http.Error(w, "Task not found", http.StatusNotFound)
}

func handleDescribeTableStatistics(w http.ResponseWriter, r *http.Request) {
	var req map[string]interface{}
	json.NewDecoder(r.Body).Decode(&req)

	arn := req["ReplicationTaskArn"].(string)

	stats, ok := tableStats[arn]
	if !ok {
		// Return empty list if no stats found (or task doesn't exist in our mock stats map)
		stats = []MockTableStat{}
	}

	response := map[string]interface{}{
		"TableStatistics": stats,
	}

	w.Header().Set("Content-Type", "application/x-amz-json-1.1")
	json.NewEncoder(w).Encode(response)
}

func epoch(t time.Time) int64 {
	// AWS SDK expects Unix epoch in seconds as int64
	return t.Unix()
}
