package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"
)

const (
	baseURL = "https://api.signaloid.io"
)

func main() {
	if len(os.Args) < 3 {
		fmt.Println("Usage: go run create-build.go <file_name.c> <uxhw: true|false>")
		os.Exit(1)
	}
	cFileName := os.Args[1]
	uxhw, err := strconv.ParseBool(os.Args[2])
	if err != nil {
		log.Fatalf("Invalid uxhw parameter (must be true or false): %v", err)
	}

	apiKey := os.Getenv("API_KEY")
	if apiKey == "" {
		log.Fatal("Error: API_KEY environment variable is not set")
	}

	coreID := os.Getenv("CORE_ID")
	if coreID == "" {
		log.Fatal("Error: CORE_ID environment variable is not set")
	}
	fmt.Printf("Using Core ID: %s\n", coreID)

	codeBytes, err := os.ReadFile(cFileName)
	if err != nil {
		log.Fatalf("Failed to read %s: %v", cFileName, err)
	}
	cCode := string(codeBytes)

	client := &http.Client{}

	fmt.Println("--- 1. Submitting Build ---")
	buildReqBody, _ := json.Marshal(map[string]string{
		"Code":     cCode,
		"Language": "C",
		"CoreID":   coreID,
	})

	req, _ := http.NewRequest("POST", baseURL+"/sourcecode/builds", bytes.NewBuffer(buildReqBody))
	req.Header.Set("Authorization", apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		log.Fatalf("Build request failed: %v", err)
	}
	defer resp.Body.Close()

	var buildResp map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&buildResp)
	buildID, ok := buildResp["BuildID"].(string)
	if !ok {
		log.Fatalf("Failed to create build. Response was: %v", buildResp)
	}
	fmt.Printf("Build ID: %s\n", buildID)

	// Update or append the build ID in the JSON list file
	const listFile = "build_id_list.json"
	currentTime := time.Now().Format(time.RFC3339)

	type BuildInfo struct {
		Time    string `json:"time"`
		BuildID string `json:"build-id"`
		Uxhw    bool   `json:"uxhw"`
	}

	builds := make(map[string]BuildInfo)

	// Read existing JSON file if it exists
	if content, err := os.ReadFile(listFile); err == nil {
		json.Unmarshal(content, &builds)
	}

	// Update the map (automatically handles insert or update)
	baseFileName := filepath.Base(cFileName)
	builds[baseFileName] = BuildInfo{
		Time:    currentTime,
		BuildID: buildID,
		Uxhw:    uxhw,
	}

	// Write back to the JSON file
	if updatedContent, err := json.MarshalIndent(builds, "", "  "); err == nil {
		if err := os.WriteFile(listFile, updatedContent, 0644); err != nil {
			log.Printf("Warning: failed to write to %s: %v", listFile, err)
		}
	} else {
		log.Printf("Warning: failed to serialize build info: %v", err)
	}

	fmt.Println("--- 2. Polling Build Status ---")
	for {
		req, _ = http.NewRequest("GET", baseURL+"/builds/"+buildID, nil)
		req.Header.Set("Authorization", apiKey)
		resp, err = client.Do(req)
		if err != nil {
			log.Fatalf("Poll build failed: %v", err)
		}

		var statusResp map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&statusResp)
		resp.Body.Close()

		buildStatus := statusResp["Status"].(string)
		fmt.Printf("Current Status: %s\n", buildStatus)

		if buildStatus == "Completed" {
			fmt.Println("Build completed successfully!")
			break
		} else if buildStatus == "Cancelled" || buildStatus == "Stopped" || buildStatus == "Failed" {
			log.Fatalf("Build failed with terminal state: %s", buildStatus)
		}
		time.Sleep(2 * time.Second)
	}
}
