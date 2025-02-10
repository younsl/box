#!/bin/bash

# Print help message
print_usage() {
    echo "Usage: $0 [--dry-run]"
    echo "  --dry-run    Show commands without generating PDF"
}

# Parse arguments
DRY_RUN=0
while [[ "$#" -gt 0 ]]; do
    case $1 in
        --dry-run) DRY_RUN=1 ;;
        -h|--help) print_usage; exit 0 ;;
        *) echo "Unknown option: $1"; print_usage; exit 1 ;;
    esac
    shift
done

# Collect files to convert (sh-compatible version)
files_to_convert=()
find . -type f -name "slide-deck.md" > /tmp/slides_list.tmp
while IFS= read -r slide_path; do
    files_to_convert+=("$slide_path")
done < /tmp/slides_list.tmp
rm /tmp/slides_list.tmp

# Check if no files found
if [ ${#files_to_convert[@]} -eq 0 ]; then
    echo "No slide-deck.md files found to convert."
    exit 1
fi

# Show files to be converted
echo "The following files will be converted to PDF:"
for file in "${files_to_convert[@]}"; do
    echo "- $file"
done

# Get user confirmation
read -p "Proceed with conversion? (y/N) " response
if [[ ! "$response" =~ ^[yY]$ ]]; then
    echo "PDF conversion cancelled."
    exit 0
fi

# Process each file
for slide_path in "${files_to_convert[@]}"; do
    dir_path=$(dirname "$slide_path")
    
    if [ $DRY_RUN -eq 1 ]; then
        echo "[DRY-RUN] Following command will be executed:"
        echo "cd $dir_path && podman run --rm --init -v $PWD:/home/marp/app/ -e LANG=$LANG marpteam/marp-cli slide-deck.md --pdf --allow-local-files"
    else
        echo "Converting: $slide_path"
        cd "$dir_path" && \
        podman run --rm --init \
            -v $PWD:/home/marp/app/ \
            -e LANG=$LANG \
            marpteam/marp-cli \
            slide-deck.md --pdf --allow-local-files
        cd - > /dev/null
    fi
done