package main

import "github.com/mannyzzle/pokedexcli/internal/pokeapi"

type config struct {
	next *string
	prev *string
	api  *pokeapi.Client
}

func newConfig() *config {
	return &config{api: pokeapi.NewClient()}
}
