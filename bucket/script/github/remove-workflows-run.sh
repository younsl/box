#!/bin/bash

# Interactive mode: Prompt user for Owner and Repository Name
read -p "Enter Owner: " OWNER
read -p "Enter Repository Name: " REPOSITORY

# Check if OWNER and REPOSITORY are not empty
if [ -z "$OWNER" ] || [ -z "$REPOSITORY" ]; then
  echo "Owner or Repository Name cannot be empty. Exiting..."
  exit 1
fi

# repeat command 30 times, if there are 30 pages of workflow history 
for i in {1..30}; do
  gh api -X GET /repos/$OWNER/$REPOSITORY/actions/runs \
    | jq '.workflow_runs[] | .id' \
    | xargs -t -I{} gh api --silent -X DELETE /repos/$OWNER/$REPOSITORY/actions/runs/{};
done
