package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/janmlew/pokedex/internal/pokecache"
)

// locationAreasResp mirrors the JSON returned by the PokeAPI
// /location-area endpoint. Next and Previous are pointers because the API
// sends JSON null for them at the ends of the list, and a *string lets us
// represent "no more pages" as nil rather than an empty string.
type locationAreasResp struct {
	Count    int     `json:"count"`
	Next     *string `json:"next"`
	Previous *string `json:"previous"`
	Results  []struct {
		Name string `json:"name"`
		URL  string `json:"url"`
	} `json:"results"`
}

// Pokemon mirrors the parts of the PokeAPI /pokemon/{name} response we care
// about. BaseExperience drives the catch difficulty; the height/weight/stats/
// types fields are modeled now so the later `inspect` command can reuse them.
type Pokemon struct {
	Name           string `json:"name"`
	BaseExperience int    `json:"base_experience"`
	Height         int    `json:"height"`
	Weight         int    `json:"weight"`
	Stats          []struct {
		BaseStat int `json:"base_stat"`
		Stat     struct {
			Name string `json:"name"`
		} `json:"stat"`
	} `json:"stats"`
	Types []struct {
		Type struct {
			Name string `json:"name"`
		} `json:"type"`
	} `json:"types"`
}

// fetchPokemon returns the details for a single named Pokemon. Results are
// cached by URL, so re-fetching the same Pokemon avoids a network round trip.
func fetchPokemon(name string, cache *pokecache.Cache) (Pokemon, error) {
	url := "https://pokeapi.co/api/v2/pokemon/" + name

	body, err := getCachedBody(url, cache)
	if err != nil {
		return Pokemon{}, err
	}

	var data Pokemon
	if err := json.Unmarshal(body, &data); err != nil {
		return Pokemon{}, fmt.Errorf("decoding pokemon %q: %w", name, err)
	}
	return data, nil
}

// locationAreaResp mirrors the JSON returned by the PokeAPI
// /location-area/{name} endpoint. We only model the fields we use: the area
// name and the list of Pokemon that can be encountered there.
type locationAreaResp struct {
	Name              string `json:"name"`
	PokemonEncounters []struct {
		Pokemon struct {
			Name string `json:"name"`
			URL  string `json:"url"`
		} `json:"pokemon"`
	} `json:"pokemon_encounters"`
}

// fetchLocationArea returns the details for a single named location area,
// including which Pokemon can be encountered there. Results are cached by URL.
func fetchLocationArea(name string, cache *pokecache.Cache) (locationAreaResp, error) {
	url := "https://pokeapi.co/api/v2/location-area/" + name

	body, err := getCachedBody(url, cache)
	if err != nil {
		return locationAreaResp{}, err
	}

	var data locationAreaResp
	if err := json.Unmarshal(body, &data); err != nil {
		return locationAreaResp{}, fmt.Errorf("decoding location area %q: %w", name, err)
	}
	return data, nil
}

// fetchLocationAreas returns the location-area page at url, decoded into a
// locationAreasResp. The raw response body is cached under url, so a repeated
// request for the same page is served from the cache instead of the network.
func fetchLocationAreas(url string, cache *pokecache.Cache) (locationAreasResp, error) {
	body, err := getCachedBody(url, cache)
	if err != nil {
		return locationAreasResp{}, err
	}

	var data locationAreasResp
	if err := json.Unmarshal(body, &data); err != nil {
		return locationAreasResp{}, fmt.Errorf("decoding location areas: %w", err)
	}
	return data, nil
}

// getCachedBody returns the raw response body for url. On a cache hit it
// returns the stored bytes without touching the network; on a miss it performs
// the GET, stores the body in the cache, and returns it.
func getCachedBody(url string, cache *pokecache.Cache) ([]byte, error) {
	if body, ok := cache.Get(url); ok {
		log.Printf("cache hit: %s", url)
		return body, nil
	}
	log.Printf("cache miss: %s", url)

	res, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("requesting %s: %w", url, err)
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response body: %w", err)
	}

	if res.StatusCode > 299 {
		return nil, fmt.Errorf("request to %s failed with status %d: %s", url, res.StatusCode, body)
	}

	cache.Add(url, body)
	return body, nil
}
