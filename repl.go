package main

import (
	"fmt"
	"os"
	"strings"
)

/* ---------- types ---------- */

type cliCommand struct {
	name        string
	description string
	callback    func(*config) error
}

/* ---------- registry ---------- */

var commands map[string]cliCommand

func initCommands() {
	commands = map[string]cliCommand{
		"help": {
			name:        "help",
			description: "Displays a help message",
			callback:    commandHelp,
		},
		"exit": {
			name:        "exit",
			description: "Exit the Pokedex",
			callback:    commandExit,
		},
		"map": {
			name:        "map",
			description: "Show next 20 location-areas",
			callback:    commandMap,
		},
		"mapb": {
			name:        "mapb",
			description: "Show previous 20 location-areas",
			callback:    commandMapBack,
		},
	}
}

/* ---------- helpers ---------- */

func cleanInput(s string) []string {
	return strings.Fields(strings.ToLower(strings.TrimSpace(s)))
}

func ptr(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

/* ---------- command impls ---------- */

func commandHelp(cfg *config) error {
	fmt.Println("Welcome to the Pokedex!")
	fmt.Println("Usage:")
	for _, c := range commands {
		fmt.Printf("%s: %s\n", c.name, c.description)
	}
	return nil
}

func commandExit(_ *config) error {
	fmt.Println("Closing the Pokedex... Goodbye!")
	os.Exit(0)
	return nil
}

func commandMap(cfg *config) error {
	names, next, prev, err := cfg.api.GetPage(ptr(cfg.next))
	if err != nil {
		return err
	}
	for _, n := range names {
		fmt.Println(n)
	}
	cfg.next, cfg.prev = next, prev
	return nil
}

func commandMapBack(cfg *config) error {
	if cfg.prev == nil {
		fmt.Println("you're on the first page")
		return nil
	}
	names, next, prev, err := cfg.api.GetPage(*cfg.prev)
	if err != nil {
		return err
	}
	for _, n := range names {
		fmt.Println(n)
	}
	cfg.next, cfg.prev = next, prev
	return nil
}
