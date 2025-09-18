package main

import (
	"bytes"
	"io"
	"os"
	"strings"
	"testing"
)

func TestIsMatch(t *testing.T) {
	table := []struct {
		str     string
		substr  string
		param   flags
		want    bool
		wantErr bool
	}{
		{"Hello Go", "Go", flags{flagF: true}, true, false},
		{"Hello go", "go", flags{flagF: true, flagI: true}, true, false},
		{"Hello Go", "lo G", flags{flagF: true}, true, false},
		{"Hello Go", "go$", flags{flagF: false}, false, false},
		{"Hello Go", "[", flags{flagF: false}, false, true},
	}

	for _, tt := range table {
		got, err := isMatch(tt.str, tt.substr, tt.param)
		if (err != nil) != tt.wantErr {
			t.Errorf("isMatch(%q,%q,%+v) error = %v, wantErr %v", tt.str, tt.substr, tt.param, err, tt.wantErr)
		}
		if got != tt.want {
			t.Errorf("isMatch(%q,%q,%+v) = %v, want %v", tt.str, tt.substr, tt.param, got, tt.want)
		}
	}
}

func TestFoundStringWithFlags(t *testing.T) {
	input := "foo\nbar\nbaz\nfoo bar\n"
	tests := []struct {
		name   string
		param  flags
		substr string
		want   string
	}{
		{
			"Simple match",
			flags{},
			"foo",
			"foo\nfoo bar\n",
		},
		{
			"Ignore case",
			flags{flagI: true},
			"BAR",
			"bar\nfoo bar\n",
		},
		{
			"Count only",
			flags{flagC: true},
			"foo",
			"2\n",
		},
		{
			"Fixed string",
			flags{flagF: true},
			"foo",
			"foo\nfoo bar\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rdr := strings.NewReader(input)
			var buf bytes.Buffer
			origStdout := os.Stdout
			r, w, _ := os.Pipe()
			os.Stdout = w

			err := foundStringWithFlags(rdr, tt.param, tt.substr, "")
			w.Close()
			os.Stdout = origStdout
			io.Copy(&buf, r)
			got := buf.String()
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tt.want {
				t.Errorf("output = %q, want %q", got, tt.want)
			}
		})
	}
}
