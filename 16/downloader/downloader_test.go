package downloader

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestDownloadFile(t *testing.T) {
	d := NewDownloader("")

	baseDir := t.TempDir()

	content := "test content"
	reader := strings.NewReader(content)

	savedPath, err := d.DownloadFile(baseDir, "subdir/file.txt", reader)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	data, err := os.ReadFile(savedPath)
	if err != nil {
		t.Fatalf("cannot read written file: %v", err)
	}
	if string(data) != content {
		t.Errorf("file content = %q; want %q", string(data), content)
	}

	reader2 := strings.NewReader(content)
	savedPath2, err := d.DownloadFile(baseDir, "folder/", reader2)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.HasSuffix(savedPath2, filepath.Join("folder", "index.html")) {
		t.Errorf("expected file path to end with 'folder/index.html', got %s", savedPath2)
	}
}

func TestGetPath(t *testing.T) {
	d := NewDownloader("")
	tests := []struct {
		input string
		want  string
	}{
		{"https://example.com/path", "example.com/path"},
		{"http://example.com/", "example.com/"},
		{"example.com/path", "example.com/path"},
	}
	for _, tt := range tests {
		got := d.GetPath(tt.input)
		if got != tt.want {
			t.Errorf("GetPath(%q) = %q; want %q", tt.input, got, tt.want)
		}
	}
}

func TestGetBasePath(t *testing.T) {
	d := NewDownloader("")
	tests := []struct {
		input string
		want  string
	}{
		{"https://example.com/path/to/page", "example.com"},
		{"http://example.com/another/path", "example.com"},
		{"example.com/path", "example.com"},
	}
	for _, tt := range tests {
		got := d.GetBasePath(tt.input)
		if got != tt.want {
			t.Errorf("GetBasePath(%q) = %q; want %q", tt.input, got, tt.want)
		}
	}
}
