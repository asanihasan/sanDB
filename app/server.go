package app

import (
	"fmt"

	"github.com/gin-gonic/gin"
)

func StartServer() {
	config, err := LoadConfig()
	if err != nil {
		fmt.Printf("Error loading config: %v\n", err)
		return
	}

	addr := fmt.Sprintf(":%d", config.Server.Port)
	fmt.Printf("Starting Gin server on %s...\n", addr)

	// Initialize the Gin router
	r := gin.Default()

	// Middleware to check authorization token
	r.Use(func(c *gin.Context) {
		token := c.GetHeader("Authorization")
		if token != config.Server.Token {
			c.AbortWithStatusJSON(401, gin.H{"error": "Unauthorized"})
			return
		}
		c.Next()
	})

	// Define routes
	RegisterRoutes(r)

	// Start the server
	if err := r.Run(addr); err != nil {
		fmt.Printf("Failed to start server: %v\n", err)
	}
}
