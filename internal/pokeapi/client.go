package pokeapi

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/mannyzzle/pokedexcli/internal/pokecache"
)

/* ------------------------------------------------------------------ */
/*  HTTP client setup                                                 */
/* ------------------------------------------------------------------ */

const firstPage = "https://pokeapi.co/api/v2/location-area?offset=0&limit=20"

type Client struct{ hc *http.Client }

func NewClient() *Client {
	tr := &http.Transport{
		Proxy: nil,
		TLSClientConfig: &tls.Config{
			MinVersion: tls.VersionTLS13, // speak TLS 1.3 first
		},
	}
	return &Client{hc: &http.Client{
		Transport: tr,
		Timeout:   15 * time.Second,
	}}
}

/* ------------------------------------------------------------------ */
/*  Response models                                                   */
/* ------------------------------------------------------------------ */

type page struct {
	Results  []struct{ Name string `json:"name"` }
	Next     *string `json:"next"`
	Previous *string `json:"previous"`
}

/* ------------------------------------------------------------------ */
/*  In-memory cache (5-second expiry)                                 */
/* ------------------------------------------------------------------ */

var cache = pokecache.NewCache(5 * time.Second)
func Cache() *pokecache.Cache { return cache }

/* ------------------------------------------------------------------ */
/*  Public API                                                        */
/* ------------------------------------------------------------------ */

// GetPage returns the list of location-area names plus next/prev URLs.
func (c *Client) GetPage(url string) (names []string, next, prev *string, err error) {
	if url == "" {
		url = firstPage
	}

	/* ---------- try cache first ---------- */
	if data, ok := cache.Get(url); ok {
		return parsePage(data)
	}

	/* ---------- make network request ---------- */
	resp, err := c.hc.Get(url)
	if err != nil {
		return nil, nil, nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, nil, nil, fmt.Errorf("pokeapi: status %s", resp.Status)
	}

	body, _ := io.ReadAll(resp.Body)
	cache.Add(url, body) // save copy for the next call
	return parsePage(body)
}

/* ------------------------------------------------------------------ */
/*  Helpers                                                           */
/* ------------------------------------------------------------------ */

func parsePage(b []byte) (names []string, next, prev *string, err error) {
	var p page
	if err = json.Unmarshal(b, &p); err != nil {
		return nil, nil, nil, err
	}
	for _, r := range p.Results {
		names = append(names, r.Name)
	}
	return names, p.Next, p.Previous, nil
}

func (c *Client) HC() *http.Client { return c.hc }
