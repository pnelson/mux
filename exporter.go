package mux

import (
	"os"
	"path/filepath"
)

// Exporter represents a route exporter.
type Exporter interface {
	Export(r *Route, b []byte) error
}

// FileSystemExporter is an Exporter implementation that writes to the
// directory by the string value. The route name is used to determine the
// exported filenames. An error is returned if an exported file already exists.
// An empty FileSystemExporter is treated as "dist".
type FileSystemExporter string

// Export implements the Exporter interface.
func (e FileSystemExporter) Export(r *Route, b []byte) error {
	dir := string(e)
	if dir == "" {
		dir = "dist"
	}
	filename := filepath.Join(dir, r.Name())
	err := os.MkdirAll(filepath.Dir(filename), 0755)
	if err != nil {
		return err
	}
	return os.WriteFile(filename, b, 0644)
}
