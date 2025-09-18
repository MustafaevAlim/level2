package parser

import (
	"errors"
	"io"
	"strings"
	"testing"
)

func TestNormalizeURL(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"https://example.com", "https://example.com/"},
		{"https://example.com/path", "https://example.com/path/"},
		{"https://example.com/file.html", "https://example.com/file.html"},
		{"https://EXAMPLE.com/", "https://example.com/"},
	}

	for _, tt := range tests {
		got, err := normalizeURL(tt.input)
		if tt.want == "" && errors.Is(err, nil) {
			t.Errorf("normalizeURL(%q) expected error but got nil", tt.input)
			continue
		}
		if tt.want != "" && (err != nil || got != tt.want) {
			t.Errorf("normalizeURL(%q) = %q, %v; want %q, nil", tt.input, got, err, tt.want)
		}
	}
}

func TestCheckHostUrl(t *testing.T) {
	if !CheckHostUrl("https://example.com") {
		t.Error("expected true for https URL")
	}
	if !CheckHostUrl("http://example.com") {
		t.Error("expected true for http URL")
	}
	if CheckHostUrl("ftp://example.com") {
		t.Error("expected false for ftp URL")
	}
	if CheckHostUrl("example.com") {
		t.Error("expected false for no scheme URL")
	}
}

func TestParserHTML_checkValidUrl(t *testing.T) {
	p := NewParserHTML(1, "https://example.com")

	ok, err := p.checkValidUrl("https://example.com/page")
	if err != nil || !ok {
		t.Errorf("expected true, got %v, error: %v", ok, err)
	}
	ok, err = p.checkValidUrl("https://other.com/page")
	if err != nil || ok {
		t.Errorf("expected false for different host, got %v, error: %v", ok, err)
	}

	ok, err = p.checkValidUrl("https://example.com/page")
	if err != nil || ok {
		t.Errorf("expected false for revisited URL, got %v, error: %v", ok, err)
	}
}

func TestParserHTML_FindAllLink(t *testing.T) {
	htmlData := `
    <html>
        <body>
            <a href="page1.html">Page1</a>
            <img src="image.png"/>
            <a href="javascript:void(0)">JS Link</a>
            <a href="#top">Anchor</a>
            <a href="https://example.com/page2.html">External</a>
        </body>
    </html>
    `

	p := NewParserHTML(1, "https://example.com")
	p.visitedURL = make(map[string]struct{})

	links, err := p.FindAllLink(io.NopCloser(strings.NewReader(htmlData)))
	if err != nil {
		t.Fatalf("FindAllLink error: %v", err)
	}

	wantLinks := []string{"page1.html", "image.png", "https://example.com/page2.html"}
	for _, w := range wantLinks {
		found := false
		for _, l := range links {
			if l == w {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("expected link %q in result, got %v", w, links)
		}
	}
}
