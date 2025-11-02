# AbletonCLI

```
     _    _     _      _               ____ _     ___ 
    / \  | |__ | | ___| |_ ___  _ __  / ___| |   |_ _|
   / _ \ | '_ \| |/ _ \ __/ _ \| '_ \| |   | |    | | 
  / ___ \| |_) | |  __/ || (_) | | | | |___| |___ | | 
 /_/   \_\_.__/|_|\___|\__\___/|_| |_|\____|_____|___|
```

A powerful command-line interface for managing Ableton Live project files (`.als`).

## Features

- 🔄 **Migrate**: Search and replace strings within Ableton Live project files
- 💾 **Backup**: Create backups of all Ableton project files in a directory
- 🐛 **Debug Mode**: Optional debug logging for troubleshooting
- 🎨 **Colored Output**: Clear, color-coded terminal output
- 🔒 **Safe Operations**: Dry-run mode and automatic cleanup of temporary files

## Installation

### Quick Install

First, download the latest release:

```bash
# Download the binary
curl -L -o abletoncli https://github.com/aubsynth/AbletonCLI/releases/download/v0.1.2/abletoncli
chmod +x abletoncli

# Download the installation script
curl -L -o install.sh https://github.com/aubsynth/AbletonCLI/releases/download/v0.1.2/install.sh
chmod +x install.sh
```
Make sure the install script and abletoncli binary are in the same folder.

Then run the installation script to automatically install AbletonCLI and set up shell completion:

```bash
./install.sh
```

The installer will:
- Install the binary to `/usr/local/bin/abletoncli/`
- Set up shell completion for your shell (bash/zsh)
- Optionally add AbletonCLI to your PATH

### Manual Build

If you prefer to build manually:

```bash
go build -o bin/abletoncli
```

Then add the binary to your PATH or move it to a directory in your PATH.

## Usage

### Migrate Command

Search and replace strings within Ableton Live project files. The migrate command decompresses `.als` files (which are gzipped XML), performs the replacement, and recompresses them.

```bash
abletoncli migrate --replace "OLD_STRING" --with "NEW_STRING" [flags]
```

#### Flags

- `--directory` - Directory to search (default: current directory)
- `--replace` - String to search for (required)
- `--with` - String to replace with (required)
- `--dry-run` - Perform a trial run with no changes made
- `-d, --debug` - Show debug logs

#### Examples

Replace a plugin path in all Ableton projects in the current directory:

```bash
abletoncli migrate --replace "/old/path/plugin" --with "/new/path/plugin"
```

Perform a dry run to see what would be changed:

```bash
abletoncli migrate --replace "old_value" --with "new_value" --dry-run
```

Search in a specific directory with debug output:

```bash
abletoncli migrate --directory ~/Music/Projects --replace "VST2" --with "VST3" -d
```

### Backup Command

Create backups of all Ableton Live project files, preserving the directory structure.

```bash
abletoncli backup [flags]
```

#### Flags

- `--directory` - Directory to backup (default: current directory)
- `--destination` - Backup destination directory (default: `../backup/`)
- `-d, --debug` - Show debug logs

#### Examples

Backup all projects in current directory:

```bash
abletoncli backup
```

Backup a specific directory to a custom location:

```bash
abletoncli backup --directory ~/Music/Projects --destination ~/Backups/AbletonProjects
```

### Common Use Cases

#### Moving from OneDrive to iCloud
```bash
abletoncli migrate --replace "/Users/username/OneDrive/" --with "/Users/username/Library/Mobile Documents/com~apple~CloudDocs/"
```

#### Migrating to a New Computer
```bash
abletoncli migrate --replace "/Users/oldusername/Documents" --with "/Users/newusername/Documents"
```

#### Reorganizing Project Structure
```bash
abletoncli migrate --replace "/Music/Old Structure/Projects" --with "/Music/New Structure/Ableton Projects"
```

### Global Flags

These flags work with any command:

- `-d, --debug` - Enable debug output (shown in blue)
- `-h, --help` - Show help for any command
- `-v, --version` - Show version information

## How It Works

### Migrate

1. Scans the specified directory for `.als` files (case-insensitive)
2. For each file:
   - Creates a temporary decompressed copy
   - Searches for the specified string
   - Replaces matches with the new string
   - Recompresses and writes back to the original file (unless `--dry-run`)
   - Cleans up temporary files
3. Provides a summary of files processed, modified, and total line changes

### Backup

1. Scans the specified directory for `.als` files
2. Replicates the directory structure in the backup location
3. Copies each file to the corresponding location in the backup directory
4. Preserves file permissions and metadata

## Technical Details

- **Built with**: Go and [Cobra CLI framework](https://github.com/spf13/cobra)
- **File Format**: Ableton Live `.als` files are gzipped XML documents
- **Version**: 0.1.2
- **Compatibility**: macOS, Linux, Windows

## Development

### Prerequisites

- Go 1.21 or higher

### Building from Source

```bash
# Clone the repository
git clone <repository-url>
cd AbletonCLI

# Build the binary
go build -o bin/abletoncli

# Run tests (if available)
go test ./...
```

### Project Structure

```
AbletonCLI/
├── bin/              # Compiled binary
├── cmd/              # Command implementations
│   ├── backup.go     # Backup command
│   ├── migrate.go    # Migrate command
│   ├── logging.go    # Colored logging utilities
│   └── root.go       # Root command and CLI setup
├── completion.sh     # Shell completion script
├── install.sh        # Installation script
├── main.go          # Entry point
└── go.mod           # Go dependencies
```

## Contributing

Contributions are welcome! Please feel free to submit issues or pull requests.

## License

See [LICENSE](LICENSE) file for details.

## Safety & Best Practices

- Always backup your Ableton projects before performing migrations
- Use `--dry-run` first to preview changes
- Enable debug mode (`-d`) when troubleshooting issues
- Temporary files are automatically cleaned up, even if errors occur

## Troubleshooting

### Command not found after installation

Make sure `/usr/local/bin/abletoncli` is in your PATH. The installer should handle this automatically, but you can manually add it:

```bash
# For zsh (add to ~/.zshrc)
export PATH="/usr/local/bin/abletoncli:$PATH"

# For bash (add to ~/.bashrc or ~/.bash_profile)
export PATH="/usr/local/bin/abletoncli:$PATH"
```

Then reload your shell configuration:

```bash
source ~/.zshrc  # or source ~/.bashrc
```

### Shell completion not working

Source the completion script manually:

```bash
source /usr/local/bin/abletoncli/completion.sh
```

Or re-run the installer and accept the prompt to update your shell configuration.
