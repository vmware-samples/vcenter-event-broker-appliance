package main

import (
	"testing"
)

func TestIsEcho(t *testing.T) {
	tests := []struct {
		input string
		want  bool
	}{
		{"blah blah blah.echo", true},
		{"blah blah blah", false},
		{"echo echo .echo blah blah blah", false},
		{"echo echo .echo blah blah blah.echo", true},
	}

	for _, tt := range tests {
		if got := isEcho(tt.input); got != tt.want {
			t.Errorf("got isEcho(%q) %t, want %t", tt.input, got, tt.want)
		}
	}
}
