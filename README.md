# AWS DMS Manager

A powerful command-line tool for managing AWS Database Migration Service (DMS) replication tasks with both CLI and interactive TUI interfaces.

## Features

- 🚀 **Multi-interface**: Both CLI commands and interactive TUI
- 🔄 **Parallel Operations**: Control multiple tasks simultaneously
- 📊 **Real-time Updates**: Auto-refresh task status in TUI
- 🎯 **Profile Support**: Switch between different AWS profiles
- ⚡ **Fast Execution**: Parallel processing for bulk operations
- 🎨 **Beautiful TUI**: Color-coded status and intuitive navigation

## Installation

### Prerequisites

- Go 1.23 or higher
- AWS credentials configured (via `~/.aws/credentials` or environment variables)

### Build from Source

```bash
git clone <repository-url>
cd dms-manager
make build
```

Or install directly:

```bash
go install github.com/eljosho/dms-manager@latest
```

### Cross-Platform Builds

Build executables for different operating systems and architectures:

```bash
# Build for current platform
make build

# Build for specific platforms
make build-linux        # Linux amd64
make build-linux-arm    # Linux arm64
make build-macos        # macOS Intel (amd64)
make build-macos-arm    # macOS Apple Silicon (arm64)
make build-windows      # Windows amd64
make build-windows-arm  # Windows arm64

# Build for ALL platforms at once
make build-all

# Clean build artifacts
make clean
```

All cross-platform binaries are output to the `build/` directory:

```
build/
├── dms-manager-linux-amd64
├── dms-manager-linux-arm64
├── dms-manager-darwin-amd64
├── dms-manager-darwin-arm64
├── dms-manager-windows-amd64.exe
└── dms-manager-windows-arm64.exe
```

## AWS Configuration

Ensure you have AWS credentials configured. You can set them up using:

```bash
aws configure
```

Or manually create `~/.aws/credentials`:

```ini
[default]
aws_access_key_id = YOUR_ACCESS_KEY
aws_secret_access_key = YOUR_SECRET_KEY

[profile staging]
aws_access_key_id = STAGING_ACCESS_KEY
aws_secret_access_key = STAGING_SECRET_KEY
```

## Usage

### CLI Commands

#### List all tasks

```bash
./dms-manager list
./dms-manager list --profile staging --region us-west-2
```

#### Describe task details

```bash
./dms-manager describe <task-name-or-arn>
./dms-manager describe task1 task2 task3
```

#### Start tasks

```bash
# Start a single task
./dms-manager start <task-name-or-arn>

# Start multiple tasks in parallel
./dms-manager start task1 task2 task3

# Start with specific type
./dms-manager start task1 --type resume-processing
```

Start types:
- `start-replication` (default) - Start fresh replication
- `resume-processing` - Resume from where it stopped
- `reload-target` - Reload target tables

#### Stop tasks

```bash
# Stop a single task
./dms-manager stop <task-name-or-arn>

# Stop multiple tasks in parallel
./dms-manager stop task1 task2 task3
```

#### Resume tasks

```bash
# Resume tasks that were stopped (uses resume-processing)
./dms-manager resume task1 task2
```

#### Reload tasks

```bash
# Reload tasks (stop then start with reload-target)
./dms-manager reload task1 task2
```

#### Using Wildcards

You can use `*` or `all` to operate on all tasks. **Important**: Use quotes to prevent shell expansion.

```bash
# Start all tasks
./dms-manager start '*'

# Stop all tasks
./dms-manager stop '*'

# Stop tasks matching a pattern
./dms-manager stop '*-database'
./dms-manager stop 'prod-*'

# Resume all tasks
./dms-manager resume all

# Reload all tasks
./dms-manager reload '*'
```

> **Note**: Always quote the wildcard (`'*'` or `"*"`) to prevent your shell from expanding it to filenames in the current directory.

### Interactive TUI

Launch the interactive terminal interface:

```bash
./dms-manager tui
./dms-manager tui --profile production --region us-east-1
```

#### TUI Keyboard Shortcuts

