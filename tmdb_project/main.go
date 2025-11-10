package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"time"
	"github.com/awalterschulze/gographviz"
)

// Graph Struct 
type Graph struct {
	nodes      map[int]string
	edges      map[[2]int]string 
	degree     map[int]int
	lastMovie  map[[2]int]string
	lastDate   map[[2]int]time.Time
}

func NewGraph() *Graph {
	return &Graph{
		nodes:     make(map[int]string),
		edges:     make(map[[2]int]string),
		degree:    make(map[int]int),
		lastMovie: make(map[[2]int]string),
		lastDate:  make(map[[2]int]time.Time),
	}
}

func (g *Graph) AddNode(id int, name string) {
	if _, exists := g.nodes[id]; !exists {
		g.nodes[id] = name
	}
}

func (g *Graph) AddEdge(a, b int, movie string, date time.Time) {
	if a == b {
		return
	}
	key := [2]int{min(a, b), max(a, b)}

	if _, exists := g.edges[key]; !exists {
		g.edges[key] = movie
		g.lastDate[key] = date
		g.degree[a]++
		g.degree[b]++
	} else {
		if date.After(g.lastDate[key]) {
			g.edges[key] = movie
			g.lastDate[key] = date
		}
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

// TMDB API 
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


func main() {
	api := TMDBAPIUtils{
		APIKey: "0c0d530e2177a2657d1a349b854bbee4",
	}

	actorID := 2975 // Lawrence Fishburne
	fmt.Println("Fetching Lawrence Fishburne movie credits...")

	credits, err := api.GetMovieCreditsForPerson(fmt.Sprint(actorID), "2000-01-01", "2025-12-31")
	if err != nil {
		panic(err)
	}

	graph := NewGraph()
	graph.AddNode(actorID, "Lawrence Fishburne")

	layout := "2006-01-02"

	for _, movie := range credits {
		cast, err := api.GetMovieCast(fmt.Sprint(movie.ID), 5, []int{actorID})
		if err != nil {
			continue
		}

		date, _ := time.Parse(layout, movie.ReleaseDate)

		for _, actor := range cast {
			graph.AddNode(actor.ID, actor.Name)
			graph.AddEdge(actorID, actor.ID, movie.Title, date)
		}
	}

	// --- Stats ---
	fmt.Println("\n=== Collaboration Graph for Lawrence Fishburne ===")
	fmt.Printf("Total Nodes: %d\n", graph.TotalNodes())
	fmt.Printf("Total Edges: %d\n", graph.TotalEdges())
	fmt.Println("Max Degree Nodes:")
	for id, name := range graph.MaxDegreeNodes() {
		fmt.Printf("  %s (ID %d)\n", name, id)
	}

	// Visualization 
	GenerateGraphViz(graph)
}

// Graph Visualization 

func GenerateGraphViz(g *Graph) {
	graphAst, _ := gographviz.ParseString(`digraph G {}`)
	graphObj := gographviz.NewGraph()
	gographviz.Analyse(graphAst, graphObj)

	graphObj.SetDir(false)
	graphObj.AddAttr("G", "splines", "true")
	graphObj.AddAttr("G", "overlap", "false")

	for id, name := range g.nodes {
		graphObj.AddNode("G", fmt.Sprintf("n%d", id), map[string]string{
			"label": fmt.Sprintf("\"%s\"", name),
			"shape": "ellipse",
			"style": "filled",
			"fillcolor": "\"#cce5ff\"",
		})
	}

	for edge, movie := range g.edges {
		src := fmt.Sprintf("n%d", edge[0])
		dst := fmt.Sprintf("n%d", edge[1])
		graphObj.AddEdge(src, dst, false, map[string]string{
			"label": fmt.Sprintf("\"%s\"", movie),
		})
	}

	dotOutput := "graph.dot"
	pngOutput := "graph.png"
	os.WriteFile(dotOutput, []byte(graphObj.String()), 0644)

	cmd := exec.Command("dot", "-Tpng", dotOutput, "-o", pngOutput)
	err := cmd.Run()
	if err != nil {
		fmt.Println("⚠️ Could not generate PNG:", err)
		return
	}
	fmt.Printf("\n✅ Graph visualization saved as %s\n", pngOutput)
}
