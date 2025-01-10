package main

import (
	"fmt"
	"log"

	"github.com/asanihasan/sanDB/app"
)

func main() {

	// Load configuration
	config, err := app.LoadConfig()
	if err != nil {
		log.Fatalf("Error loading configuration: %v", err)
	}

	// Use the loaded configuration
	fmt.Printf("Server Port: %d\n", config.Server.Port)

	// Example: Start server using the loaded port
	fmt.Printf("Starting server on port %d...\n", config.Server.Port)
}