| Key | Action |
|-----|--------|
| `↑/k` | Move up |
| `↓/j` | Move down |
| `Space` | Select/deselect task |
| `Enter` | View task details |
| `s` | Start selected tasks |
| `x` | Stop selected tasks |
| `r` | Resume selected tasks |
| `l` | Reload selected tasks |
| `c` | Clear all selections |
| `f` | Manually refresh task list |
| `a` | Toggle auto-refresh (default: on) |
| `q` | Quit |

### Global Flags

- `--profile, -p` - AWS profile to use (default: default profile)
- `--region, -r` - AWS region (default: from profile or `AWS_REGION`)

## Examples

### Working with Multiple Tasks

```bash
# List all tasks to see what's available
./dms-manager list --profile production

# Start multiple tasks at once
./dms-manager start task-1 task-2 task-3 --profile production

# Resume stopped tasks
./dms-manager resume task-1 task-2 --profile production

# Reload tasks (full reload)
./dms-manager reload task-1 task-2 --profile production

# Use the TUI for interactive management
./dms-manager tui --profile production
```

### Cross-Profile Management

```bash
# Check staging tasks
./dms-manager list --profile staging

# Start production tasks
./dms-manager start prod-task-1 --profile production

# Use different regions
./dms-manager list --profile dev --region eu-west-1
```

## Example Output

### CLI List Command

```
Region: us-east-1
Profile: production

NAME                          STATUS     TYPE              ARN
────                          ──────     ────              ───
orders-replication            running    full-load-and-cdc arn:aws:dms:us-east-1:123456789:task:ABC123
users-sync                    stopped    cdc               arn:aws:dms:us-east-1:123456789:task:DEF456
inventory-migration           running    full-load         arn:aws:dms:us-east-1:123456789:task:GHI789

Total tasks: 3
```

### CLI Describe Command

```
Task: orders-replication
ARN: arn:aws:dms:us-east-1:123456789:task:ABC123
Status: running
Migration Type: full-load-and-cdc

Endpoints:
  Replication Instance: arn:aws:dms:us-east-1:123456789:rep:INSTANCE1
  Source: arn:aws:dms:us-east-1:123456789:endpoint:SOURCE1
  Target: arn:aws:dms:us-east-1:123456789:endpoint:TARGET1

Created At: 2025-01-15 10:30:00
Started At: 2025-01-15 10:35:22

Statistics:
  Full Load Progress: 100%
  Tables Loaded: 45
  Tables Loading: 0
  Tables Queued: 0
  Tables Errored: 0
  Elapsed Time: 2h 15m
```

### CLI List with Stats

```
Region: us-east-1

Task: orders-replication
Status: running
Type: full-load-and-cdc
ARN: arn:aws:dms:us-east-1:123456789:task:ABC123

Statistics:
  Full Load Progress: 100%
  Tables Loaded: 45
  Tables Loading: 0
  Tables Queued: 0
  Tables Errored: 0
  Elapsed Time: 2h 15m

────────────────────────────────────────────────────────────────────────────────
Task: users-sync
Status: stopped
Type: cdc
ARN: arn:aws:dms:us-east-1:123456789:task:DEF456

Statistics:
  Full Load Progress: 100%
  Tables Loaded: 12
  Tables Loading: 0
  Tables Queued: 0
  Tables Errored: 0
  Stop Reason: Stop Reason FULL_LOAD_ONLY_FINISHED

Total tasks: 2
```

### Interactive TUI

```
AWS DMS Tasks - us-east-1 (Profile: production)

→ [✓] orders-replication - running (full-load-and-cdc)
  [ ] users-sync - stopped (cdc)
  [ ] inventory-migration - running (full-load)

[↑/k] up • [↓/j] down • [space] select • [enter] details • [s] start • [x] stop • [r] resume • [l] reload
[c] clear • [f] refresh • [a] auto-refresh: on • [q] quit
```

### TUI Task Details View

```
Task Details

Name: orders-replication
Status: running
Type: full-load-and-cdc
ARN: arn:aws:dms:us-east-1:123456789:task:ABC123

Endpoints:
  Source: arn:aws:dms:us-east-1:123456789:endpoint:SOURCE1
  Target: arn:aws:dms:us-east-1:123456789:endpoint:TARGET1
  Instance: arn:aws:dms:us-east-1:123456789:rep:INSTANCE1

Created: 2025-01-15 10:30:00
Started: 2025-01-15 10:35:22

Statistics:
  Progress: 100%
  Tables - Loaded: 45, Loading: 0, Queued: 0, Errored: 0
  Elapsed: 2h 15m

Press [t] extended stats: off • [T] table stats • [ESC] back
```

