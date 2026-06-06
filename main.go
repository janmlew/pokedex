package main

import (
	"bufio"
	"fmt"
	"os"
	"time"

	"github.com/janmlew/pokedex/internal/pokecache"
)

// config holds state that persists across commands. It carries the pagination
// URLs for the location-area endpoint so that successive `map` calls walk
// forward (and `mapb` walks backward) through the pages, plus a shared cache
// reused across every request.
type config struct {
	nextLocationsURL *string
	prevLocationsURL *string
	cache            *pokecache.Cache
}

type cliCommand struct {
	name        string
	description string
	callback    func(cfg *config) error
}

func getCommands() map[string]cliCommand {
	return map[string]cliCommand{
		"help": {
			name:        "help",
			description: "Displays a help message",
			callback:    commandHelp,
		},
		"map": {
			name:        "map",
			description: "Displays the next 20 location areas",
			callback:    commandMap,
		},
		"mapb": {
			name:        "mapb",
			description: "Displays the previous 20 location areas",
			callback:    commandMapb,
		},
		"exit": {
			name:        "exit",
			description: "Exit the Pokedex",
			callback:    commandExit,
		},
	}
}

func main() {
	scanner := bufio.NewScanner(os.Stdin)
	commands := getCommands()
	cfg := &config{
		cache: pokecache.NewCache(5 * time.Second),
	}
	for {
		fmt.Print("Pokedex > ")
		// Scan blocks until the user presses enter. It returns false on EOF,
		// which happens when the user presses Ctrl+D (or stdin is piped and
		// runs out). In that case we break out of the loop to exit cleanly,
		// rather than spinning forever on empty input.
		if !scanner.Scan() {
			break
		}
		words := cleanInput(scanner.Text())
		if len(words) == 0 {
			continue
		}
		commandName := words[0]
		command, ok := commands[commandName]
		if !ok {
			fmt.Println("Unknown command")
			continue
		}
		if err := command.callback(cfg); err != nil {
			fmt.Fprintln(os.Stderr, err)
		}
	}
	if err := scanner.Err(); err != nil {
		fmt.Fprintln(os.Stderr, "error reading input:", err)
	}
}

func commandExit(cfg *config) error {
	fmt.Println("Closing the Pokedex... Goodbye!")
	os.Exit(0)
	return nil
}

func commandHelp(cfg *config) error {
	fmt.Println("Welcome to the Pokedex!")
	fmt.Println("Usage:")
	fmt.Println()
	for _, command := range getCommands() {
		fmt.Printf("%s: %s\n", command.name, command.description)
	}
	fmt.Println()
	fmt.Println("You can also press Ctrl+D to exit at any time.")
	return nil
}

func commandMap(cfg *config) error {
	// On the first call nextLocationsURL is nil, so start at the canonical
	// first-page URL. Spelling out offset/limit matters: it's the exact URL
	// the API's "previous" pointer uses for page 1, so a later `mapb` back to
	// this page reuses the same cache key and hits the cache. On later calls
	// we follow the "next" URL the API handed us.
	url := "https://pokeapi.co/api/v2/location-area?offset=0&limit=20"
	if cfg.nextLocationsURL != nil {
		url = *cfg.nextLocationsURL
	}

	resp, err := fetchLocationAreas(url, cfg.cache)
	if err != nil {
		return err
	}

	// Remember the pagination cursors for subsequent map/mapb calls.
	cfg.nextLocationsURL = resp.Next
	cfg.prevLocationsURL = resp.Previous

	for _, area := range resp.Results {
		fmt.Println(area.Name)
	}
	return nil
}

func commandMapb(cfg *config) error {
	// A nil prevLocationsURL means there is no earlier page, either because
	// we've never paged forward or because we're back at the start.
	if cfg.prevLocationsURL == nil {
		fmt.Println("you're on the first page")
		return nil
	}

	resp, err := fetchLocationAreas(*cfg.prevLocationsURL, cfg.cache)
	if err != nil {
		return err
	}

	cfg.nextLocationsURL = resp.Next
	cfg.prevLocationsURL = resp.Previous

	for _, area := range resp.Results {
		fmt.Println(area.Name)
	}
	return nil
}
