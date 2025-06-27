package main

import (
	"strings"
)

func cleanInput(text string) []string {
	trimmed := strings.TrimSpace(text)
	lowered := strings.ToLower(trimmed)
	words := strings.Fields(lowered) // automatically splits by any whitespace
	return words
}
