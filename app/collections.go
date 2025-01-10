package app

import (
	"fmt"
	"os"

	"github.com/gin-gonic/gin"
)

func collections(c *gin.Context) {
	dataPath := "./data" // Path to the data directory

	// Open the directory
	files, err := os.ReadDir(dataPath)
	if err != nil {
		c.JSON(500, gin.H{"error": fmt.Sprintf("Failed to read data directory: %v", err)})
		return
	}

	// Collect folder names
	collections := []string{}
	for _, file := range files {
		if file.IsDir() {
			collections = append(collections, file.Name())
		}
	}

	// Return the list of collections as JSON
	c.JSON(200, gin.H{"collections": collections})
}

func collection_detail(c *gin.Context) {
	collectionName := c.Param("collection_name")
	dataPath := "./data" // Path to the data directory

	// Check if the collection directory exists
	collectionPath := fmt.Sprintf("%s/%s", dataPath, collectionName)
	info, err := os.Stat(collectionPath)
	if os.IsNotExist(err) || !info.IsDir() {
		c.JSON(404, gin.H{"error": fmt.Sprintf("Collection '%s' does not exist", collectionName)})
		return
	}

	c.JSON(200, gin.H{"message": fmt.Sprintf("Collection '%s' exists", collectionName)})
}

func add_collection(c *gin.Context) {
	collectionName := c.Param("collection_name")
	dataPath := "./data" // Path to the data directory

	// Create the collection directory if it doesn't exist
	collectionPath := fmt.Sprintf("%s/%s", dataPath, collectionName)
	if _, err := os.Stat(collectionPath); os.IsNotExist(err) {
		if err := os.Mkdir(collectionPath, os.ModePerm); err != nil {
			c.JSON(500, gin.H{"error": fmt.Sprintf("Failed to create collection '%s': %v", collectionName, err)})
			return
		}
		c.JSON(201, gin.H{"message": fmt.Sprintf("Collection '%s' created", collectionName)})
		return
	}

	c.JSON(200, gin.H{"message": fmt.Sprintf("Collection '%s' already exists", collectionName)})
}

func delete_collection(c *gin.Context) {
	collectionName := c.Param("collection_name")
	dataPath := "./data" // Path to the data directory

	// Delete the collection directory
	collectionPath := fmt.Sprintf("%s/%s", dataPath, collectionName)
	if _, err := os.Stat(collectionPath); os.IsNotExist(err) {
		c.JSON(404, gin.H{"error": fmt.Sprintf("Collection '%s' does not exist", collectionName)})
		return
	}

	if err := os.RemoveAll(collectionPath); err != nil {
		c.JSON(500, gin.H{"error": fmt.Sprintf("Failed to delete collection '%s': %v", collectionName, err)})
		return
	}

	c.JSON(200, gin.H{"message": fmt.Sprintf("Collection '%s' deleted successfully", collectionName)})
}

func update_collection(c *gin.Context) {
	oldName := c.Param("collection_name")
	newName := c.Query("new_name")
	dataPath := "./data" // Path to the data directory

	if oldName == "" || newName == "" {
		c.JSON(400, gin.H{"error": "Both 'old_name' and 'new_name' must be provided"})
		return
	}

	oldPath := fmt.Sprintf("%s/%s", dataPath, oldName)
	newPath := fmt.Sprintf("%s/%s", dataPath, newName)

	// Check if the old collection exists
	if _, err := os.Stat(oldPath); os.IsNotExist(err) {
		c.JSON(404, gin.H{"error": fmt.Sprintf("Collection '%s' does not exist", oldName)})
		return
	}

	// Check if the new  collection name already exists
	if _, err := os.Stat(newPath); err == nil {
		c.JSON(400, gin.H{"error": fmt.Sprintf("Collection '%s' already exists", newName)})
		return
	}

	// Rename the collection
	if err := os.Rename(oldPath, newPath); err != nil {
		c.JSON(500, gin.H{"error": fmt.Sprintf("Failed to rename collection '%s' to '%s': %v", oldName, newName, err)})
		return
	}

	c.JSON(200, gin.H{"message": fmt.Sprintf("Collection '%s' renamed to '%s'", oldName, newName)})
}