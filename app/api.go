package app

import (
	"fmt"

	"github.com/gin-gonic/gin"
)

// RegisterRoutes defines all routes for the application.
func RegisterRoutes(r *gin.Engine) {
	r.GET("/", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "success"})
	})
	
	// /collection endpoint
	r.GET("/collections", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "Get all collections"})
	})

	// /collection/:collection_name endpoints
	r.GET("/collections/:collection_name", func(c *gin.Context) {
		collectionName := c.Param("collection_name")
		c.JSON(200, gin.H{"message": fmt.Sprintf("Get collection: %s", collectionName)})
	})
	r.DELETE("/collections/:collection_name", func(c *gin.Context) {
		collectionName := c.Param("collection_name")
		c.JSON(200, gin.H{"message": fmt.Sprintf("Delete collection: %s", collectionName)})
	})
	r.PUT("/collections/:collection_name", func(c *gin.Context) {
		collectionName := c.Param("collection_name")
		c.JSON(200, gin.H{"message": fmt.Sprintf("Update collection: %s", collectionName)})
	})
	r.PATCH("/collections/:collection_name", func(c *gin.Context) {
		collectionName := c.Param("collection_name")
		c.JSON(200, gin.H{"message": fmt.Sprintf("Patch collection: %s", collectionName)})
	})
}
