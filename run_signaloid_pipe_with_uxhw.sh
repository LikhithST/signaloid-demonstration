#!/bin/bash

# Configuration
# Replace with your actual Signaloid API Key
# API_KEY="YOUR_SIGNALOID_API_KEY"
CORE_ID="cor_9a3efb0094405df5aeb61cf1f29606a0"
BASE_URL="https://api.signaloid.io"

# Helper function to extract JSON values using python3
parse_json() {
    python3 -c "import sys, json; print(json.load(sys.stdin)['$1'])"
}

echo "--- 1. Submitting Build ---"
BUILD_RESPONSE=$(curl -s -X POST "$BASE_URL/sourcecode/builds" \
    -H "Authorization: $API_KEY" \
    -H "Content-Type: application/json" \
    -d "{
        \"Code\": \"#include <stdio.h>\\n#include <uxhw.h>\\n\\nint main() {\\n    double principal = 100000.0;\\n\\n    // We define the market return as a known distribution of possibilities.\\n    // The hardware will propagate this uncertainty through the formula.\\n    double daily_return = UxHwDoubleUniformDist(0.05, 0.07);\\n\\n    // One single calculation, zero loops.\\n    double final_value = principal * (1 + daily_return);\\n\\n    // The output is the entire probability distribution of the result.\\n    printf(\\\"Portfolio outcome distribution: %lf\\\\n\\\", final_value);\\n\\n    return 0;\\n}\",
        \"Language\": \"C\",
        \"CoreID\": \"$CORE_ID\"
    }")

BUILD_ID=$(echo "$BUILD_RESPONSE" | parse_json "BuildID")
echo "Build ID: $BUILD_ID"

# Poll Build Status
echo "--- 2. Polling Build Status ---"
while true; do
    BUILD_STATUS_RESPONSE=$(curl -s -H "Authorization: $API_KEY" "$BASE_URL/builds/$BUILD_ID")
    STATUS=$(echo "$BUILD_STATUS_RESPONSE" | parse_json "Status")
    echo "Current Status: $STATUS"
    
    if [ "$STATUS" == "Completed" ]; then
        break
    elif [ "$STATUS" == "Cancelled" ] || [ "$STATUS" == "Stopped" ]; then
        echo "Build terminal state reached: $STATUS"
        exit 1
    fi
    sleep 2
done

# Execute Task
echo "--- 3. Submitting Task ---"
TASK_RESPONSE=$(curl -s -X POST "$BASE_URL/builds/$BUILD_ID/tasks" \
    -H "Authorization: $API_KEY")
TASK_ID=$(echo "$TASK_RESPONSE" | parse_json "TaskID")
echo "Task ID: $TASK_ID"

# Poll Task Status
echo "--- 4. Polling Task Status ---"
while true; do
    TASK_STATUS_RESPONSE=$(curl -s -H "Authorization: $API_KEY" "$BASE_URL/tasks/$TASK_ID")
    STATUS=$(echo "$TASK_STATUS_RESPONSE" | parse_json "Status")
    echo "Current Status: $STATUS"
    
    if [ "$STATUS" == "Completed" ]; then
        break
    elif [ "$STATUS" == "Cancelled" ] || [ "$STATUS" == "Stopped" ]; then
        echo "Task terminal state reached: $STATUS"
        exit 1
    fi
    sleep 2
done

# Fetch Outputs
echo "--- 5. Retrieving Output ---"
OUTPUT_RESPONSE=$(curl -s -H "Authorization: $API_KEY" "$BASE_URL/tasks/$TASK_ID/outputs")
OUTPUT_URL=$(echo "$OUTPUT_RESPONSE" | parse_json "Stdout")

echo "Resulting Output:"
curl -s "$OUTPUT_URL"
echo ""
