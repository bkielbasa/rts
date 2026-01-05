// Command genmap generates a terrain map file
package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/bklimczak/tanks/engine/terrain"
)

func main() {
	var (
		width       = flag.Float64("width", 3840, "Map width in pixels")
		height      = flag.Float64("height", 2880, "Map height in pixels")
		seed        = flag.Int64("seed", 42, "Random seed for generation")
		output      = flag.String("output", "maps/default.yaml", "Output file path")
		name        = flag.String("name", "Default Map", "Map name")
		description = flag.String("desc", "Auto-generated map", "Map description")
		author      = flag.String("author", "System", "Map author")
	)
	flag.Parse()

	// Ensure output directory exists
	if err := os.MkdirAll("maps", 0755); err != nil {
		log.Fatalf("Failed to create maps directory: %v", err)
	}

	fmt.Printf("Generating map %dx%d pixels with seed %d...\n", int(*width), int(*height), *seed)

	// Generate the map
	m := terrain.NewMap(*width, *height)
	m.Generate(*seed)

	// Place some metal deposits near common starting positions
	m.PlaceMetalDeposit(400, 150)  // Near player start
	m.PlaceMetalDeposit(450, 150)
	m.PlaceMetalDeposit(3400, 2700) // Far corner

	// Save to file
	if err := terrain.SaveMapToFile(m, *output, *name, *description, *author); err != nil {
		log.Fatalf("Failed to save map: %v", err)
	}

	fmt.Printf("Map saved to %s\n", *output)
	fmt.Printf("  - %d x %d tiles\n", m.Width, m.Height)
	fmt.Printf("  - Tile size: %.0f pixels\n", terrain.TileSize)
	fmt.Printf("\nYou can now edit %s to customize the terrain.\n", *output)
}
