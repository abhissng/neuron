# Upload File Validator

A flexible file upload validation package with functional options pattern.

## Usage

### Basic Validation with Profile

```go
import "github.com/abhissng/neuron/adapters/file/uploadFile"

// Create validator with profile
validator := uploadFile.NewUploadFileValidator(
    uploadFile.WithProfile(uploadFile.ProfileImage),
)

// Validate a single image file
err := validator.ValidateFile(fileHeader)
```

### Validation with Virus Scanner

```go
// With ClamAV scanner
validator := uploadFile.NewUploadFileValidator(
    uploadFile.WithProfile(uploadFile.ProfileDocument),
    uploadFile.WithClamAV("localhost:3310"),
)
err := validator.ValidateFile(fileHeader)

// With custom virus scanner
scanner := &MyCustomScanner{}
validator := uploadFile.NewUploadFileValidator(
    uploadFile.WithProfile(uploadFile.ProfileVideo),
    uploadFile.WithVirusScanner(scanner),
)
err := validator.ValidateFile(fileHeader)
```

### Batch File Validation

```go
// Create validator once
validator := uploadFile.NewUploadFileValidator(
    uploadFile.WithProfile(uploadFile.ProfileMixed),
)

// Validate multiple files at once
err := validator.ValidateFiles(fileHeaders)
```

### Custom Validation Rules

```go
// Create custom rule inline
validator := uploadFile.NewUploadFileValidator(
    uploadFile.WithCustomRule(
        10*uploadFile.MB,
        []string{"image/jpeg", "image/png"},
        []string{".jpg", ".png"},
    ),
)
err := validator.ValidateFile(fileHeader)

// Or use a predefined custom rule
customRule := uploadFile.NewCustomRule(
    50*uploadFile.MB,
    []string{"application/json", "text/plain"},
    []string{".json", ".txt"},
)

validator := uploadFile.NewUploadFileValidator(
    uploadFile.WithRule(customRule),
)
err := validator.ValidateFile(fileHeader)
```

### Combining Options

```go
// Combine multiple options
validator := uploadFile.NewUploadFileValidator(
    uploadFile.WithProfile(uploadFile.ProfileImage),
    uploadFile.WithClamAV("localhost:3310"),
)
err := validator.ValidateFile(fileHeader)
```

## Integration with AppContext

You can easily store validator instances in your app context:

```go
type AppContext struct {
    ImageValidator   *uploadFile.Config
    DocumentValidator *uploadFile.Config
}

func NewAppContext() *AppContext {
    return &AppContext{
        ImageValidator: uploadFile.NewUploadFileValidator(
            uploadFile.WithProfile(uploadFile.ProfileImage),
            uploadFile.WithClamAV("localhost:3310"),
        ),
        DocumentValidator: uploadFile.NewUploadFileValidator(
            uploadFile.WithProfile(uploadFile.ProfileDocument),
            uploadFile.WithClamAV("localhost:3310"),
        ),
    }
}

// Usage in handlers
func (app *AppContext) UploadImageHandler(file *multipart.FileHeader) error {
    return app.ImageValidator.ValidateFile(file)
}

func (app *AppContext) UploadDocumentsHandler(files []*multipart.FileHeader) error {
    return app.DocumentValidator.ValidateFiles(files)
}
```

## Available Profiles

- `ProfileImage` - JPEG, PNG, WebP (max 20MB)
- `ProfileVideo` - MP4, WebM, MKV (max 5GB)
- `ProfileAudio` - MP3, WAV (max 500MB)
- `ProfileDocument` - PDF, TXT, CSV, XLS, XLSX (max 50MB)
- `ProfileMixed` - All of the above
- `ProfileCustom` - Use with custom rules

## Available Options

- `WithProfile(profile UploadProfile)` - Use predefined profile
- `WithRule(rule *FileRule)` - Use custom rule object
- `WithCustomRule(maxSize int64, mimes []string, exts []string)` - Create custom rule inline
- `WithVirusScanner(scanner VirusScanner)` - Add custom virus scanner
- `WithClamAV(address string)` - Add ClamAV scanner

## Methods

- `ValidateFile(file *multipart.FileHeader) error` - Validate a single file
- `ValidateFiles(files []*multipart.FileHeader) error` - Validate multiple files

## Error Handling

```go
validator := uploadFile.NewUploadFileValidator(
    uploadFile.WithProfile(uploadFile.ProfileImage),
)

err := validator.ValidateFile(fileHeader)
if err != nil {
    switch err {
    case uploadFile.ErrFileTooLarge:
        // Handle file size error
    case uploadFile.ErrInvalidMimeType:
        // Handle MIME type error
    case uploadFile.ErrInvalidExtension:
        // Handle extension error
    case uploadFile.ErrVirusDetected:
        // Handle virus detection
    default:
        // Handle other errors
    }
}
```
