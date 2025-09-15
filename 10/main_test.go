package main

import (
	"bufio"
	"bytes"
	"io"
	"os"
	"strings"
	"testing"
)

func TestGetCol(t *testing.T) {
	tests := []struct {
		row  string
		col  int
		want string
	}{
		{"a\tb\tc", 1, "a"},
		{"a\tb\tc", 2, "b"},
		{"a\tb\tc", 3, "c"},
		{"a\tb\tc", 4, ""},
	}
	for _, tt := range tests {
		got := getCol(tt.row, tt.col)
		if got != tt.want {
			t.Errorf("getCol(%q, %d) = %q; want %q", tt.row, tt.col, got, tt.want)
		}
	}
}

func TestParseHumanSize(t *testing.T) {
	tests := []struct {
		input   string
		want    float64
		wantErr bool
	}{
		{"100", 100, false},
		{"1K", 1024, false},
		{"1.5M", 1.5 * 1024 * 1024, false},
		{"2G", 2 * 1024 * 1024 * 1024, false},
		{"", 0, true},
		{"abc", 0, true},
	}
	for _, tt := range tests {
		got, err := parseHumanSize(tt.input)
		if (err != nil) != tt.wantErr {
			t.Errorf("parseHumanSize(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
		}
		if !tt.wantErr && got != tt.want {
			t.Errorf("parseHumanSize(%q) = %v, want %v", tt.input, got, tt.want)
		}
	}
}

func TestCompareRows(t *testing.T) {
	flagMap := map[string]bool{"hasN": true}
	rows := []string{"10", "2"}
	if !compareRows(1, 0, rows, 0, flagMap) {
		t.Error("compareRows numeric: expected row 1 < row 0")
	}

	flagMap = map[string]bool{"hasR": true}
	rows = []string{"a", "b"}
	if !compareRows(1, 0, rows, 0, flagMap) {
		t.Error("compareRows reverse: expected row 1 < row 0")
	}

	flagMap = map[string]bool{"hasB": true}
	rows = []string{" a ", "a"}
	if compareRows(1, 0, rows, 0, flagMap) {

		t.Error("compareRows -b flag: expected equality ignoring spaces")
	}
}

func TestSortWithFlags(t *testing.T) {
	rows := []string{"c", "a", "b", "a"}
	flagMap := map[string]bool{"hasU": true}
	sorted := sortWithFlags(rows, 0, flagMap)
	if len(sorted) != 3 {
		t.Errorf("sortWithFlags unique: expected 3 unique rows, got %d", len(sorted))
	}
}

func TestProcess(t *testing.T) {
	input := "c\nb\na\na\n"
	flags := []string{"-u"}
	err := procces(flags, strings.NewReader(input))
	if err != nil {
		t.Errorf("procces failed: %v", err)
	}
}

func TestCheckSorted(t *testing.T) {
	sortedInput := "a\nb\nc\n"
	unsortedInput := "b\na\nc\n"
	flagMap := make(map[string]bool)

	err := checkSorted(strings.NewReader(sortedInput), 0, flagMap)
	if err != nil {
		t.Errorf("checkSorted failed on sorted input: %v", err)
	}

	err = checkSorted(strings.NewReader(unsortedInput), 0, flagMap)
	if err == nil {
		t.Errorf("checkSorted did not detect unsorted input")
	}
}

func TestParseArgs(t *testing.T) {
	args := []string{"cmd", "-n", "1"}
	flags, in, err := parseArgs(args)
	if err != nil {
		t.Errorf("parseArgs failed: %v", err)
	}
	if len(flags) != 2 {
		t.Errorf("parseArgs flags length = %d; want 2", len(flags))
	}
	if in == nil {
		t.Errorf("parseArgs in reader is nil")
	}
}

func TestProcessWithFileSorting(t *testing.T) {
	data := "delta\nalpha\ncharlie\nbravo\n"

	tmpFile, err := os.CreateTemp("", "testfile")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	_, err = tmpFile.WriteString(data)
	if err != nil {
		t.Fatalf("Failed to write to temp file: %v", err)
	}
	if err := tmpFile.Close(); err != nil {
		t.Fatalf("Failed to close temp file: %v", err)
	}

	file, err := os.Open(tmpFile.Name())
	if err != nil {
		t.Fatalf("Failed to reopen temp file: %v", err)
	}
	defer file.Close()

	flags := []string{}

	var buf bytes.Buffer
	origStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	done := make(chan struct{})
	go func() {
		io.Copy(&buf, r)
		close(done)
	}()

	err = procces(flags, file)
	if err != nil {
		t.Fatalf("procces returned error: %v", err)
	}

	w.Close()
	os.Stdout = origStdout
	<-done

	scanner := bufio.NewScanner(&buf)
	var prev string
	lineNum := 0
	for scanner.Scan() {
		line := scanner.Text()
		if lineNum > 0 && line < prev {
			t.Errorf("Output is not sorted at line %d: %q < %q", lineNum+1, line, prev)
		}
		prev = line
		lineNum++
	}
	if err := scanner.Err(); err != nil {
		t.Errorf("Scanner error: %v", err)
	}
}
