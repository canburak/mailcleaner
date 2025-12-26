// Package main provides the web server for mailcleaner
package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"

	"github.com/mailcleaner/mailcleaner/internal/api"
	"github.com/mailcleaner/mailcleaner/internal/storage"
)

func getPort() int {
	// Check PORT environment variable first (used by Render, Railway, etc.)
	if envPort := os.Getenv("PORT"); envPort != "" {
		if p, err := strconv.Atoi(envPort); err == nil {
			return p
		}
	}
	return 8080
}

func main() {
	port := flag.Int("port", getPort(), "port to listen on")
	dbPath := flag.String("db", "", "path to database file (default: ~/.mailcleaner/data.db)")
	staticDir := flag.String("static", "", "path to static files directory")
	flag.Parse()

	// Determine database path
	if *dbPath == "" {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			log.Fatalf("Failed to get home directory: %v", err)
		}
		dataDir := filepath.Join(homeDir, ".mailcleaner")
		if err := os.MkdirAll(dataDir, 0755); err != nil {
			log.Fatalf("Failed to create data directory: %v", err)
		}
		*dbPath = filepath.Join(dataDir, "data.db")
	}

	log.Printf("Using database: %s", *dbPath)

	// Initialize storage
	store, err := storage.New(*dbPath)
	if err != nil {
		log.Fatalf("Failed to initialize storage: %v", err)
	}
	defer store.Close()

	// Create API handler and router
	handler := api.NewHandler(store)
	router := api.NewRouter(handler)

	// Add WebSocket routes
	api.AddWebSocketRoutes(router, store)

	// Serve static files if directory provided
	if *staticDir != "" {
		log.Printf("Serving static files from: %s", *staticDir)
		fs := http.FileServer(http.Dir(*staticDir))
		router.Handle("/*", http.StripPrefix("/", fs))
	}

	// Start server
	addr := fmt.Sprintf(":%d", *port)
	log.Printf("Starting server on http://localhost%s", addr)
	log.Printf("API available at http://localhost%s/api", addr)
	log.Printf("WebSocket available at ws://localhost%s/ws/preview", addr)

	if err := http.ListenAndServe(addr, router); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}
