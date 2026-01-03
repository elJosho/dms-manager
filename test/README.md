# Mock DMS Server for Testing

A lightweight Go-based mock DMS API server for testing the dms-manager without AWS credentials.

## Quick Start

### Terminal 1: Start Mock Server

```bash
cd test/mock-server
go run main.go
```

The server starts on `http://localhost:4566`

### Terminal 2: Test DMS Manager

```bash
# Set environment
source test/mock-env.sh

# Run commands
./dms-manager list
./dms-manager describe mock-task-1
./dms-manager start mock-task-2
./dms-manager stop mock-task-1
./dms-manager tui
```

## One-Command Test

```bash
./test/quick-test.sh
```

This script:
1. Starts the mock server
2. Builds dms-manager
3. Runs a test
4. Stops the mock server

## Mock Data

The server provides **3 mock tasks**:

### mock-task-1
- Status: `running`
- Type: `full-load`
- Progress: 75%
- Tables: 15 loaded, 3 loading, 2 queued

### mock-task-2
- Status: `stopped`
- Type: `full-load-and-cdc`
- Progress: 100%
- Tables: 25 loaded

### mock-task-3
- Status: `failed`
- Type: `cdc`
- Progress: 45%
- Error: "Connection timeout to source database"

## Supported Operations

The mock server implements these DMS API operations:

- ✅ **DescribeReplicationTasks** - List all tasks or filter by ARN
- ✅ **StartReplicationTask** - Changes task status to "starting"
- ✅ **StopReplicationTask** - Changes task status to "stopping"
- ✅ Mock endpoints and replication instances

## Features

- **No dependencies** - Pure Go, no Docker required
- **Fast startup** - Ready in seconds
- **Stateful** - Task status changes persist while server runs
- **AWS SDK compatible** - Uses same endpoint mechanism
- **Logs requests** - See what the DMS manager is doing

## Testing Different Scenarios

### Test with different statuses

```bash
# Start with mock-task-2 (stopped)
./dms-manager start mock-task-2

# Stop mock-task-1 (running)
./dms-manager stop mock-task-1

# Try to start mock-task-3 (failed)
./dms-manager start mock-task-3
```

### Test multi-task operations

```bash
./dms-manager start mock-task-1 mock-task-2 mock-task-3
./dms-manager stop mock-task-1 mock-task-2
```

### Test TUI

```bash
./dms-manager tui
# Use 'r' to resume, 'l' to reload, 's' to start, 'x' to stop
```

## Make Commands

```bash
make mock-server    # Build mock server binary
make mock-start     # Start mock server in background
make mock-stop      # Stop mock server
make test           # Run test (requires mock server running)
```

## Server Details

- **Port**: 4566 (same as LocalStack for compatibility)
- **Protocol**: HTTP with AWS JSON protocol
- **Authentication**: Accepts any credentials (test/test)
- **Region**: Configured via AWS_DEFAULT_REGION

## Limitations

This is a **simple mock** for testing:
- State resets on restart
- No persistent storage
- Simplified API responses
- No validation of all parameters
- Immediate state changes (no delays)

Perfect for development and testing the dms-manager application!

## Extending the Mock

To add more mock tasks, edit `test/mock-server/main.go` and add entries to the `tasks` map.

To support more API operations, add handlers in the `handleDMSRequest` function.
