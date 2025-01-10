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

// RegisterRoutes defines all routes for the application.
func RegisterRoutes(r *gin.Engine) {
	r.GET("/", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "All is well"})
	})
	
	// /collection endpoint
	r.GET("/collections", collections)

	// /collection/:collection_name endpoints
	r.GET("/collections/:collection_name", collection_detail)
	r.PUT("/collections/:collection_name", add_collection)
	r.DELETE("/collections/:collection_name", delete_collection)
	r.PATCH("/collections/:collection_name", update_collection)
	
	r.PUT("/data/:collection_name", add_data)
	r.GET("/data/:collection_name", get_data)
	r.DELETE("/data/:collection_name", delete_data)
}
