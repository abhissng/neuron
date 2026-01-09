package uploadFile

import (
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/abhissng/neuron/utils/helpers"
)

/*
========================================
 Unified File Validator
========================================
*/

func (cfg *Config) ValidateFile(file *multipart.FileHeader) error {

	if cfg.rule == nil {
		return errors.New("validation rule is required")
	}
	return cfg.validateSingleFile(file)
}

func (cfg *Config) validateSingleFile(file *multipart.FileHeader) error {
	defer func() { helpers.RecoverException(recover()) }()
	if file.Size > cfg.rule.MaxSizeBytes {
		return ErrFileTooLarge
	}

	ext := strings.ToLower(filepath.Ext(file.Filename))
	if _, ok := cfg.rule.AllowedExts[ext]; !ok {
		return ErrInvalidExtension
	}

	f, err := file.Open()
	if err != nil {
		return err
	}
	defer func() {
		_ = f.Close()
	}()

	header := make([]byte, 512)
	n, err := f.Read(header)
	if err != nil && err != io.EOF {
		return err
	}

	mime := http.DetectContentType(header[:n])
	if _, ok := cfg.rule.AllowedMIMEs[mime]; !ok {
		return ErrInvalidMimeType
	}

	if cfg.virusScanner != nil {
		if seeker, ok := f.(io.Seeker); ok {
			_, _ = seeker.Seek(0, io.SeekStart)
		}

		clean, err := cfg.virusScanner.Scan(f)
		if err != nil {
			return err
		}
		if !clean {
			return ErrVirusDetected
		}
	}

	return nil
}

func (cfg *Config) ValidateFiles(files []*multipart.FileHeader) error {
	if len(files) == 0 {
		return errors.New("no files to validate")
	}

	for _, file := range files {
		if err := cfg.validateSingleFile(file); err != nil {
			return fmt.Errorf("%s: %w", file.Filename, err)
		}
	}
	return nil
}
