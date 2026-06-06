package main

import "testing"

func TestGetCommands(t *testing.T) {
	commands := getCommands()

	// Every command we rely on must be registered.
	required := []string{"help", "exit"}
	for _, name := range required {
		if _, ok := commands[name]; !ok {
			t.Errorf("getCommands(): expected a %q command to be registered", name)
		}
	}

	// Guard against the map-key / name mismatch footgun: the key used for
	// lookup must equal the command's own name field, and every command must
	// have a description and a non-nil callback.
	for key, command := range commands {
		if command.name != key {
			t.Errorf("command %q: name field is %q, want it to match the map key",
				key, command.name)
		}
		if command.description == "" {
			t.Errorf("command %q: description is empty", key)
		}
		if command.callback == nil {
			t.Errorf("command %q: callback is nil", key)
		}
	}
}

func TestCleanInput(t *testing.T) {
	cases := []struct {
		input    string
		expected []string
	}{
		{
			input:    "  hello  world  ",
			expected: []string{"hello", "world"},
		},
		{
			input:    "Charmander Bulbasaur PIKACHU",
			expected: []string{"charmander", "bulbasaur", "pikachu"},
		},
		{
			input:    "",
			expected: []string{},
		},
		{
			input:    "   ",
			expected: []string{},
		},
		{
			input:    "SingleWord",
			expected: []string{"singleword"},
		},
		{
			input:    "\tCaterpie\nWeedle  ",
			expected: []string{"caterpie", "weedle"},
		},
	}
	for _, c := range cases {
		actual := cleanInput(c.input)
		// Check the length of the actual slice against the expected slice
		// if they don't match, use t.Errorf to print an error message
		// and fail the test
		if len(actual) != len(c.expected) {
			t.Errorf("cleanInput(%q): expected length %d, got %d (%v)",
				c.input, len(c.expected), len(actual), actual)
			continue
		}
		for i := range actual {
			word := actual[i]
			expectedWord := c.expected[i]
			// Check each word in the slice
			// if they don't match, use t.Errorf to print an error message
			// and fail the test
			if word != expectedWord {
				t.Errorf("cleanInput(%q): at index %d expected %q, got %q",
					c.input, i, expectedWord, word)
			}
		}
	}
}
