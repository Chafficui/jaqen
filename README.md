<div align="center"><pre>
        _____   ____    _______   __
       / /   | / __ \  / ____/ | / /
  __  / / /| |/ / / / / __/ /  |/ /
 / /_/ / ___ / /_/ / / /___/ /|  /  
 \____/_/  |_\___\_\/_____/_/ |_/

Jaqen NewGen Tool
</pre></div>

<div align="center">

**Jaqen NewGen Tool** - Football Manager Face Manager

A modern GUI application for creating and managing image file mappings to face profiles in Football Manager.

[![License: GPL v3](https://img.shields.io/badge/License-GPLv3-blue.svg)](https://www.gnu.org/licenses/gpl-3.0)
[![Version](https://img.shields.io/badge/Version-1.0.0-green.svg)](https://github.com/chafficui/jaqen-newgen-tool/releases)
[![Go Version](https://img.shields.io/badge/Go-1.22+-00ADD8?style=flat&logo=go)](https://golang.org/)
[![Platform](https://img.shields.io/badge/Platform-Windows%20%7C%20macOS%20%7C%20Linux-lightgrey)](https://github.com/chafficui/jaqen-newgen-tool)


## 📺 Watch Tutorial Video

[![▶️ Jaqen NewGen Tool Tutorial Video](https://img.youtube.com/vi/aHnrpfH--ic/maxresdefault.jpg)](https://www.youtube.com/watch?v=aHnrpfH--ic)

*Click the thumbnail above to watch the complete setup tutorial*

</div>

## About

Jaqen NewGen Tool is a modern, cross-platform GUI application for managing Football Manager regen face mappings. It automatically assigns face images to newgen players based on their nationality and ethnic groups.

### Key Features

- **🚀 One-Click Setup** - Auto-distributes views/filters to all FM installations on startup
- **🔄 Auto-Detection** - Automatically finds FM installations and paths
- **📁 File Management** - Auto-generates config.xml and detects RTF files
- **🌍 Cross-Platform** - Works on Windows, macOS, and Linux
- **⚙️ Smart Mapping** - Maps nations to ethnic groups with override support
- **📊 Visual Progress** - Real-time feedback during processing

## Quick Start

### 1. Download and Run

1. **Download** the latest release for your platform from the [Releases page](https://github.com/chafficui/jaqen-newgen-tool/releases)
2. **Extract** the downloaded file to a folder of your choice
3. **Run** the application:
   - **Windows**: Double-click `jaqen-newgen-tool.exe`
   - **macOS**: Double-click `jaqen-newgen-tool` (you may need to allow it in Security & Privacy)
   - **Linux**: Run `./jaqen-newgen-tool` in terminal

### 2. Setup Football Manager

1. **Install Face Pack:**
   - Download your preferred face pack
   - Extract it to your FM graphics folder (e.g., `Documents/Sports Interactive/Football Manager 2024/graphics/`)

2. **Export RTF from Football Manager:**
   - Go to Scouting → Players in Range
   - Import "SCRIPT FACES player search" view (auto-distributed by Jaqen)
   - Apply "is newgen search filter" (auto-distributed by Jaqen)
   - Select all players (Ctrl+A) → Print to text file (Ctrl+P)
   - Save as "newgen.rtf" in your face pack folder

3. **Configure Jaqen NewGen Tool:**
   - Select your face pack directory
   - Choose settings (Preserve, Allow Duplicates, etc.)
   - Click "Assign Face Mappings"

4. **Apply in Football Manager:**
   - Restart Football Manager
   - Newgen faces will use assigned images

### 3. Watch the Tutorial

📺 **[Complete Setup Video Tutorial](https://youtu.be/aHnrpfH--ic)**

## Installation

### Download Pre-built Binaries

Download the latest release for your platform from the [Releases page](https://github.com/chafficui/jaqen-newgen-tool/releases).

**Supported Platforms:**
- **Windows**: `jaqen-newgen-tool-windows-x.x.x.zip`
- **macOS**: `jaqen-newgen-tool-macos-x.x.x.tar.gz` (Intel + Apple Silicon)
- **Linux**: `jaqen-newgen-tool-linux-x.x.x.tar.gz`

### Build from Source (Advanced Users)

For developers or users who want to build from source:

```bash
# Clone repository
git clone https://github.com/chafficui/jaqen-newgen-tool.git
cd jaqen-newgen-tool

# Build for current platform
make build

# Build for all platforms
make build-all

# Run
./jaqen-newgen-tool
```

## Configuration

### Ethnic Group Mapping

The tool automatically maps Football Manager nations to ethnic groups:

| Ethnic Group | Code |
|--------------|------|
| African | African |
| Asian | Asian |
| Caucasian | Caucasian |
| Central European | Central European |
| Eastern European Central Asian | EECA |
| Italian Mediterranean | Italmed |
| Middle East North African | MENA |
| Middle East South Asian | MESA |
| South American Mediterranean | SAMed |
| Scandinavian | Scandinavian |
| South East Asian | Seasian |
| South American | South American |
| Spanish Mediterranean | SpanMed |
| Yugoslav Greek | YugoGreek |

### Custom Mappings

You can override default mappings in the settings:

```toml
[mapping_override]
AFG = 'MESA'  # Afghanistan → Middle East South Asian
ENG = 'Caucasian'  # England → Caucasian
```

## How It Works

1. **Parse RTF File** - Extracts player data (ID, nationality, ethnic group)
2. **Map Nations** - Converts nations to ethnic groups (with override support)
3. **Select Images** - Randomly selects images from appropriate ethnic directories
4. **Generate XML** - Creates Football Manager mapping file
5. **Update Config** - Writes updated config.xml for FM

## Credits

This project is a fork and continuation of the original work:

- **Base Project**: [Jaqen](https://github.com/imfulee/jaqen) by [@imfulee](https://github.com/imfulee)
- **Views & Filters**: [NewGAN-Manager](https://github.com/Maradonna90/NewGAN-Manager) by [@Maradonna90](https://github.com/Maradonna90)
- **Inspiration**: Named after Jaqen H'ghar from Game of Thrones (wall of faces)

## License

This project is licensed under the GPL v3 License - see the [LICENSE](LICENSE) file for details.

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## Support

- 📖 **Documentation**: Check the [docs/](docs/) folder
- 🐛 **Issues**: Report bugs on [GitHub Issues](https://github.com/chafficui/jaqen-newgen-tool/issues)