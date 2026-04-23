package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"
)

func main() {
	// Format the current date and time (e.g., 20231024_153000)
	dateTime := time.Now().Format("20060102_150405")
	targetDir := filepath.Join("history", dateTime)

	// Safely check for .json files
	jsonFiles, err := filepath.Glob("*.json")
	if err != nil {
		log.Fatalf("Error finding JSON files: %v", err)
	}

	if len(jsonFiles) == 0 {
		fmt.Println("No JSON files found in the current directory.")
		return
	}

	// Create the target directory and move the files
	if err := os.MkdirAll(targetDir, 0755); err != nil {
		log.Fatalf("Failed to create target directory: %v", err)
	}

	for _, file := range jsonFiles {
		if file != "build_id_list.json" {
			targetPath := filepath.Join(targetDir, filepath.Base(file))
			if err := os.Rename(file, targetPath); err != nil {
				log.Printf("Failed to move file %s: %v\n", file, err)
			}
		}

	}

	fmt.Printf("Successfully moved %d JSON file(s) to %s\n", len(jsonFiles), targetDir)
}
