package main

import (
	"os"
	"reflect"
	"strings"
	"testing"
)

func createTempFile(t *testing.T, content string) string {
	f, err := os.CreateTemp("", "testinput")
	if err != nil {
		t.Fatal(err)
	}
	_, err = f.WriteString(content)
	if err != nil {
		t.Fatal(err)
	}
	f.Close()
	return f.Name()
}

func TestParseRedirectInputOutput(t *testing.T) {
	var opened []*os.File
	input := createTempFile(t, "data")
	output := createTempFile(t, "")
	defer os.Remove(input)
	defer os.Remove(output)

	args := []string{"cat", "<", input, ">", output}
	cmd, err := parseRedirect(&opened, args)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cmd.Stdin == nil || cmd.Stdout == nil {
		t.Errorf("Stdin or Stdout not set")
	}
	if !reflect.DeepEqual(len(opened), 2) {
		t.Errorf("opened files %v", opened)
	}
}

func TestRunPipeLineSimpleEcho(t *testing.T) {
	// Должен успешно выполниться, если echo есть в системе
	err := runPipeLine("echo hello")
	if err != nil {
		t.Errorf("expected nil error, got %v", err)
	}
}

func TestRunPipeLinePipe(t *testing.T) {
	// Не проверяет вывод, но pipeline должен не падать
	err := runPipeLine("echo hi | grep h")
	if err != nil {
		t.Errorf("expected nil error, got %v", err)
	}
}

func TestRunPipeLineFileRedirect(t *testing.T) {
	out := createTempFile(t, "")
	defer os.Remove(out)
	err := runPipeLine("echo test > " + out)
	if err != nil {
		t.Errorf("expected nil error, got %v", err)
	}
	data, _ := os.ReadFile(out)
	got := strings.TrimSpace(string(data))
	if got != "test" {
		t.Errorf("got %q, want %q", got, "test")
	}
}
