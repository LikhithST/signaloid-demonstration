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
	if len(os.Args) < 4 {
		fmt.Println("Usage: go run create-task.go <cfilename> <min_val> <max_val>")
		os.Exit(1)
	}
	cFileName := os.Args[1]
	minVal, err := strconv.Atoi(os.Args[2])
	if err != nil || minVal < 1 {
		log.Fatalf("Invalid min_val (must be >= 1): %v", os.Args[2])
	}
	maxVal, err := strconv.Atoi(os.Args[3])
	if err != nil || maxVal < minVal {
		log.Fatalf("Invalid max_val (must be >= min_val): %v", os.Args[3])
	}

	apiKey := os.Getenv("API_KEY")
	if apiKey == "" {
		log.Fatal("Error: API_KEY environment variable is not set")
	}

	// Read build-id from JSON
	listFile := "build_id_list.json"
	content, err := os.ReadFile(listFile)
	if err != nil {
		log.Fatalf("Failed to read %s: %v", listFile, err)
	}

	type BuildInfo struct {
		Time    string `json:"time"`
		BuildID string `json:"build-id"`
		Uxhw    bool   `json:"uxhw"`
	}
	builds := make(map[string]BuildInfo)
	if err := json.Unmarshal(content, &builds); err != nil {
		log.Fatalf("Failed to parse %s: %v", listFile, err)
	}

	baseFileName := filepath.Base(cFileName)
	buildInfo, ok := builds[baseFileName]
	if !ok {
		log.Fatalf("No build ID found for %s in %s", baseFileName, listFile)
	}
	buildID := buildInfo.BuildID
	fmt.Printf("Found Build ID: %s for %s\n", buildID, baseFileName)

	client := &http.Client{}
	resultsFile := "task-results.json"

	for currentVal := minVal; currentVal <= maxVal; currentVal *= 10 {
		var req *http.Request
		var err error

		// 1. Submit Task
		if buildInfo.Uxhw {
			fmt.Printf("\n--- Submitting Task (UxHw enabled, no arguments) ---\n")
			req, err = http.NewRequest("POST", baseURL+"/builds/"+buildID+"/tasks", nil)
		} else {
			taskArgs := strconv.Itoa(currentVal)
			fmt.Printf("\n--- Submitting Task with Argument: %s ---\n", taskArgs)

			taskPayload := map[string]interface{}{
				"Arguments":   taskArgs,
				"DataSources": []interface{}{},
			}
			payloadBytes, _ := json.Marshal(taskPayload)
			req, err = http.NewRequest("POST", baseURL+"/builds/"+buildID+"/tasks", bytes.NewBuffer(payloadBytes))
		}

		if err != nil {
			log.Printf("Failed to create task request: %v\n", err)
			continue
		}
		req.Header.Set("Authorization", apiKey)
		if !buildInfo.Uxhw {
			req.Header.Set("Content-Type", "application/json")
		}

		resp, err := client.Do(req)
		if err != nil {
			log.Printf("Task submission request failed: %v\n", err)
			continue
		}

		if resp.StatusCode >= 300 {
			log.Printf("Task submission failed with status: %s\n", resp.Status)
			resp.Body.Close()
			continue
		}

		var taskResp map[string]interface{}
		if err := json.NewDecoder(resp.Body).Decode(&taskResp); err != nil {
			log.Printf("Failed to decode task response: %v\n", err)
			resp.Body.Close()
			continue
		}
		resp.Body.Close()

		taskID, ok := taskResp["TaskID"].(string)
		if !ok {
			log.Printf("Failed to get TaskID from response: %v\n", taskResp)
			continue
		}
		fmt.Printf("Task ID: %s\n", taskID)

		// 2. Poll Task Status
		fmt.Println("Polling Task Status...")
		var finalData interface{}
		for {
			pollReq, _ := http.NewRequest("GET", baseURL+"/tasks/"+taskID, nil)
			pollReq.Header.Set("Authorization", apiKey)

			pollResp, err := client.Do(pollReq)
			if err != nil {
				log.Printf("Poll task failed: %v\n", err)
				finalData = map[string]string{"error": fmt.Sprintf("Poll failed: %v", err)}
				break
			}

			var statusResp map[string]interface{}
			json.NewDecoder(pollResp.Body).Decode(&statusResp)
			pollResp.Body.Close()

			taskStatus := statusResp["Status"].(string)
			fmt.Printf("Current Status: %s\n", taskStatus)

			if taskStatus == "Completed" || taskStatus == "Cancelled" || taskStatus == "Stopped" || taskStatus == "Failed" {
				if args, ok := statusResp["Arguments"]; ok {
					statusResp["iteration_value"] = args
					delete(statusResp, "Arguments")
				}
				finalData = statusResp
				break
			}
			time.Sleep(2 * time.Second)
		}

		// 3. Save to task-results.json
		taskResults := make(map[string]interface{})
		if resContent, err := os.ReadFile(resultsFile); err == nil {
			json.Unmarshal(resContent, &taskResults)
		}

		taskResults[taskID] = finalData

		if updatedRes, err := json.MarshalIndent(taskResults, "", "  "); err == nil {
			if err := os.WriteFile(resultsFile, updatedRes, 0644); err != nil {
				log.Printf("Warning: failed to write to %s: %v\n", resultsFile, err)
			}
		}
	}
	fmt.Printf("\nDone! Results saved to %s\n", resultsFile)
}
