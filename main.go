package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/cmonge21/pokedexcli/internal/pokecache"
)

type cliCommand struct {
	name        string
	description string
	callback    func([]string) error
}

type LocationAreaResp struct {
	Count    int    `json:"count"`
	Next     string `json:"next"`
	Previous string `json:"previous"`
	Results  []struct {
		Name string `json:"name"`
		URL  string `json:"url"`
	} `json:"results"`
}

type LocationArea struct {
	PokemonEncounters []struct {
		Pokemon struct {
			Name string `json:"name"`
		} `json:"pokemon"`
	} `json:"pokemon_encounters"`
}

type Config struct {
	NextURL     string
	PreviousURL string
	Pokedex     map[string]Pokemon
}

type Pokemon struct {
	Name           string `json:"name"`
	BaseExperience int    `json:"base_experience"`
	Height         int    `json:"height"`
	Weight         int    `json:"weight"`
	Stats          []Stat `json:"stats"`
	Types          []Type `json:"types"`
}

type Stat struct {
	BaseStat int      `json:"base_stat"`
	Stat     StatName `json:"stat"`
}

type StatName struct {
	Name string `json:"name"`
}

type Type struct {
	Type TypeName `json:"type"`
}

type TypeName struct {
	Name string `json:"name"`
}

var config Config

var commands map[string]cliCommand

func commandExit() error {
	fmt.Println("Closing the Pokedex... Goodbye!")
	os.Exit(0)
	return nil
}

func commandHelp() error {
	fmt.Println("Welcome to the Pokedex!")
	fmt.Println("Usage:")
	fmt.Println()
	for commandName, cmd := range commands {
		fmt.Printf("%s: %s\n", commandName, cmd.description)
	}
	return nil
}

func commandMap(cfg *Config, cache *pokecache.Cache) error {
	var resp LocationAreaResp
	url := cfg.NextURL

	if url == "" {
		url = "https://pokeapi.co/api/v2/location-area"
	}

	if cachedData, ok := cache.Get(url); ok {
		if err := json.Unmarshal(cachedData, &resp); err != nil {
			return err
		}
	} else {
		res, err := http.Get(url)
		if err != nil {
			log.Fatal(err)
		}

		body, err := io.ReadAll(res.Body)
		res.Body.Close()
		if res.StatusCode > 299 {
			log.Fatalf("Response failed with status code: %d and\n body: %s\n", res.StatusCode, body)
		}
		if err != nil {
			log.Fatal(err)
		}

		cache.Add(url, body)

		if err = json.Unmarshal(body, &resp); err != nil {
			return err
		}
	}

	cfg.NextURL = resp.Next
	cfg.PreviousURL = resp.Previous

	for _, area := range resp.Results {
		fmt.Println(area.Name)
	}
	return nil
}

func commandMapb(cfg *Config, cache *pokecache.Cache) error {
	if cfg.PreviousURL == "" {
		fmt.Println("You're on the first page")
		return nil
	}

	var resp LocationAreaResp

	url := cfg.PreviousURL
	if url == "" {
		url = "https://pokeapi.co/api/v2/location-area"
	}

	if cachedData, ok := cache.Get(url); ok {
		if err := json.Unmarshal(cachedData, &resp); err != nil {
			return err
		}
	} else {
		res, err := http.Get(url)
		if err != nil {
			log.Fatal(err)
		}

		body, err := io.ReadAll(res.Body)
		res.Body.Close()
		if res.StatusCode > 299 {
			log.Fatalf("Response failed with status code: %d and\n body: %s\n", res.StatusCode, body)
		}
		if err != nil {
			log.Fatal(err)
		}

		cache.Add(url, body)

		if err = json.Unmarshal(body, &resp); err != nil {
			return err
		}
	}

	cfg.NextURL = resp.Next
	cfg.PreviousURL = resp.Previous

	for _, area := range resp.Results {
		fmt.Println(area.Name)
	}
	return nil
}

func commandExplore(cfg *Config, cache *pokecache.Cache, location string) error {
	var resp LocationArea

	url := "https://pokeapi.co/api/v2/location-area/" + location

	if cachedData, ok := cache.Get(url); ok {
		if err := json.Unmarshal(cachedData, &resp); err != nil {
			return err
		}
	} else {
		res, err := http.Get(url)
		if err != nil {
			log.Fatal(err)
		}

		body, err := io.ReadAll(res.Body)
		res.Body.Close()
		if res.StatusCode > 299 {
			log.Fatalf("Response failed with status code: %d and\n body: %s\n", res.StatusCode, body)
		}
		if err != nil {
			log.Fatal(err)
		}

		cache.Add(url, body)

		if err = json.Unmarshal(body, &resp); err != nil {
			return err
		}
	}

	for _, area := range resp.PokemonEncounters {
		fmt.Println(area.Pokemon.Name)
	}
	return nil
}

