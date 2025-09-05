#!/bin/bash

check_github_auth() {
    echo "Checking GitHub authentication..."
    if ! gh auth status --hostname github.com; then
        echo "Error: GitHub authentication required. Please run 'gh auth login'"
        exit 1
    fi
}

get_repo_info() {
    read -p "Enter Owner: " owner
    read -p "Enter Repository Name: " repo

    if [ -z "$owner" ] || [ -z "$repo" ]; then
        echo "Owner or Repository name cannot be empty. Exiting..."
        exit 1
    fi
}

delete_workflow_runs() {
    local owner=$1
    local repo=$2
    
    # Get total workflow count
    local total_runs=$(gh api -X GET "/repos/$owner/$repo/actions/runs?per_page=1" | jq '.total_count')
    echo "Found $total_runs workflow runs to delete"

    local current=1
    local page=1
    
    while ((current <= total_runs)); do
        echo "Processing page $page..."
        
        local temp_file=$(mktemp)
        gh api -X GET "/repos/$owner/$repo/actions/runs?per_page=100&page=$page" | \
        jq -r '.workflow_runs[].id' > "$temp_file"
        
        while IFS= read -r run_id; do
            if [ ! -z "$run_id" ]; then
                echo "Deleting: Workflow $current of $total_runs (ID: $run_id)"
                gh api --silent -X DELETE "/repos/$owner/$repo/actions/runs/$run_id"
                ((current++))
            fi
        done < "$temp_file"
        
        rm -f "$temp_file"
        ((page++))
    done
}

main() {
    check_github_auth
    get_repo_info
    delete_workflow_runs "$owner" "$repo"
    echo "All workflow runs deleted successfully!"
}

main