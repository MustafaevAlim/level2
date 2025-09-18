package main

import (
	"bytes"
	"io"
	"os"
	"strings"
	"testing"
)

func TestCutTheFlags(t *testing.T) {
	tests := []struct {
		name  string
		flags customFlags
		input string
		want  string
	}{
		{
			"Test FlagF",
			customFlags{flagF: "2", flagD: " "},
			"ab cd eb",
			"cd \n",
		},
		{
			"Test FlagD",
			customFlags{flagD: ";", flagF: "1"},
			"ab;cd;eb",
			"ab \n",
		},
		{
			"Test FlagS",
			customFlags{flagD: ":", flagF: "1", flagS: true},
			"ab cd eb",
			"",
		},
		{
			"Test FlagF interval",
			customFlags{flagD: " ", flagF: "2-3"},
			"ab cd eb",
			"cd eb \n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reader := strings.NewReader(tt.input)
			var b bytes.Buffer
			origStdout := os.Stdout
			r, w, _ := os.Pipe()
			os.Stdout = w

			err := cutWithFlags(reader, tt.flags)
			w.Close()
			io.Copy(&b, r)

			os.Stdout = origStdout

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			got := b.String()
			if got != tt.want {
				t.Errorf("got %s; want %s", got, tt.want)
			}

		})
	}
}
