# Changelog

All notable changes to this project will be documented in this file.

## [v0.2.0] - 2026-05-03

### Added
- `tables` CLI command for listing table statistics for one or more tasks (#13)
- `reload --table schema.table` support for reloading a single table within one or more tasks (#10)
- Wildcard and multi-task support for `describe` (#12)

### Changed
- `list` output now shows total table counts and elapsed load time (#14)
- TUI task list now uses aligned columns and includes table counts and elapsed time (#15)
- Table statistics output uses aligned columns and comma-formatted counts for readability (#7, #8)
- TUI task details and table statistics views are scrollable (#9)

### Fixed
- Running tasks with table errors are now surfaced with clearer status in the TUI (#11)
- Reload no longer requires a task to already be running before it can be issued (#6)
