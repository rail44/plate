package plate

import (
	"fmt"
	"go/format"
	"os"
	"path/filepath"
)

// WriteOptions contains options for writing generated files
type WriteOptions struct {
	// BaseDir is the base directory where files will be written
	BaseDir string
	// Clean determines whether to remove the output directory before writing
	Clean bool
}

// defaultWriteOptions returns default write options
func defaultWriteOptions() WriteOptions {
	return WriteOptions{
		BaseDir: ".",
		Clean:   false, // Keep false by default for safety
	}
}

// writeFiles writes the generated files to disk
func writeFiles(files map[string]string, opts WriteOptions) error {
	// Clean the base directory if requested
	if opts.Clean && opts.BaseDir != "" && opts.BaseDir != "." {
		if err := os.RemoveAll(opts.BaseDir); err != nil {
			return fmt.Errorf("failed to clean directory %s: %w", opts.BaseDir, err)
		}
	}

	for path, content := range files {
		if err := writeFile(filepath.Join(opts.BaseDir, path), content, opts); err != nil {
			return fmt.Errorf("failed to write %s: %w", path, err)
		}
	}
	return nil
}

// writeFile writes a single file to disk
func writeFile(path, content string, opts WriteOptions) error {
	// Create directory if it doesn't exist
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory %s: %w", dir, err)
	}

	// Format Go code
	if filepath.Ext(path) == ".go" {
		formatted, err := format.Source([]byte(content))
		if err != nil {
			// If formatting fails, write the original content and warn
			fmt.Printf("Warning: failed to format %s: %v\n", path, err)
		} else {
			content = string(formatted)
		}
	}

	// Write file
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}

// GeneratedFiles represents the output of the generator
type GeneratedFiles struct {
	Files map[string]string
}

// WriteToDirectory writes all generated files to the specified directory
func (gf GeneratedFiles) WriteToDirectory(dir string) error {
	return gf.WriteToDirectoryWithOptions(dir, defaultWriteOptions())
}

// WriteToDirectoryWithOptions writes all generated files with custom options
func (gf GeneratedFiles) WriteToDirectoryWithOptions(dir string, opts WriteOptions) error {
	opts.BaseDir = dir
	return writeFiles(gf.Files, opts)
}

// WriteToDirectoryClean writes all generated files after cleaning the directory
func (gf GeneratedFiles) WriteToDirectoryClean(dir string) error {
	opts := defaultWriteOptions()
	opts.Clean = true
	return gf.WriteToDirectoryWithOptions(dir, opts)
}
