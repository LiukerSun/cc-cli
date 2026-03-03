# Changelog

All notable changes to this project will be documented in this file.

## [1.1.0] - 2025-03-03

### Fixed
- **Windows PowerShell BOM handling** - Fixed UTF-8 BOM detection and removal in fix-config.ps1 using byte arrays instead of strings
- **UTF-8 encoding issues** - All scripts now save configuration files without BOM to prevent JSON parsing errors
- **PowerShell array unwrapping** - Fixed single-element array unwrapping issue by using @() wrapper for Get-Models
- **Empty argument handling** - Added check to skip null or whitespace arguments to prevent "Unknown option" errors
- **PowerShell wrapper function** - Simplified wrapper and auto-update mechanism for better compatibility
- **Unicode character display** - Replaced ✓ and other Unicode characters with ASCII equivalents ([OK]) for better terminal compatibility
- **Output formatting** - Fixed header display issues in Select-Interactive function

### Changed
- install.ps1 - Added Save-FileNoBOM helper function and improved wrapper replacement logic
- fix-config.ps1 - Rewrote with English messages and proper BOM byte handling
- bin/cc.ps1 - Added Save-JsonNoBOM helper and @() wrappers for all Get-Models calls

### Windows Compatibility
- Improved compatibility with Windows PowerShell 5.1
- Better handling of UTF-8 encoding across different PowerShell versions
- Fixed issues with Chinese character display in Windows terminals

## [1.0.0] - 2024-03-03

### Added
- Interactive model selection with arrow keys
- Direct model selection by number
- API key management (view, add, edit)
- Bypass permissions support
- Configuration file management
- Colored terminal output
- Installation and uninstall scripts
- Comprehensive documentation

### Features
- Zero external dependencies (pure Bash)
- Compatible with Bash 3.2+ (macOS default)
- Automatic shell integration
- Persistent configuration
- Model history tracking

### Documentation
- README.md
- LICENSE
- Installation guide
- Configuration examples
