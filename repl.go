package main

import (
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/mannyzzle/pokedexcli/internal/pokeapi"
)

/* ------------------------------------------------------------------ */
/*  Shared cache instance                                             */
/* ------------------------------------------------------------------ */

var cache = pokeapi.Cache()

/* ------------------------------------------------------------------ */
/*  Command registry / types                                          */
/* ------------------------------------------------------------------ */

type cliCommand struct {
	name, description string
	callback          func(*config, []string) error
}

var commands map[string]cliCommand

func initCommands() {
	commands = map[string]cliCommand{
		"help":    {"help", "Displays a help message", commandHelp},
		"exit":    {"exit", "Exit the Pokedex", commandExit},
		"map":     {"map", "Show next 20 location-areas", commandMap},
		"mapb":    {"mapb", "Show previous 20 location-areas", commandMapBack},
		"explore": {"explore", "List Pokémon in an area", commandExplore},
		"catch":   {"catch", "Attempt to catch a Pokémon", commandCatch},
		"inspect": {"inspect", "Show details of a caught Pokémon", commandInspect},
		"pokedex": {"pokedex", "List all caught Pokémon", commandPokedex},
	}
	rand.Seed(time.Now().UnixNano())
}
/* ------------------------------------------------------------------ */
/*  Helpers                                                           */
/* ------------------------------------------------------------------ */

func cleanInput(s string) []string {
	return strings.Fields(strings.ToLower(strings.TrimSpace(s)))
}

func ptr(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

/* ------------------------------------------------------------------ */
/*  Command implementations                                           */
/* ------------------------------------------------------------------ */

func commandHelp(_ *config, _ []string) error {
	fmt.Println("Welcome to the Pokedex!")
	fmt.Println("Usage:")
	for _, c := range commands {
		fmt.Printf("%s: %s\n", c.name, c.description)
	}
	return nil
}

func commandExit(_ *config, _ []string) error {
	fmt.Println("Closing the Pokedex... Goodbye!")
	os.Exit(0)
	return nil
}

func commandMap(cfg *config, _ []string) error {
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

func commandMapBack(cfg *config, _ []string) error {
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

func commandExplore(cfg *config, args []string) error {
	if len(args) == 0 {
		fmt.Println("usage: explore <location-area>")
		return nil
	}
	areaName := args[0]
	url := fmt.Sprintf("https://pokeapi.co/api/v2/location-area/%s/", areaName)

	raw, ok := cache.Get(url)
	if !ok {
		resp, err := cfg.api.HC().Get(url)
		if err != nil {
			return err
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			return fmt.Errorf("pokeapi: status %s", resp.Status)
		}
		raw, _ = io.ReadAll(resp.Body)
		cache.Add(url, raw)
	}

	var payload struct {
		PokemonEncounters []struct {
			Pokemon struct {
				Name string `json:"name"`
			} `json:"pokemon"`
		} `json:"pokemon_encounters"`
	}
	if err := json.Unmarshal(raw, &payload); err != nil {
		return err
	}

	if len(payload.PokemonEncounters) == 0 {
		fmt.Printf("No Pokémon found in %s\n", areaName)
		return nil
	}
	fmt.Printf("Pokémon found in %s:\n", areaName)
	for _, p := range payload.PokemonEncounters {
		fmt.Println("-", p.Pokemon.Name)
	}
	return nil
}

func commandCatch(cfg *config, args []string) error {
	if len(args) == 0 {
		fmt.Println("usage: catch <pokemon>")
		return nil
	}
	name := strings.ToLower(args[0])

	if _, ok := cfg.caught[name]; ok {
		fmt.Printf("%s is already in your Pokédex!\n", name)
		return nil
	}

	fmt.Printf("Throwing a Pokeball at %s...\n", name)

	url := fmt.Sprintf("https://pokeapi.co/api/v2/pokemon/%s", name)

	raw, ok := cache.Get(url)
	if !ok {
		resp, err := cfg.api.HC().Get(url)
		if err != nil {
			return err
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			fmt.Printf("%s doesn't seem to exist.\n", name)
			return nil
		}
		raw, _ = io.ReadAll(resp.Body)
		cache.Add(url, raw)
	}

	var poke struct {
		Name           string `json:"name"`
		BaseExperience int    `json:"base_experience"`
		Height         int    `json:"height"`
		Weight         int    `json:"weight"`
		Stats          []struct {
			Base int `json:"base_stat"`
			Stat struct {
				Name string `json:"name"`
			} `json:"stat"`
		} `json:"stats"`
		Types []struct {
			Type struct {
				Name string `json:"name"`
			} `json:"type"`
		} `json:"types"`
	}
	if err := json.Unmarshal(raw, &poke); err != nil {
		return err
	}

	// catch probability
	p := 200.0 / float64(poke.BaseExperience+50)
	if p < 0.2 {
		p = 0.2
	}
	if rand.Float64() >= p {
		fmt.Printf("%s escaped!\n", poke.Name)
		return nil
	}

	// Build stats map and types slice
	stats := make(map[string]int)
	for _, s := range poke.Stats {
		stats[s.Stat.Name] = s.Base
	}
	var types []string
	for _, t := range poke.Types {
		types = append(types, t.Type.Name)
	}

	cfg.caught[poke.Name] = PokedexEntry{
		Name:           poke.Name,
		BaseExperience: poke.BaseExperience,
		Height:         poke.Height,
		Weight:         poke.Weight,
		Stats:          stats,
		Types:          types,
	}

	fmt.Printf("%s was caught!\n", poke.Name)
	return nil
}

/* ------------------------------------------------------------------ */
/*  inspect                                                           */
/* ------------------------------------------------------------------ */

func commandInspect(cfg *config, args []string) error {
	if len(args) == 0 {
		fmt.Println("usage: inspect <pokemon>")
		return nil
	}
	name := strings.ToLower(args[0])

	entry, ok := cfg.caught[name]
	if !ok {
		fmt.Println("you have not caught that pokemon")
		return nil
	}

	fmt.Printf("Name: %s\n", entry.Name)
	fmt.Printf("Height: %d\n", entry.Height)
	fmt.Printf("Weight: %d\n", entry.Weight)
	fmt.Println("Stats:")
	for stat, val := range entry.Stats {
		fmt.Printf("  -%s: %d\n", stat, val)
	}
	fmt.Println("Types:")
	for _, t := range entry.Types {
		fmt.Printf("  - %s\n", t)
	}
	return nil
}



/* ------------------------------------------------------------------ */
/*  Pokedex                                                           */
/* ------------------------------------------------------------------ */


func commandPokedex(cfg *config, _ []string) error {
	if len(cfg.caught) == 0 {
		fmt.Println("Your Pokedex is empty — go catch some Pokémon first!")
		return nil
	}

	fmt.Println("Your Pokedex:")
	for name := range cfg.caught {
		fmt.Printf(" - %s\n", name)
	}
	return nil
}
