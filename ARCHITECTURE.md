# Converzen - Desktop Application Architecture

## Overview

Converzen is a cross-platform desktop application built with Wails.js that converts video and image files between different formats. The application uses Go for the backend, SvelteKit with Svelte 5 for the frontend, and SQLite for persistent storage.

## High-Level Architecture

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                              Desktop Application                             │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                              │
│  ┌─────────────────────────────────────────────────────────────────────┐    │
│  │                         Frontend (SvelteKit)                         │    │
│  │  ┌──────────────┐ ┌──────────────┐ ┌──────────────┐ ┌────────────┐  │    │
│  │  │  File Picker │ │Format Select │ │   Options    │ │  Progress  │  │    │
│  │  │  Component   │ │  Component   │ │  Component   │ │  Component │  │    │
│  │  └──────────────┘ └──────────────┘ └──────────────┘ └────────────┘  │    │
│  │                                                                      │    │
│  │  ┌──────────────────────────────────────────────────────────────┐   │    │
│  │  │                    Wails Runtime Bridge                       │   │    │
│  │  │              (Auto-generated TypeScript bindings)             │   │    │
│  │  └──────────────────────────────────────────────────────────────┘   │    │
│  └─────────────────────────────────────────────────────────────────────┘    │
│                                      │                                       │
│                                      ▼                                       │
│  ┌─────────────────────────────────────────────────────────────────────┐    │
│  │                          Backend (Go)                                │    │
│  │                                                                      │    │
│  │  ┌──────────────────────────────────────────────────────────────┐   │    │
│  │  │                      App Controller                           │   │    │
│  │  │          (Wails-bound methods, request handling)              │   │    │
│  │  └──────────────────────────────────────────────────────────────┘   │    │
│  │                                │                                     │    │
│  │         ┌──────────────────────┼──────────────────────┐             │    │
│  │         ▼                      ▼                      ▼             │    │
│  │  ┌────────────┐         ┌────────────┐         ┌────────────┐       │    │
│  │  │  File      │         │ Converter  │         │  Settings  │       │    │
│  │  │  Service   │         │  Service   │         │  Service   │       │    │
│  │  └────────────┘         └────────────┘         └────────────┘       │    │
│  │         │                      │                      │             │    │
│  │         ▼                      ▼                      ▼             │    │
│  │  ┌────────────┐         ┌────────────┐         ┌────────────┐       │    │
│  │  │  File Type │         │  FFmpeg    │         │  Database  │       │    │
│  │  │  Detector  │         │  Wrapper   │         │  Repository│       │    │
│  │  └────────────┘         └────────────┘         └────────────┘       │    │
│  │                                                       │             │    │
│  │                                                       ▼             │    │
│  │                                              ┌────────────────┐     │    │
│  │                                              │   SQLite DB    │     │    │
│  │                                              │   (via GORM)   │     │    │
│  │                                              └────────────────┘     │    │
│  └─────────────────────────────────────────────────────────────────────┘    │
│                                                                              │
└─────────────────────────────────────────────────────────────────────────────┘
```

## Component Interaction Flow

### File Conversion Workflow

```
User selects files → Frontend validates file types → Backend receives file list
                                                              │
                                                              ▼
User selects output format ← Frontend shows compatible formats ← Backend returns supported formats
         │
         ▼
User configures options (filename, output dir, copy/replace)
         │
         ▼
User clicks Convert → Frontend sends conversion request → Backend queues conversion
                                                                    │
                                                                    ▼
Frontend displays progress ← Backend emits progress events ← Converter processes files
         │                                                          │
         ▼                                                          ▼
