#!/bin/bash

# Script to find a valid checkpoint ID for testing

set -e

# Load environment
source ./scripts/setup-virtuoso.sh

echo "Finding a valid checkpoint for testing..."
echo "========================================"

# First, let's list projects to find a valid project
echo "1. Listing projects..."
PROJECTS_DATA=$(./bin/api-cli list-projects -o json)
PROJECT_ID=$(echo "$PROJECTS_DATA" | jq -r '.projects[0].id // empty')

if [ -z "$PROJECT_ID" ]; then
    echo "No projects found. Creating a test project..."
    # Generate unique project name with timestamp
    PROJECT_NAME="Step Commands Test Project $(date +%s)"
    CREATE_RESULT=$(./bin/api-cli create-project "$PROJECT_NAME" -o json)
    PROJECT_ID=$(echo "$CREATE_RESULT" | jq -r '.project_id')
    echo "Created project with ID: $PROJECT_ID"
else
    echo "Found existing project with ID: $PROJECT_ID"
fi

# Get goals for the project
echo "2. Getting goals for project $PROJECT_ID..."
GOALS_DATA=$(./bin/api-cli list-goals $PROJECT_ID -o json)
GOAL_ID=$(echo "$GOALS_DATA" | jq -r '.goals[0].id // empty')

if [ -z "$GOAL_ID" ]; then
    echo "No goals found. Creating a test goal..."
    CREATE_GOAL_RESULT=$(./bin/api-cli create-goal $PROJECT_ID "Step Commands Test Goal" --url "https://example.com" -o json)
    GOAL_ID=$(echo "$CREATE_GOAL_RESULT" | jq -r '.goal_id')
    echo "Created goal with ID: $GOAL_ID"
    
    # Wait for goal to be fully created
    sleep 2
    
    # Re-fetch goals to get snapshot ID
    GOALS_DATA=$(./bin/api-cli list-goals $PROJECT_ID -o json)
else
    echo "Found existing goal with ID: $GOAL_ID"
fi

# Get snapshot ID for the goal
echo "3. Getting snapshot ID for goal $GOAL_ID..."
SNAPSHOT_ID=$(echo "$GOALS_DATA" | jq -r --arg gid "$GOAL_ID" '.goals[] | select(.id == ($gid | tonumber)) | .defaultSnapshot // empty')

if [ -z "$SNAPSHOT_ID" ]; then
    # Try alternative field name
    SNAPSHOT_ID=$(echo "$GOALS_DATA" | jq -r --arg gid "$GOAL_ID" '.goals[] | select(.id == ($gid | tonumber)) | .default_snapshot // empty')
fi

if [ -z "$SNAPSHOT_ID" ]; then
    # Default to 1 if not found (common pattern in Virtuoso)
    echo "Warning: Could not find snapshot ID for goal $GOAL_ID, using default value 1"
    SNAPSHOT_ID=1
fi

echo "Found snapshot ID: $SNAPSHOT_ID"

# Get journeys for the goal
echo "4. Getting journeys for goal $GOAL_ID..."
JOURNEYS_DATA=$(./bin/api-cli list-journeys $GOAL_ID $SNAPSHOT_ID -o json)
JOURNEY_ID=$(echo "$JOURNEYS_DATA" | jq -r '.journeys[0].id // empty')

if [ -z "$JOURNEY_ID" ]; then
    echo "No journeys found. This should not happen as goals auto-create journeys."
    # Try creating a journey manually
    CREATE_JOURNEY_RESULT=$(./bin/api-cli create-journey $GOAL_ID $SNAPSHOT_ID "Test Journey" -o json)
    JOURNEY_ID=$(echo "$CREATE_JOURNEY_RESULT" | jq -r '.journey_id')
    if [ -z "$JOURNEY_ID" ]; then
        echo "Error: Could not create journey"
        exit 1
    fi
fi

echo "Found journey ID: $JOURNEY_ID"

# Get checkpoints for the journey
echo "5. Getting checkpoints for journey $JOURNEY_ID..."
CHECKPOINTS_DATA=$(./bin/api-cli list-checkpoints $JOURNEY_ID -o json)
CHECKPOINT_ID=$(echo "$CHECKPOINTS_DATA" | jq -r '.checkpoints[0].id // empty')

if [ -z "$CHECKPOINT_ID" ]; then
    echo "No checkpoints found. Creating a test checkpoint..."
    CREATE_CHECKPOINT_RESULT=$(./bin/api-cli create-checkpoint $JOURNEY_ID "Step Commands Test Checkpoint" --navigation "https://example.com/test" -o json)
    CHECKPOINT_ID=$(echo "$CREATE_CHECKPOINT_RESULT" | jq -r '.checkpoint_id')
    echo "Created checkpoint with ID: $CHECKPOINT_ID"
else
    echo "Found existing checkpoint with ID: $CHECKPOINT_ID"
fi

echo ""
echo "========================================"
echo "Test Environment Ready!"
echo "========================================"
echo "Project ID: $PROJECT_ID"
echo "Goal ID: $GOAL_ID"
echo "Snapshot ID: $SNAPSHOT_ID"
echo "Journey ID: $JOURNEY_ID"
echo "Checkpoint ID: $CHECKPOINT_ID"
echo ""
echo "To run the step command tests, use:"
echo "TEST_CHECKPOINT_ID=$CHECKPOINT_ID ./test-all-step-commands.sh"
echo ""
echo "Or export the checkpoint ID:"
echo "export TEST_CHECKPOINT_ID=$CHECKPOINT_ID"