# File Watcher Logic Plan

## Overview
Extend the server mode to automatically detect file changes, reload configurations, and regenerate relevant instances without manual intervention.

## Core Requirements

### 1. File Watching
- **Watch Config Files**: Monitor all loaded config.toml files for changes
- **Watch SCAD Files**: Monitor the main .scad file referenced in each config
- **Watch Template Files**: Monitor any template files used in the SCAD generation
- **Watch Library Files**: Monitor any library files imported by SCAD files

### 2. Change Detection & Analysis
- **Config Changes**: Detect when config.toml is modified
  - Parse new config and compare with cached version
  - Identify which parameters changed
  - Determine which instances need regeneration
- **SCAD Changes**: Detect when .scad file is modified
  - All instances using this SCAD file need regeneration
  - May need to re-parse SCAD to understand dependencies
- **Template Changes**: Detect when template files change
  - Regenerate all instances using affected templates
- **Library Changes**: Detect when imported libraries change
  - Regenerate all instances that import the changed library

### 3. Smart Regeneration
- **Incremental Updates**: Only regenerate instances affected by changes (use a-h/templ and htmx polling/oob updates)
- **Batch Processing**: Group multiple changes to avoid redundant work
- **Dependency Tracking**: Understand which instances depend on which files
- **Progress Tracking**: Show regeneration progress in web UI

## Technical Implementation

### File Watcher Setup
```go
// Use fsnotify for cross-platform file watching
import "github.com/fsnotify/fsnotify"

type FileWatcher struct {
    watcher *fsnotify.Watcher
    configs map[string]*ConfigInfo
    scadFiles map[string][]string  // scad file -> list of configs using it
    templates map[string][]string  // template -> list of configs using it
    libraries map[string][]string  // library -> list of configs importing it
}
```

### Configuration Tracking
```go
type ConfigInfo struct {
    Path string
    Config *models.Config
    LastModified time.Time
    Instances []models.InstanceConfig
    Dependencies FileDependencies
}

type FileDependencies struct {
    SCADFile string
    TemplateFiles []string
    LibraryFiles []string
}
```

### Change Detection Logic
1. **Config File Changed**:
   - Parse new config
   - Compare with cached version
   - Identify parameter changes
   - Queue affected instances for regeneration

2. **SCAD File Changed**:
   - Find all configs using this SCAD file
   - Queue all instances from those configs
   - Re-parse SCAD to update dependency tracking

3. **Template/Library Changed**:
   - Find all configs depending on this file
   - Queue all instances from those configs

### Regeneration Queue
```go
type RegenerationJob struct {
    ConfigPath string
    InstanceIDs []string
    Reason string
    Priority int
    Created time.Time
}

type RegenerationQueue struct {
    jobs chan RegenerationJob
    workers int
    processing map[string]bool
}
```

## Web UI Integration

### HTMX + Templ Architecture
- **Polling**: Use htmx polling to check for updates every 2-3 seconds
- **OOB Updates**: Send out-of-band updates for immediate feedback
- **Templ Components**: Modular components for different UI sections
- **Progressive Enhancement**: Works without JavaScript, enhanced with htmx

### Progress Updates
- Extend existing progress endpoint to show file watching status
- Show which files are being watched
- Display regeneration queue status
- Real-time updates via htmx polling and out-of-band (OOB) updates
- Use a-h/templ components for dynamic rendering

### Configuration Display
- Show file watching status in config view
- Display last modification times
- Show dependency relationships
- Allow manual trigger of regeneration
- Use htmx for interactive controls (pause/resume watching, manual triggers)
- Templ components for status indicators and progress bars

### HTMX Endpoints
- `GET /api/watcher/status` - Current watching status (polled)
- `POST /api/watcher/pause` - Pause file watching
- `POST /api/watcher/resume` - Resume file watching
- `GET /api/queue/status` - Regeneration queue status (polled)
- `POST /api/queue/trigger/{config}` - Manual trigger for specific config
- `GET /api/queue/job/{id}` - Individual job status (OOB updates)

## Important Considerations

### Performance
- **Debouncing**: Wait for file writes to complete before processing
- **Batch Processing**: Group multiple file changes to avoid redundant work
- **Resource Limits**: Limit concurrent regeneration jobs
- **Memory Management**: Cache configs but avoid memory leaks

### Reliability
- **Error Handling**: Handle file system errors gracefully
- **Recovery**: Re-establish watches after errors
- **Validation**: Validate configs before regeneration
- **Rollback**: Keep previous results if regeneration fails

### User Experience
- **Non-blocking**: File watching shouldn't block web UI
- **Feedback**: Clear indication of what's happening
- **Control**: Allow users to pause/resume watching
- **History**: Track what was regenerated and when

## Implementation Phases

### Phase 1: Basic File Watching
- Set up fsnotify for config files
- Basic change detection and logging
- Manual trigger for regeneration

### Phase 2: Smart Change Detection
- Parse and compare config changes
- Track SCAD file dependencies
- Implement incremental regeneration

### Phase 3: Advanced Features
- Template and library watching
- Dependency graph management
- Advanced UI integration

### Phase 4: Optimization
- Performance tuning
- Memory optimization
- Advanced error handling

## File Structure Changes

### New Files
- `server/file_watcher.go` - Core file watching logic
- `server/regeneration_queue.go` - Job queue management
- `server/dependency_tracker.go` - Track file dependencies
- `pkg/templates/watcher.templ` - Templ components for file watching UI
- `pkg/templates/queue.templ` - Templ components for regeneration queue

### Modified Files
- `server/http_server.go` - Integrate file watcher and add htmx endpoints
- `server/progress.go` - Extend progress tracking with templ components
- `pkg/processor.go` - Support for incremental processing
- `pkg/templates/report.templ` - Add file watching status and controls

## Dependencies
- `github.com/fsnotify/fsnotify/v2` - Alternative for newer features
- `github.com/a-h/templ` - Server-side templating (already used)
- `htmx` - Client-side for dynamic updates (via CDN)

## Testing Strategy
- Unit tests for change detection logic
- Integration tests for file watching
- E2E tests for complete workflow
- Performance tests for large file sets 