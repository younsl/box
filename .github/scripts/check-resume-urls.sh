#!/bin/bash

# Extract and check URLs from resume.html
# Only checks http/https URLs, skips mailto links

set -e

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# HTTP Status Code Categories
SUCCESS_CODES=(200 201 301 302 303 307 308)
WARNING_CODES=(400 401 403 405 406 429)
FAILURE_CODES=(404 408 410 500 502 503 504)

# Check if code is in array
is_in_array() {
    local code=$1
    shift
    local arr=("$@")
    for c in "${arr[@]}"; do
        [[ "$c" == "$code" ]] && return 0
    done
    return 1
}

# Extract URLs from resume.html
# Check if running from repo root or resume directory
if [ -f "box/resume/resume.html" ]; then
    RESUME_FILE="box/resume/resume.html"
elif [ -f "resume.html" ]; then
    RESUME_FILE="resume.html"
else
    echo "Error: resume.html not found!"
    exit 1
fi

URLS=$(grep -oE 'href="https?://[^"]+' "$RESUME_FILE" | cut -d'"' -f2 | sort -u)

if [ -z "$URLS" ]; then
    echo "No URLs found in resume.html"
    exit 0
fi

echo "Checking URLs in resume.html ..."

FAILED=0
TOTAL=0

for URL in $URLS; do
    TOTAL=$((TOTAL + 1))
    printf "Checking: %-70s " "$URL"
    
    # Use curl with timeout and follow redirects
    HTTP_CODE=$(curl -o /dev/null -s -w "%{http_code}" -L --connect-timeout 5 --max-time 10 "$URL" 2>/dev/null || echo "000")
    
    # Special handling for connection errors
    if [ "$HTTP_CODE" = "000" ]; then
        echo -e "${RED}✗${NC} [TIMEOUT/ERROR]"
        FAILED=$((FAILED + 1))
    # Special handling for known false positives
    elif [[ "$URL" == *"fonts.googleapis.com"* ]] || [[ "$URL" == *"fonts.gstatic.com"* ]]; then
        echo -e "${GREEN}✓${NC} [PRECONNECT]"
    # Check against status code arrays
    elif is_in_array "$HTTP_CODE" "${SUCCESS_CODES[@]}"; then
        echo -e "${GREEN}✓${NC} [$HTTP_CODE]"
    elif is_in_array "$HTTP_CODE" "${WARNING_CODES[@]}"; then
        echo -e "${YELLOW}⚠${NC} [$HTTP_CODE]"
    elif is_in_array "$HTTP_CODE" "${FAILURE_CODES[@]}"; then
        echo -e "${RED}✗${NC} [$HTTP_CODE]"
        FAILED=$((FAILED + 1))
    else
        # Unknown status code - treat as warning
        echo -e "${YELLOW}⚠${NC} [$HTTP_CODE]"
    fi
done

echo "Ready URLs: ${TOTAL}, Dead URLS: ${FAILED}"

if [ $FAILED -gt 0 ]; then
    echo -e "${RED}URL check failed! Please fix broken links before committing.${NC}"
    exit 1
else
    echo -e "${GREEN}All URLs are accessible!${NC}"
    exit 0
fi