Conversion complete ← Frontend shows result ← Backend saves to DB & returns result
```

## Backend Architecture (Go)

### Directory Structure

```
├── main.go                     # Application entry point
├── app.go                      # Main app controller (Wails bindings)
├── internal/
│   ├── config/
│   │   └── config.go           # Application configuration
│   ├── logger/
│   │   └── logger.go           # Logging service with file output
│   ├── models/
│   │   ├── file.go             # File model
│   │   ├── conversion.go       # Conversion job model
│   │   └── settings.go         # User settings model
│   ├── services/
│   │   ├── interfaces.go       # Service interfaces (SOLID: Interface Segregation)
│   │   ├── file_service.go     # File operations service
│   │   ├── converter_service.go # Conversion orchestration
│   │   ├── video_converter.go  # Video conversion (FFmpeg)
│   │   ├── image_converter.go  # Image conversion
│   │   └── settings_service.go # Settings management
│   ├── repository/
│   │   ├── interfaces.go       # Repository interfaces
│   │   ├── conversion_repo.go  # Conversion history repository
│   │   └── settings_repo.go    # Settings repository
│   └── database/
│       └── database.go         # SQLite connection and migrations
└── pkg/
    └── ffmpeg/
        └── ffmpeg.go           # FFmpeg wrapper
```

### SOLID Principles Implementation

1. **Single Responsibility Principle (SRP)**

   - Each service handles one concern (file operations, conversion, settings)
   - Repositories handle only data persistence
   - Logger handles only logging concerns

2. **Open/Closed Principle (OCP)**

   - Converter interface allows adding new format converters without modifying existing code
   - File type detectors can be extended via interface implementation

3. **Liskov Substitution Principle (LSP)**

   - All converter implementations are interchangeable via the Converter interface
   - Repository implementations can be swapped (e.g., mock for testing)

4. **Interface Segregation Principle (ISP)**

   - Small, focused interfaces (FileReader, FileWriter, Converter)
   - Clients depend only on methods they use

5. **Dependency Inversion Principle (DIP)**
   - High-level modules depend on abstractions (interfaces)
   - Dependencies injected via constructors

### Service Interfaces

```go
// Converter defines the contract for file conversion
type Converter interface {
    Convert(input ConversionJob) (ConversionResult, error)
    SupportedInputFormats() []string
    SupportedOutputFormats(inputFormat string) []string
    CanConvert(inputFormat, outputFormat string) bool
}

// FileService handles file operations
type FileService interface {
    SelectFiles() ([]FileInfo, error)
    SelectDirectory() (string, error)
    GetFileType(path string) (FileType, error)
    ValidateFiles(paths []string) ([]FileInfo, error)
}

// ConversionRepository handles conversion history
type ConversionRepository interface {
    Save(conversion *Conversion) error
    GetHistory(limit int) ([]Conversion, error)
    GetByID(id uint) (*Conversion, error)
}
```

## Frontend Architecture (SvelteKit + Svelte 5)

### Directory Structure

```
frontend/
├── src/
│   ├── lib/
│   │   ├── components/
│   │   │   ├── ui/             # shadcn-svelte components
│   │   │   ├── FileDropzone.svelte
│   │   │   ├── FormatSelector.svelte
│   │   │   ├── ConversionOptions.svelte
│   │   │   ├── ProgressDisplay.svelte
│   │   │   ├── ConversionHistory.svelte
│   │   │   └── Header.svelte
│   │   ├── stores/
│   │   │   ├── files.svelte.ts     # File selection state (Svelte 5 runes)
│   │   │   ├── conversion.svelte.ts # Conversion state
│   │   │   └── settings.svelte.ts   # User settings state
│   │   ├── types/
│   │   │   └── index.ts         # TypeScript type definitions
│   │   ├── utils/
│   │   │   └── index.ts         # Utility functions
│   │   └── wailsjs/             # Auto-generated Wails bindings
│   ├── routes/
│   │   ├── +layout.svelte       # Root layout with theme
│   │   ├── +layout.ts           # Layout config (prerender, ssr: false)
│   │   ├── +page.svelte         # Main converter page
│   │   └── history/
│   │       └── +page.svelte     # Conversion history page
│   └── app.html                 # HTML template
└── static/
    └── ...
```

### State Management (Svelte 5 Runes)

```typescript
// Using Svelte 5's $state rune for reactive state
class FilesStore {
  files = $state<FileInfo[]>([]);
  selectedFormat = $state<string>("");
  outputDirectory = $state<string>("");

