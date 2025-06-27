package main

import (
	"testing"
)

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
			input:    "\n Eevee\tSnorlax ",
			expected: []string{"eevee", "snorlax"},
		},
		{
			input:    "",
			expected: []string{},
		},
	}

	for _, c := range cases {
		actual := cleanInput(c.input)

		if len(actual) != len(c.expected) {
			t.Errorf("expected %d words, got %d for input %q", len(c.expected), len(actual), c.input)
			continue
		}

		for i := range actual {
			if actual[i] != c.expected[i] {
				t.Errorf("expected word %q but got %q at index %d for input %q", c.expected[i], actual[i], i, c.input)
			}
		}
	}
}
