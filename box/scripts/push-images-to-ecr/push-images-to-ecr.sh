#!/bin/bash

# AWS Region
AWS_REGION="ap-northeast-2"

# Get AWS Account ID dynamically
print_info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Colors for output (moved up for early use)
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

get_aws_account_id() {
    local ACCOUNT_ID=$(aws sts get-caller-identity --query 'Account' --output text 2>/dev/null)
    if [ $? -ne 0 ] || [ -z "$ACCOUNT_ID" ]; then
        print_error "Failed to get AWS account ID. Please check your AWS credentials."
        exit 1
    fi
    echo "$ACCOUNT_ID"
}

# ECR Registry (dynamically constructed)
ACCOUNT_ID=$(get_aws_account_id)
ECR_REGISTRY="${ACCOUNT_ID}.dkr.ecr.${AWS_REGION}.amazonaws.com"

# Source images
MYSQL_SOURCE="mysql:8.0.39"
VALKEY_SOURCE="valkey/valkey:8.1.0-alpine"

# Target repositories
MYSQL_TARGET="${ECR_REGISTRY}/mysql"
VALKEY_TARGET="${ECR_REGISTRY}/valkey"

# Function to print colored messages (print_warning only, others moved up)

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

# Check if required tools are installed
check_requirements() {
    print_info "Checking requirements..."
    
    if ! command -v podman &> /dev/null; then
        print_error "Podman is not installed"
        exit 1
    fi
    
    if ! command -v aws &> /dev/null; then
        print_error "AWS CLI is not installed"
        exit 1
    fi
    
    print_info "All requirements satisfied"
}

# Login to ECR
ecr_login() {
    print_info "Logging in to ECR..."
    aws ecr get-login-password --region ${AWS_REGION} | podman login --username AWS --password-stdin ${ECR_REGISTRY}
    
    if [ $? -ne 0 ]; then
        print_error "Failed to login to ECR"
        exit 1
    fi
    
    print_info "Successfully logged in to ECR"
}

# Function to push multi-arch image
push_multiarch_image() {
    local SOURCE_IMAGE=$1
    local TARGET_IMAGE=$2
    local IMAGE_NAME=$3
    
    print_info "Processing ${IMAGE_NAME}..."
    
    # Pull and tag images for both architectures
    print_info "Pulling ${SOURCE_IMAGE} for linux/amd64..."
    podman pull --platform linux/amd64 ${SOURCE_IMAGE}
    
    if [ $? -ne 0 ]; then
        print_error "Failed to pull ${SOURCE_IMAGE} for linux/amd64"
        return 1
    fi
    
    # Tag AMD64 immediately after pulling
    print_info "Tagging AMD64 image..."
    podman tag ${SOURCE_IMAGE} ${TARGET_IMAGE}-amd64
    
    print_info "Pulling ${SOURCE_IMAGE} for linux/arm64..."
    podman pull --platform linux/arm64 ${SOURCE_IMAGE}
    
    if [ $? -ne 0 ]; then
        print_error "Failed to pull ${SOURCE_IMAGE} for linux/arm64"
        return 1
    fi
    
    # Tag ARM64 immediately after pulling
    print_info "Tagging ARM64 image..."
    podman tag ${SOURCE_IMAGE} ${TARGET_IMAGE}-arm64
    
    # Push both architecture-specific images
    print_info "Pushing AMD64 image..."
    podman push ${TARGET_IMAGE}-amd64
    
    if [ $? -ne 0 ]; then
        print_error "Failed to push ${TARGET_IMAGE}-amd64"
        return 1
    fi
    
    print_info "Pushing ARM64 image..."
    podman push ${TARGET_IMAGE}-arm64
    
    if [ $? -ne 0 ]; then
        print_error "Failed to push ${TARGET_IMAGE}-arm64"
        return 1
    fi
    
    # Create and push manifest
    print_info "Creating manifest for ${IMAGE_NAME}..."
    
    # Remove existing manifest if it exists
    podman manifest rm ${TARGET_IMAGE} 2>/dev/null || true
    
    # Create manifest
    podman manifest create ${TARGET_IMAGE} \
        ${TARGET_IMAGE}-amd64 \
        ${TARGET_IMAGE}-arm64
    
    if [ $? -ne 0 ]; then
        print_error "Failed to create manifest for ${TARGET_IMAGE}"
        return 1
    fi
    
    # Annotate manifest with architecture information
    podman manifest annotate ${TARGET_IMAGE} ${TARGET_IMAGE}-amd64 --arch amd64 --os linux
    podman manifest annotate ${TARGET_IMAGE} ${TARGET_IMAGE}-arm64 --arch arm64 --os linux
    
    # Push manifest
    podman manifest push ${TARGET_IMAGE}
    
    if [ $? -ne 0 ]; then
        print_error "Failed to push manifest for ${TARGET_IMAGE}"
        return 1
    fi
    
    print_info "Successfully pushed ${IMAGE_NAME} with multi-arch support"
    
    # Clean up local images to save space
    print_info "Cleaning up local images for ${IMAGE_NAME}..."
    podman rmi ${TARGET_IMAGE}-amd64 2>/dev/null || true
    podman rmi ${TARGET_IMAGE}-arm64 2>/dev/null || true
    
    return 0
}

# Main execution
main() {
    print_info "Starting multi-architecture image push to ECR"
    
    # Check requirements
    check_requirements
    
    # Login to ECR
    ecr_login
    
    # Podman supports manifest operations natively, no need for experimental flag
    
    # Process MySQL
    print_info "=== Processing MySQL ==="
    push_multiarch_image "${MYSQL_SOURCE}" "${MYSQL_TARGET}:8.0.39" "MySQL"
    
    if [ $? -ne 0 ]; then
        print_error "Failed to process MySQL"
        exit 1
    fi
    
    # Process Valkey
    print_info "=== Processing Valkey ==="
    push_multiarch_image "${VALKEY_SOURCE}" "${VALKEY_TARGET}:8.1.0-alpine" "Valkey"
    
    if [ $? -ne 0 ]; then
        print_error "Failed to process Valkey"
        exit 1
    fi
    
    print_info "=== Summary ==="
    print_info "Successfully pushed the following images with multi-arch support:"
    print_info "  - ${MYSQL_TARGET}:8.0.39 (linux/amd64, linux/arm64)"
    print_info "  - ${VALKEY_TARGET}:8.1.0-alpine (linux/amd64, linux/arm64)"
    
    print_info "Script completed successfully!"
}

# Run main function
main
