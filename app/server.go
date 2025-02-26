package app

import (
	"fmt"
	"context"
	"net/http"
	"time"
	"os"
	"os/signal"
	"syscall"
	"github.com/gin-gonic/gin"
)

func StartServer() {
	addr := fmt.Sprintf(":%d", AppConfig.Server.Port)
	fmt.Printf("Starting Gin server on %s...\n", addr)

	r := gin.Default()
	r.Use(func(c *gin.Context) {
		token := c.GetHeader("Authorization")
		if token != AppConfig.Server.Token {
			c.AbortWithStatusJSON(401, gin.H{"error": "Unauthorized"})
			return
		}
		c.Next()
	})

	RegisterRoutes(r)

	// Create HTTP server with timeout settings
	srv := &http.Server{
		Addr:         addr,
		Handler:      r,
		ReadTimeout:  time.Duration(AppConfig.Server.Timeout.ReadTimeout) * time.Second,
		WriteTimeout: time.Duration(AppConfig.Server.Timeout.WriteTimeout) * time.Second,
		IdleTimeout:  time.Duration(AppConfig.Server.Timeout.IdleTimeout) * time.Second,
	}

	// Run the server in a Goroutine so it doesn’t block
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			fmt.Printf("Failed to start server: %v\n", err)
		}
	}()

	// Capture termination signals (CTRL+C, Docker Stop, Kubernetes SIGTERM)
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit // Wait for termination signal

	fmt.Println("\nShutting down server...")

	// Convert YAML shutdown-timeout value to duration
	shutdownTimeout := time.Duration(AppConfig.Server.ShutdownTimeout) * time.Second

	// Create a timeout context for shutdown
	ctx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		fmt.Printf("Server forced to shutdown: %v\n", err)
	}

	fmt.Println("Server gracefully stopped.")
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
