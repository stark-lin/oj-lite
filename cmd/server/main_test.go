package main

import (
	"io"
	"testing"
)

func TestParseServerOptions(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		args         []string
		wantSkipSeed bool
	}{
		{
			name: "default seeds demo data",
		},
		{
			name:         "skip seed flag",
			args:         []string{"--skip-seed"},
			wantSkipSeed: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := parseServerOptions(tt.args, io.Discard)
			if err != nil {
				t.Fatalf("parse server options: %v", err)
			}

			if got.skipSeed != tt.wantSkipSeed {
				t.Fatalf("skipSeed = %v, want %v", got.skipSeed, tt.wantSkipSeed)
			}
		})
	}
}
