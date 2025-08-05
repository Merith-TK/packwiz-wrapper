# PackWiz Wrapper Feature Plan & Architecture Document

**Version:** 0.8.0  
**Date:** January 2025  
**Status:** Current Implementation Analysis & Future Planning

---

## Table of Contents
1. [Project Overview](#project-overview)
2. [Current Architecture](#current-architecture)
3. [CLI Implementation Status](#cli-implementation-status)
4. [GUI Implementation Status](#gui-implementation-status)
5. [Critical Issues & Fixes Needed](#critical-issues--fixes-needed)
6. [Planned Features](#planned-features)
7. [Development Guidelines](#development-guidelines)

---

## Project Overview

**PackWiz Wrapper** is an enhanced wrapper around the popular `packwiz` Minecraft modpack tool, providing both CLI enhancements and a user-friendly GUI interface.

### Core Principles
- **CLI Independence**: GUI code must never be integral to CLI operations
- **Code Reuse**: GUI may use CLI functions, but not vice versa
- **Packwiz Compatibility**: All packwiz commands should pass through exactly as expected
- **Directory Awareness**: All operations must respect the selected pack directory

### Target Users
- **CLI Users**: Power users who need quick, scriptable actions
- **GUI Users**: Casual users who prefer visual interfaces for pack management

---

## Current Architecture

### Project Structure
```
packwrap2/
├── cmd/pw/                     # Main executable entry point
├── internal/
│   ├── commands/               # CLI command implementations
│   │   ├── cmd_*.go           # Individual command files
│   │   ├── command.go         # Command registry & framework
│   │   └── template.go        # Command template for developers
│   ├── core/                   # Core business logic (shared)
│   │   └── manager.go         # Pack management operations
│   ├── gui/                    # GUI implementation (conditional build)
│   │   ├── app.go             # Main GUI application
│   │   ├── tab_*.go           # Individual tab implementations
│   │   ├── shared.go          # Shared GUI utilities
│   │   └── terminal.go        # Command execution display
│   ├── packwiz/               # Embedded packwiz integration
│   │   ├── client.go          # Packwiz command passthrough
│   │   └── types.go           # Packwiz data structures
│   └── utils/                 # Shared utilities
└── pkg/packwrap/              # Public interfaces
    └── interfaces.go          # Core interface definitions
```

### Build System
- **Full Build**: `go build -tags gui` (includes GUI)
- **Headless Build**: `go build` (CLI only, smaller binary)
- **Development**: `make dev` (both versions for testing)

---

## CLI Implementation Status

### ✅ Implemented Commands

#### Core Commands
| Command | Aliases | Status | Description |
|---------|---------|--------|-------------|
| `version` | `v`, `--version` | ✅ Complete | Show version information |
| `help` | `h`, `--help` | ✅ Complete | Display help information |
| `gui` | - | ✅ Complete | Launch GUI (if built with gui tag) |

#### Mod Management
| Command | Aliases | Status | Description |
|---------|---------|--------|-------------|
| `mod` | `m` | ✅ Complete | Enhanced mod management with smart URL parsing |
| `modlist` | `list-mods`, `mods` | ✅ Complete | Generate mod lists in various formats |
| `reinstall` | `refresh-mods` | ✅ Complete | Reinstall/refresh all mods |

#### Pack Operations
| Command | Aliases | Status | Description |
|---------|---------|--------|-------------|
| `build` | `export` | ✅ Complete | Export packs to various formats (CF, MR, MMC, etc.) |
| `import` | `load` | ✅ Complete | Import mods from files or URLs |
| `detect` | `detect-url`, `url` | ✅ Complete | Detect URLs and pack information |
| `release` | `changelog` | ✅ Complete | Generate release files and changelogs |

#### Development & Testing
| Command | Aliases | Status | Description |
|---------|---------|--------|-------------|
| `server` | `test-server`, `start` | ✅ Complete | Start development server with current pack |
| `batch` | `multi` | ✅ Complete | Run commands across multiple pack directories |
| `arbitrary` | `exec`, `run` | ✅ Complete | Execute arbitrary commands in pack directories |

### Command Framework Features
- **Function-based registration**: Simple command creation with `CmdCommandName()` functions
- **Alias support**: Multiple names per command
- **Help system**: Integrated short and long help
- **Packwiz passthrough**: Unknown commands automatically passed to packwiz
- **Directory awareness**: All commands respect current pack directory

### Packwiz Integration
```go
// Current passthrough mechanism
func ExecuteSelfCommand(args []string, packDir string) error {
    // Changes to pack directory
    if packDir != "" {
        os.Chdir(packDir)
    }
    
    // Calls integrated packwiz directly
    originalArgs := os.Args
    os.Args = append([]string{os.Args[0]}, args...)
    packwiz.PackwizExecute()
    os.Args = originalArgs
    
    return nil
}
```

---

## GUI Implementation Status

### ✅ Implemented Tabs

#### Welcome Tab (Consolidated)
- **Pack Directory Selection**: Browse and select modpack directories
- **Pack Status Display**: Shows current pack information or errors
- **Pack Information Panel**: Detailed pack metadata in scrollable view
- **Quick Actions**: Refresh, Create New Pack, Refresh Pack
- **Thread-safe Updates**: All UI updates properly wrapped with `fyne.Do()`

#### Other Tabs
| Tab | Status | Features |
|-----|--------|----------|
| **Logs** | ✅ Complete | Real-time command output and application logs |
| **Mods** | 🚧 Partial | Mod listing and management interface |
| **Import/Export** | 🚧 Partial | File import and pack export operations |
| **Server** | 🚧 Partial | Development server management |

### GUI Framework Details
- **Framework**: Fyne v2 with native look and feel
- **App ID**: `xyz.merith.packwrap` (eliminates preferences warnings)
- **Threading**: All UI updates use `fyne.Do()` for thread safety
- **Pack Directory Management**: Global state with callback system for cross-tab updates

---

## Critical Issues & Fixes Needed

### 🚨 HIGH PRIORITY

#### 1. Pack Directory Consistency Issue
**Problem**: GUI commands like "Create New Pack" always execute in `./` instead of the selected pack directory.

**Root Cause**: 
```go
// In tab_welcome.go - this works correctly
RunPwCommand("Create New Pack", []string{
    "init",
    "--name", name,
    // ...
}, packDir)  // ✅ Passes pack directory

// But RunPwCommand in terminal.go ignores working directory
func RunPwCommand(title string, args []string, packDir string) {
    execPath, _ := os.Executable()
    ShowCommandOutput(title, execPath, args, packDir)  // ❌ packDir not used correctly
}
```

**Required Fix**: 
- Update `RunPwCommand` to actually change to the specified directory before execution
- Ensure all GUI operations respect the global pack directory
- Add validation that pack directory is set before operations

#### 2. Packwiz Passthrough Verification
**Problem**: Need to verify that packwiz commands work exactly as expected across all application entry points.

**Testing Required**:
- CLI passthrough: `pw refresh` → `packwiz refresh`
- GUI command execution: GUI button → CLI command → packwiz
- Directory context preservation
- Argument passing accuracy
- Error handling and output capture

### 🔧 MEDIUM PRIORITY

#### 3. Command Execution Feedback
**Current**: Basic terminal output dialog
**Needed**: Real-time progress, better error handling, cancel ability

#### 4. Pack Directory Auto-Detection
**Current**: Manual selection required
**Needed**: Smart detection of current directory pack.toml on startup

---

## Planned Features

### 📋 Short Term (Next Release)

#### CLI Enhancements
- [ ] **Command aliasing system**: Custom user-defined aliases
- [ ] **Configuration file**: Store user preferences and default settings
- [ ] **Pack templates**: Pre-configured pack setups for common modpack types
- [ ] **Dependency management**: Automatic mod dependency resolution

#### GUI Improvements
- [ ] **Pack directory auto-detection**: Smart detection on startup
- [ ] **Mod search and installation**: Built-in mod browser (Modrinth/CurseForge)
- [ ] **Pack comparison**: Compare two packs and show differences
- [ ] **Export wizard**: Step-by-step pack export with validation

### 🚀 Medium Term (Future Releases)

#### Advanced Features
- [ ] **Pack versioning**: Git-based pack version management
- [ ] **Collaborative editing**: Multi-user pack editing with conflict resolution
- [ ] **Pack analytics**: Mod usage statistics and recommendations
- [ ] **Remote pack management**: Edit packs on remote servers

#### Integration Features
- [ ] **CI/CD integration**: GitHub Actions for automated pack building
- [ ] **Discord bot**: Pack status and update notifications
- [ ] **Web interface**: Browser-based pack management
- [ ] **API server**: RESTful API for external tool integration

### 🔬 Long Term (Research)

#### Experimental Ideas
- [ ] **AI-powered mod suggestions**: ML-based mod recommendations
- [ ] **Visual pack editor**: Drag-and-drop mod management
- [ ] **Performance profiling**: Modpack performance analysis tools
- [ ] **Cloud sync**: Pack synchronization across devices

---

## Development Guidelines

### Architecture Rules

#### 1. CLI/GUI Separation
```go
// ✅ CORRECT: GUI uses CLI functions
func (gui *ModTab) addMod(modURL string) {
    // GUI calls CLI command
    commands.CmdMod().Execute([]string{"add", modURL})
}

// ❌ INCORRECT: CLI depends on GUI
func CmdMod() error {
    if gui.IsRunning() {  // ❌ CLI should never check GUI state
        // ...
    }
}
```

#### 2. Pack Directory Handling
```go
// ✅ CORRECT: Always respect pack directory parameter
func ExecuteInPackDir(packDir string, args []string) error {
    if packDir != "" {
        oldDir, _ := os.Getwd()
        defer os.Chdir(oldDir)
        os.Chdir(packDir)
    }
    return executeCommand(args)
}

// ❌ INCORRECT: Ignoring pack directory
func ExecuteCommand(args []string) error {
    // Always uses current directory - wrong!
    return executeCommand(args)
}
```

#### 3. Error Handling
```go
// ✅ CORRECT: Proper error propagation
func (m *Manager) AddMod(packDir, modRef string) error {
    if err := validatePackDir(packDir); err != nil {
        return fmt.Errorf("invalid pack directory: %w", err)
    }
    // ...
}

// ❌ INCORRECT: Silent failures
func (m *Manager) AddMod(packDir, modRef string) error {
    validatePackDir(packDir)  // Ignores error
    // ...
}
```

### Command Development

#### Adding New CLI Commands
1. Create function following the pattern:
```go
func CmdMyCommand() (names []string, shortHelp, longHelp string, execute func([]string) error) {
    return []string{"mycommand", "alias1"},
        "Brief description",
        `Detailed help text`,
        func(args []string) error {
            // Implementation
            return nil
        }
}
```

2. Register in `cmd/pw/main.go`:
```go
registry.RegisterAll(
    commands.CmdMyCommand,  // Just add the function reference
    // ...
)
```

#### Adding GUI Features
1. Ensure thread safety with `fyne.Do()`:
```go
func updateUI() {
    fyne.Do(func() {
        widget.SetText("New text")
    })
}
```

2. Use global pack directory:
```go
func performAction() {
    packDir := GetGlobalPackDir()
    if packDir == "" {
        showError("Please select a pack directory first")
        return
    }
    // Use packDir for operations
}
```

### Testing Strategy

#### Unit Tests
- **CLI Commands**: Test each command in isolation
- **Core Logic**: Test pack operations and data parsing
- **Utilities**: Test helper functions and utilities

#### Integration Tests
- **Packwiz Passthrough**: Verify exact command forwarding
- **Directory Operations**: Test pack directory handling
- **GUI Integration**: Test GUI → CLI → Packwiz chain

#### Manual Testing
- **Cross-platform**: Test on Windows, macOS, Linux
- **Various pack types**: Test with different pack configurations
- **Error scenarios**: Test error handling and recovery

---

## Implementation Priorities

### Phase 1: Fix Critical Issues
1. **Fix pack directory handling in GUI operations**
2. **Verify packwiz passthrough works correctly**
3. **Add comprehensive error handling**
4. **Test all existing features thoroughly**

### Phase 2: Complete Core Features
1. **Finish remaining GUI tabs implementation**
2. **Add pack auto-detection**
3. **Improve command execution feedback**
4. **Add configuration system**

### Phase 3: Advanced Features
1. **Implement pack templates**
2. **Add mod search and browser**
3. **Build pack comparison tools**
4. **Create export wizards**

---

## Conclusion

PackWiz Wrapper has a solid foundation with a clean separation between CLI and GUI components. The immediate focus should be on fixing the pack directory consistency issue and verifying that packwiz passthrough works correctly across all code paths.

The architecture is well-designed for future expansion, with clear interfaces and a modular command system that makes adding new features straightforward.

**Next Steps:**
1. Fix the `RunPwCommand` function to properly respect pack directories
2. Add comprehensive testing for packwiz passthrough
3. Complete the remaining GUI tab implementations
4. Begin work on advanced features like mod search and pack templates

---

*This document should be updated as features are implemented and architectural decisions are made.*
