package dms

import "time"

// Task represents a DMS replication task
type Task struct {
	ARN                    string
	Name                   string
	Status                 string
	ReplicationInstanceARN string
	SourceEndpointARN      string
	TargetEndpointARN      string
	MigrationType          string
	TableMappings          string
	CreatedAt              *time.Time
	StartedAt              *time.Time
	StoppedAt              *time.Time
	LastFailureMessage     string
	ReplicationTaskStats   *TaskStats
}

// TaskStats contains statistics about a replication task
type TaskStats struct {
	FullLoadProgressPercent int32
	ElapsedTimeMillis       int64
	TablesLoaded            int32
	TablesLoading           int32
	TablesQueued            int32
	TablesErrored           int32
	StopReason              string
}

// TableStatistic represents statistics for a single table in a replication task
type TableStatistic struct {
	SchemaName      string
	TableName       string
	Inserts         int64
	Deletes         int64
	Updates         int64
	Ddls            int64
	FullLoadRows    int64
	LastUtctime     *time.Time
	ValidationState string
}

// TaskOperation represents the result of an operation on a task
type TaskOperation struct {
	TaskARN string
	Success bool
	Error   error
	Message string
}
