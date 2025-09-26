# Jaqen GUI - Modern Interface

## Overview

Jaqen now includes a modern, sleek graphical user interface built with Fyne that makes the Football Manager face management process much more user-friendly.

## Features

### 🎨 Modern Design
- Clean, intuitive interface with card-based layout
- Responsive design that works on different screen sizes
- Native look and feel across platforms

### 📁 Easy File Selection
- **Native file dialogs** - Uses your operating system's native file browser for the most familiar experience
- **File type filtering** - Automatically filters for correct file extensions (XML, RTF, TOML)
- **Directory selection** - Native folder browser for image directories
- **Path validation** - Ensures required files exist before processing
- **Cross-platform** - Native dialogs work consistently on Windows, macOS, and Linux

### ⚙️ Configuration Options
- **Preserve existing mappings** - checkbox to keep current face assignments
- **Allow duplicate images** - option to reuse images for multiple players
- **Football Manager version** - dropdown selection for different FM versions
- **Config file support** - optional TOML configuration file

### 📊 Real-time Progress
- **Progress bar** - visual indication of processing status
- **Status updates** - detailed text showing current operation
- **Non-blocking UI** - interface remains responsive during processing

### 🚀 One-Click Processing
- **Process Files** button - starts the entire face mapping process
- **Format Config** button - formats TOML configuration files
- **Error handling** - user-friendly error messages and validation

## How to Use

### Building and Running

```bash
# Build (GUI mode by default)
make build

# Run GUI (default mode)
./jaqen

# Or run directly
go run .
```

### CLI Mode Alternative

If you prefer the command-line interface:

```bash
# Run CLI mode
./jaqen cli

# Or build CLI-only version
make build-cli
./jaqen cli
```

### GUI Workflow

1. **Select Files**
   - Click "Browse..." next to XML Config File to select your `config.xml`
   - Click "Browse..." next to RTF Player File to select your `newgen.rtf`
   - Click "Browse..." next to Image Directory to select your facepack folder

2. **Configure Options**
   - Check "Preserve existing mappings" if you want to keep current assignments
   - Check "Allow duplicate images" if you want to reuse images
   - Select your Football Manager version from the dropdown

3. **Process Files**
   - Click "Process Files" to start the face mapping
   - Watch the progress bar and status updates
   - Wait for completion confirmation

4. **Optional: Format Config**
   - If you have a TOML config file, use "Format Config" to clean it up

## Technical Details

### Build Tags
The project uses Go build tags to conditionally compile:
- **Default version**: `go build .` (includes GUI, launches GUI by default)
- **CLI-only version**: `go build -tags cli .` (CLI-only, no GUI dependencies)

### Dependencies
- **Fyne v2.6.3** - Modern Go GUI framework
- **sqweek/dialog** - Native file dialog library for true OS integration
- **Cross-platform** - Works on Windows, macOS, and Linux
- **Native performance** - Uses system-native widgets and file dialogs

### File Structure
```
cmd/
├── gui.go          # GUI implementation
├── root.go         # CLI commands (includes GUI command)
└── format.go       # Config formatting

main.go             # CLI entry point
main_gui.go         # GUI entry point (build tag)
```

## Benefits Over CLI

- **No command-line knowledge required** - perfect for casual users
- **Visual feedback** - see exactly what's happening during processing
- **Error prevention** - file validation before processing starts
- **Easy configuration** - checkboxes and dropdowns instead of flags
- **Progress tracking** - know how long operations will take
- **User-friendly** - intuitive interface for Football Manager players

The GUI maintains all the power and flexibility of the CLI version while making it accessible to users who prefer graphical interfaces.
