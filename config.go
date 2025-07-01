package main

import "github.com/mannyzzle/pokedexcli/internal/pokeapi"

// PokedexEntry stores everything we need for `inspect`.
type PokedexEntry struct {
	Name           string
	BaseExperience int
	Height         int
	Weight         int
	Stats          map[string]int // hp, attack, etc.
	Types          []string
}

type config struct {
	next, prev *string
	api        *pokeapi.Client
	caught     map[string]PokedexEntry
}

func newConfig() *config {
	return &config{
		api:    pokeapi.NewClient(),
		caught: make(map[string]PokedexEntry),
	}
}
