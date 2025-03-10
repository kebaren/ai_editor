# GoTextBuffer Editor

A text editor GUI application built with GTK4 using the gotk4 framework and the custom textbuffer implementation.

## Features

- Open and save text files
- Undo/redo functionality
- Search and replace with options (case sensitive, whole word, regex)
- Large file support
- Status bar with cursor position and file status

## Requirements

- Go 1.21 or later
- GTK 4.0 or later
- gotk4 dependencies

## Installation

### Install GTK4

#### Windows

1. Install MSYS2 from https://www.msys2.org/
2. Open MSYS2 MINGW64 shell and run:
   ```
   pacman -S mingw-w64-x86_64-gtk4 mingw-w64-x86_64-pkgconf mingw-w64-x86_64-gcc
   ```

#### Linux

```bash
# Ubuntu/Debian
sudo apt install libgtk-4-dev pkg-config gcc

# Fedora
sudo dnf install gtk4-devel pkgconf gcc
```

#### macOS

```bash
brew install gtk4 pkg-config gcc
```

### Install Go Dependencies

```bash
cd gui
go mod tidy
```

## Building and Running

```bash
cd gui
go build -o editor.exe
./editor.exe
```

## Usage

### File Operations

- **New File**: Create a new empty file
- **Open File**: Open an existing file
- **Save**: Save the current file
- **Save As**: Save the current file with a new name

### Editing

- **Undo**: Undo the last operation
- **Redo**: Redo the last undone operation
- **Cut/Copy/Paste**: Standard clipboard operations

### Search and Replace

- **Find**: Search for text in the document
- **Replace**: Replace text in the document

## Implementation Details

The application uses a custom textbuffer implementation that provides efficient text handling, especially for large files. The GUI is built with GTK4 using the gotk4 framework, which provides Go bindings for GTK.

The main components are:

- **TextBuffer**: The core text handling component
- **EditorWindow**: The main application window
- **BufferAdapter**: Handles the integration between GTK's text buffer and our custom textbuffer 