func commandCatch(cfg *Config, cache *pokecache.Cache, name string) error {
	fmt.Printf("Throwing a Pokeball at %v...\n", name)

	var resp Pokemon

	url := "https://pokeapi.co/api/v2/pokemon/" + name

	if cachedData, ok := cache.Get(url); ok {
		if err := json.Unmarshal(cachedData, &resp); err != nil {
			return err
		}
	} else {
		res, err := http.Get(url)
		if err != nil {
			log.Fatal(err)
		}

		body, err := io.ReadAll(res.Body)
		res.Body.Close()
		if res.StatusCode > 299 {
			log.Fatalf("Response failed with status code: %d and\n body: %s\n", res.StatusCode, body)
		}
		if err != nil {
			log.Fatal(err)
		}

		cache.Add(url, body)

		if err = json.Unmarshal(body, &resp); err != nil {
			return err
		}
	}

	chance := 50 - (resp.BaseExperience / 10) // Base experience reduces chance slightly
	if chance < 1 {
		chance = 1
	}
	caught := rand.Intn(100) < chance

	if caught {
		fmt.Printf("%s was caught!\n", name)
		cfg.Pokedex[name] = resp
	} else {
		fmt.Printf("%s escaped!\n", name)
	}

	return nil
}

func commandInspect(cfg *Config, pokemonName string) {
	pokemon, exists := cfg.Pokedex[pokemonName]
	if !exists {
		fmt.Println("you have not caught that pokemon")
		return
	}

	fmt.Printf("Name: %s\n", pokemon.Name)
	fmt.Printf("Height: %d\n", pokemon.Height)
	fmt.Printf("Weight: %d\n", pokemon.Weight)

	fmt.Println("Stats:")
	for _, stat := range pokemon.Stats {
		fmt.Printf("  -%s: %d\n", stat.Stat.Name, stat.BaseStat)
	}

	fmt.Println("Types:")
	for _, t := range pokemon.Types {
		fmt.Printf("  - %s\n", t.Type.Name)
	}

}

func commandPokedex(cfg *Config) {
	fmt.Println("Your Pokedex")
	for _, pokemon := range cfg.Pokedex {
		fmt.Printf("- %s\n", pokemon.Name)
	}
}

func initializeCommands(cache *pokecache.Cache) {
	commands = map[string]cliCommand{
		"exit": {
			name:        "exit",
			description: "Exit the Pokedex",
			callback: func(args []string) error {
				return commandExit()
			},
		},
		"help": {
			name:        "help",
			description: "Displays a help message",
			callback: func(args []string) error {
				return commandHelp()
			},
		},
		"map": {
			name:        "map",
			description: "Displays names of 20 location areas in the Pokemon world",
			callback: func(args []string) error {
				return commandMap(&config, cache)
			},
		},
		"mapb": {
			name:        "mapb",
			description: "Displays names of previous 20 location areas in the Pokemon world",
			callback: func(args []string) error {
				return commandMapb(&config, cache)
			},
		},
		"explore": {
			name:        "explore",
			description: "Displays list of pokemon in a given location",
			callback: func(args []string) error {
				return commandExplore(&config, cache, args[1])
			},
		},
		"catch": {
			name:        "catch",
			description: "Adds pokemon to user's Pokedex",
			callback: func(args []string) error {
				return commandCatch(&config, cache, args[1])
			},
		},
		"inspect": {
			name:        "inspect",
			description: "Prints the name, height, weight, stats and type(s) of the Pokemon",
			callback: func(args []string) error {
				if len(args) < 2 {
					return fmt.Errorf("please provde a pokemon name")
				}
				commandInspect(&config, args[1])
				return nil
			},
		},
		"pokedex": {
			name:        "pokdex",
			description: "Prints list of pokemon caught",
			callback: func(args []string) error {
				commandPokedex(&config)
				return nil
			},
		},
	}
}

func main() {
	config = Config{
		Pokedex: make(map[string]Pokemon),
	}
	cache := pokecache.NewCache(5 * time.Second)

	initializeCommands(cache)
	scanner := bufio.NewScanner(os.Stdin)

	for {
		fmt.Print("Pokedex >")
		scanner.Scan()
		userInput := scanner.Text()
		trimmed := strings.TrimSpace(userInput)
		lowercase := strings.ToLower(trimmed)
		slices := strings.Fields(lowercase)
		if len(slices) == 0 {
			continue
		}

		command := slices[0] // get the first word

		if cmd, ok := commands[command]; ok {
			if command == "explore" {
				if len(slices) < 2 {
					fmt.Println("Please provide a location name")
					continue
				}
				err := commandExplore(&config, cache, slices[1])
				if err != nil {
					fmt.Println(err)
				}
			} else {
				err := cmd.callback(slices)
				if err != nil {
					fmt.Println(err)
				}
			}
		}

	}
}
