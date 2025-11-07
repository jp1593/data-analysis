package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// Graph and Struct/Methods 

type Graph struct {
	nodes  map[int]string  // Actor ID (Name)
	edges  map[[2]int]bool // Unique edges
	degree map[int]int     // Node counter
}

func NewGraph() *Graph {
	return &Graph{
		nodes:  make(map[int]string),
		edges:  make(map[[2]int]bool),
		degree: make(map[int]int),
	}
}

func (g *Graph) AddNode(id int, name string) {
	if _, exists := g.nodes[id]; !exists {
		g.nodes[id] = name
	}
}

func (g *Graph) AddEdge(a, b int) {
	if a == b {
		return
	}
	key := [2]int{min(a, b), max(a, b)}
	if !g.edges[key] {
		g.edges[key] = true
		g.degree[a]++
		g.degree[b]++
	}
}

func (g *Graph) TotalNodes() int {
	return len(g.nodes)
}

func (g *Graph) TotalEdges() int {
	return len(g.edges)
}

func (g *Graph) MaxDegreeNodes() map[int]string {
	maxDeg := 0
	for _, d := range g.degree {
		if d > maxDeg {
			maxDeg = d
		}
	}
	result := make(map[int]string)
	for id, d := range g.degree {
		if d == maxDeg {
			result[id] = g.nodes[id]
		}
	}
	return result
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// TMDB Api Utils
const baseURL = "https://api.themoviedb.org/3"

type TMDBAPIUtils struct {
	APIKey string
}

type CastMember struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Order int    `json:"order"`
}

type MovieCredit struct {
	ID          int    `json:"id"`
	Title       string `json:"title"`
	ReleaseDate string `json:"release_date"`
}

func (api *TMDBAPIUtils) GetMovieCast(movieID string, limit int, excludeIDs []int) ([]CastMember, error) {
	url := fmt.Sprintf("%s/movie/%s/credits?api_key=%s", baseURL, movieID, api.APIKey)
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result struct {
		Cast []CastMember `json:"cast"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	exclude := make(map[int]bool)
	for _, id := range excludeIDs {
		exclude[id] = true
	}

	finalCast := []CastMember{}
	for _, member := range result.Cast {
		if exclude[member.ID] {
			continue
		}
		if len(finalCast) < limit {
			finalCast = append(finalCast, member)
		}
	}
	return finalCast, nil
}

func (api *TMDBAPIUtils) GetMovieCreditsForPerson(personID, startDate, endDate string) ([]MovieCredit, error) {
	url := fmt.Sprintf("%s/person/%s/movie_credits?api_key=%s", baseURL, personID, api.APIKey)
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result struct {
		Cast []MovieCredit `json:"cast"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	layout := "2006-01-02"
	start, _ := time.Parse(layout, startDate)
	end, _ := time.Parse(layout, endDate)

	filtered := []MovieCredit{}
	for _, movie := range result.Cast {
		date, err := time.Parse(layout, movie.ReleaseDate)
		if err == nil && date.After(start) && date.Before(end) {
			filtered = append(filtered, movie)
		}
	}
	return filtered, nil
}

// Main Program

func main() {
	api := TMDBAPIUtils{
		APIKey: "0c0d530e2177a2657d1a349b854bbee4",
	}

	// Laurence Fishburne (ID)
	fishburneID := 2975

	fmt.Println("Fetching Laurence Fishburne movie credits...")
	credits, err := api.GetMovieCreditsForPerson(fmt.Sprint(fishburneID), "2000-01-01", "2025-12-31")
	if err != nil {
		panic(err)
	}

	graph := NewGraph()
	graph.AddNode(fishburneID, "Laurence Fishburne")

	for _, movie := range credits {
		cast, err := api.GetMovieCast(fmt.Sprint(movie.ID), 5, []int{fishburneID})
		if err != nil {
			continue
		}
		for _, actor := range cast {
			graph.AddNode(actor.ID, actor.Name)
			graph.AddEdge(fishburneID, actor.ID)
		}
	}

	fmt.Println("\n=== Collaboration Graph for Laurence Fishburne ===")
	fmt.Printf("Total Nodes: %d\n", graph.TotalNodes())
	fmt.Printf("Total Edges: %d\n", graph.TotalEdges())
	fmt.Println("Max Degree Nodes:")
	for id, name := range graph.MaxDegreeNodes() {
		fmt.Printf("  %s (ID %d)\n", name, id)
	}
}
