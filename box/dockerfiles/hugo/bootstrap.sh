#!/bin/bash

local_repo_path=$HOME/github/younsl/younsl.github.io

select_option() {
    echo "Welcome to the Hugo Docker Script"
    printf "\n"
    echo "Select an option:"
    echo "1. Build Docker image"
    echo "2. Run Docker container"
    read option
}

build_docker_image() {
    printf "\n"
    echo "Building docker image ..."
    docker build -t hugo .
    echo "Docker image built successfully."
}

run_docker_container() {
    if [ -d "$local_repo_path" ]; then
        printf "\n"
        echo "Running Docker container ..."
        docker run -d \
            --name hugo \
            -p 1313:1313 \
            -v $local_repo_path:/app \
            hugo:latest

        printf "\n"
        echo "Docker container is now running."
        docker ps
    else
        echo "Local repository path not found: $local_repo_path"
    fi
}

process_option() {
    case $1 in
        1)
            build_docker_image
            ;;
        2)
            run_docker_container
            ;;
        *)
            echo "Invalid option"
            ;;
    esac
}

main() {
    select_option
    process_option $option
}

# Main script entry point
main