  addFiles(newFiles: FileInfo[]) {
    this.files = [...this.files, ...newFiles];
  }

  clear() {
    this.files = [];
  }
}

export const filesStore = new FilesStore();
```

## Database Schema (SQLite + GORM)

### Tables

```sql
-- Conversion history
CREATE TABLE conversions (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    input_path TEXT NOT NULL,
    output_path TEXT NOT NULL,
    input_format TEXT NOT NULL,
    output_format TEXT NOT NULL,
    file_size INTEGER,
    status TEXT NOT NULL,  -- 'pending', 'processing', 'completed', 'failed'
    error_message TEXT,
    started_at DATETIME,
    completed_at DATETIME,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- User settings
CREATE TABLE settings (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    key TEXT UNIQUE NOT NULL,
    value TEXT NOT NULL,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);
```

### GORM Models

```go
type Conversion struct {
    gorm.Model
    InputPath     string
    OutputPath    string
    InputFormat   string
    OutputFormat  string
    FileSize      int64
    Status        string
    ErrorMessage  string
    StartedAt     *time.Time
    CompletedAt   *time.Time
}

type Setting struct {
    gorm.Model
    Key   string `gorm:"uniqueIndex"`
    Value string
}
```

## Logging System

### Log Levels

- **DEBUG**: Detailed information for debugging
- **INFO**: General operational information
- **WARN**: Warning messages for potential issues
- **ERROR**: Error conditions that should be addressed

### Log Output

- Console output for development
- File output (`~/.converzen/logs/app.log`) for production
- Automatic log rotation (configurable size limit)

### Log Format

```
2024-01-15 10:30:45 [INFO] [converter] Starting conversion: input.mp4 -> output.webm
2024-01-15 10:30:46 [DEBUG] [ffmpeg] Command: ffmpeg -i input.mp4 -c:v libvpx-vp9 output.webm
2024-01-15 10:31:15 [INFO] [converter] Conversion completed successfully
```

## Supported Formats

### Video Formats

| Input | Output Options           |
| ----- | ------------------------ |
| MP4   | WebM, AVI, MKV, MOV, GIF |
| WebM  | MP4, AVI, MKV, MOV, GIF  |
| AVI   | MP4, WebM, MKV, MOV, GIF |
| MKV   | MP4, WebM, AVI, MOV, GIF |
| MOV   | MP4, WebM, AVI, MKV, GIF |

### Image Formats

| Input    | Output Options                  |
| -------- | ------------------------------- |
| PNG      | JPG, JPEG, WebP, GIF, BMP, TIFF |
| JPG/JPEG | PNG, WebP, GIF, BMP, TIFF       |
| WebP     | PNG, JPG, JPEG, GIF, BMP, TIFF  |
| GIF      | PNG, JPG, JPEG, WebP, BMP, TIFF |
| BMP      | PNG, JPG, JPEG, WebP, GIF, TIFF |
| TIFF     | PNG, JPG, JPEG, WebP, GIF, BMP  |

## Error Handling Strategy

### Backend Errors

1. **Validation Errors**: Return structured error with field-specific messages
2. **Conversion Errors**: Log full stack trace, return user-friendly message
3. **System Errors**: Log with context, attempt recovery or graceful degradation

### Frontend Errors

1. **Display toast notifications** for user-actionable errors
2. **Show detailed error dialogs** for conversion failures
3. **Provide retry options** where applicable

## Cross-Platform Considerations

### macOS

- Uses native file dialogs via Wails
- FFmpeg installed via Homebrew or bundled
- App bundle (.app) for distribution

### Windows

- Uses native file dialogs via Wails
- FFmpeg bundled or installed separately
- NSIS installer for distribution

### Linux

- Uses native file dialogs via Wails
- FFmpeg installed via package manager
- AppImage or .deb/.rpm for distribution

## Security Considerations

1. **File Path Validation**: Sanitize all file paths to prevent directory traversal
2. **Input Validation**: Validate file types before processing
3. **Subprocess Execution**: Use proper escaping for FFmpeg commands
4. **No Network Access**: Application operates entirely offline
