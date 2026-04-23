package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
)

const baseURL = "https://api.signaloid.io"

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run fetch-outputs.go <cfilename>")
		os.Exit(1)
	}
	cFileName := os.Args[1]

	apiKey := os.Getenv("API_KEY")
	if apiKey == "" {
		log.Fatal("Error: API_KEY environment variable is not set")
	}

	// Read build-id from JSON to filter tasks
	listFile := "build_id_list.json"
	listContent, err := os.ReadFile(listFile)
	if err != nil {
		log.Fatalf("Failed to read %s: %v", listFile, err)
	}

	type BuildInfo struct {
		Time    string `json:"time"`
		BuildID string `json:"build-id"`
		Uxhw    bool   `json:"uxhw"`
	}
	builds := make(map[string]BuildInfo)
	if err := json.Unmarshal(listContent, &builds); err != nil {
		log.Fatalf("Failed to parse %s: %v", listFile, err)
	}

	baseFileName := filepath.Base(cFileName)
	buildInfo, ok := builds[baseFileName]
	if !ok {
		log.Fatalf("No build ID found for %s in %s", baseFileName, listFile)
	}
	targetBuildID := buildInfo.BuildID
	isUxhw := buildInfo.Uxhw
	fmt.Printf("Filtering tasks for Build ID: %s (%s, uxhw: %v)\n", targetBuildID, baseFileName, isUxhw)

	// Read the task-results.json file
	resultsFile := "task-results.json"
	content, err := os.ReadFile(resultsFile)
	if err != nil {
		log.Fatalf("Failed to read %s: %v", resultsFile, err)
	}

	var taskResults map[string]map[string]interface{}
	if err := json.Unmarshal(content, &taskResults); err != nil {
		log.Fatalf("Failed to parse %s: %v", resultsFile, err)
	}

	client := &http.Client{}
	var finalData []map[string]interface{}

	for taskID, taskInfo := range taskResults {
		fmt.Printf("Fetching output for Task ID: %s\n", taskID)

		// Extract BuildID and Stats
		buildID, _ := taskInfo["BuildID"].(string)
		if buildID != targetBuildID {
			continue // Skip tasks that don't belong to the specified C file
		}

		stats, _ := taskInfo["Stats"].(map[string]interface{})
		iterationValue, ok := taskInfo["iteration_value"].(string)
		if !ok {
			iterationValue, _ = taskInfo["Arguments"].(string)
		}

		// Call the outputs API
		req, err := http.NewRequest("GET", baseURL+"/tasks/"+taskID+"/outputs", nil)
		if err != nil {
			log.Printf("Failed to create request for %s: %v\n", taskID, err)
			continue
		}
		req.Header.Set("Authorization", apiKey)

		resp, err := client.Do(req)
		if err != nil {
			log.Printf("Failed to fetch outputs for %s: %v\n", taskID, err)
			continue
		}

		var outputResp map[string]interface{}
		if err := json.NewDecoder(resp.Body).Decode(&outputResp); err != nil {
			log.Printf("Failed to decode output response for %s: %v\n", taskID, err)
			resp.Body.Close()
			continue
		}
		resp.Body.Close()

		outputContent := ""
		if stdoutURL, ok := outputResp["Stdout"].(string); ok && stdoutURL != "" {
			outResp, err := http.Get(stdoutURL)
			if err == nil {
				b, _ := io.ReadAll(outResp.Body)
				outputContent = string(b)
				outResp.Body.Close()
			} else {
				log.Printf("Failed to download stdout content for %s: %v\n", taskID, err)
			}
		}

		// Create the combined object
		finalEntry := map[string]interface{}{
			"buildID":         buildID,
			"taskId":          taskID,
			"Stats":           stats,
			"output":          outputContent,
			"uxhw":            isUxhw,
			"iteration_value": iterationValue,
		}
		finalData = append(finalData, finalEntry)
	}

	// Save to final-outputs.json
	finalFile := "final-outputs.json"
	var existingData []map[string]interface{}
	if existingContent, err := os.ReadFile(finalFile); err == nil {
		json.Unmarshal(existingContent, &existingData)
	}

	for _, newEntry := range finalData {
		bID, _ := newEntry["buildID"].(string)
		tID, _ := newEntry["taskId"].(string)

		found := false
		for i, existingEntry := range existingData {
			exBID, _ := existingEntry["buildID"].(string)
			exTID, _ := existingEntry["taskId"].(string)
			if bID == exBID && tID == exTID {
				existingData[i] = newEntry // Update existing entry
				found = true
				break
			}
		}
		if !found {
			existingData = append(existingData, newEntry)
		}
	}

	finalJSON, _ := json.MarshalIndent(existingData, "", "  ")
	os.WriteFile(finalFile, finalJSON, 0644)

	fmt.Printf("\nDone! Combined output saved to %s\n", finalFile)
}
