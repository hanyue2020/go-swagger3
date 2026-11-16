package parser

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFindBestMainFile_SelectsRootMainWhenPresent(t *testing.T) {
	dir := t.TempDir()
	mustWriteFile(t, filepath.Join(dir, "main.go"), "package main\nfunc main() {}\n")
	mustWriteFile(t, filepath.Join(dir, "domain", "admin", "main.go"), "package main\nfunc main() {}\n")

	selected, err := findBestMainFile(dir)
	assert.NoError(t, err)
	assert.Equal(t, filepath.Join(dir, "main.go"), selected)
}

func TestFindBestMainFile_ReturnsErrorForAmbiguousServiceDetection(t *testing.T) {
	dir := t.TempDir()
	mustWriteFile(t, filepath.Join(dir, "domain", "admin", "main.go"), "package main\nfunc main() {}\n")
	mustWriteFile(t, filepath.Join(dir, "domain", "bot", "main.go"), "package main\nfunc main() {}\n")

	selected, err := findBestMainFile(dir)
	assert.Error(t, err)
	assert.Empty(t, selected)
	assert.Contains(t, err.Error(), "multiple main files found")
}

func mustWriteFile(t *testing.T, path, content string) {
	t.Helper()
	assert.NoError(t, os.MkdirAll(filepath.Dir(path), 0o755))
	assert.NoError(t, os.WriteFile(path, []byte(content), 0o644))
}