## Architecture

```
dms-manager/
├── main.go                 # Application entry point
├── cmd/                    # CLI commands
│   ├── root.go            # Root command and global flags
│   ├── list.go            # List tasks command
│   ├── describe.go        # Describe tasks command
│   ├── start.go           # Start tasks command
│   ├── stop.go            # Stop tasks command
│   ├── restart.go         # Restart tasks command
│   ├── tui.go             # TUI launcher
│   └── helpers.go         # Shared utilities
├── pkg/dms/               # DMS client library
│   ├── client.go          # AWS SDK wrapper
│   ├── operations.go      # DMS operations
│   └── types.go           # Type definitions
└── internal/tui/          # TUI implementation
    ├── model.go           # Bubble Tea model
    ├── views.go           # View rendering
    ├── commands.go        # Async commands
    └── styles.go          # UI styling
```

## Features in Detail

### Parallel Processing

All multi-task operations (start, stop, restart) use Go goroutines to process tasks in parallel, significantly reducing execution time when managing multiple tasks.

### Auto-Refresh

In TUI mode, the task list automatically refreshes every 5 seconds to show real-time status updates. Toggle this feature with the `a` key.

### Task Selection

The TUI supports multi-selection:
1. Use `Space` to select/deselect individual tasks
2. Perform operations on all selected tasks
3. Use `c` to clear selection
4. If no tasks are selected, operations apply to the task under the cursor

### Error Handling

The application provides clear error messages and handles:
- Network failures
- Invalid credentials
- Missing tasks
- Partial failures in bulk operations

## Testing with Mock Server

You can test the DMS manager locally without AWS credentials using the included mock server.

### Quick Start

```bash
# Terminal 1: Start mock server
make mock-start

# Terminal 2: Test
source test/mock-env.sh
./dms-manager list
./dms-manager tui
```

### One-Command Test

```bash
# Using make
make mock-start && source test/mock-env.sh && make test

# Or using the test script
./test/quick-test.sh
```

### Stop Mock Server

```bash
make mock-stop
```

### See [test/README.md](test/README.md) for complete testing documentation.

## Troubleshooting

### "No credentials found"

Ensure AWS credentials are configured:
```bash
aws configure
```

### "Task not found"

Verify the task name or ARN is correct:
```bash
./dms-manager list
```

### Permission Issues

Ensure your AWS user/role has the necessary DMS permissions:
- `dms:DescribeReplicationTasks`
- `dms:StartReplicationTask`
- `dms:StopReplicationTask`

## Development

### Makefile Commands

Run `make help` to see all available commands:

| Command | Description |
|---------|-------------|
| `make help` | Show all available make targets |
| `make build` | Build for current platform |
| `make build-all` | Build for all platforms (Linux, macOS, Windows) |
| `make build-linux` | Build for Linux amd64 |
| `make build-linux-arm` | Build for Linux arm64 |
| `make build-macos` | Build for macOS Intel (amd64) |
| `make build-macos-arm` | Build for macOS Apple Silicon (arm64) |
| `make build-windows` | Build for Windows amd64 |
| `make build-windows-arm` | Build for Windows arm64 |
| `make clean` | Remove build artifacts |
| `make test` | Run quick test against mock server |
| `make mock-server` | Build the mock DMS server binary |
| `make mock-start` | Start mock server in background |
| `make mock-stop` | Stop mock server |

### Building

```bash
# Build for current platform
make build

# Build for all platforms
make build-all

# Or build directly with Go
go build -o dms-manager
```

### Testing

```bash
# Run Go tests
go test ./...

# Test with mock server
make mock-start      # Start mock server
source test/mock-env.sh
make test            # Run quick test
make mock-stop       # Stop mock server
```

### Dependencies

- AWS SDK for Go v2
- Cobra (CLI framework)
- Bubble Tea (TUI framework)
- Bubbles (TUI components)
- Lipgloss (TUI styling)

## License

MIT License

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.
