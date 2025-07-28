#!/bin/bash

#==================================================
# [Last Modified Date]
# 2023-08-01
#
# [Author]
# Younsung Lee (cysl@kakao.com)
#
# [Description]
# Change all gp2 volumes to gp3 in a specific region
# Prerequisite: You have to install `jq` and `awscli`, FIRST.
#==================================================

region='ap-northeast-2'
aws_cmd_path=$(command -v aws)

# Function to check if a command exists
function command_exists() {
    command -v "$1" >/dev/null 2>&1
}

# Function to check if jq is installed, and if not, exit the script
function check_jq_installed() {
    if ! command_exists "jq"; then
        echo "[e] 'jq' command not found. Please install 'jq' before running this script."
        exit 1
    fi
}

# Function to check if AWS CLI is installed, and if not, exit the script
function check_awscli_installed() {
    if ! command_exists "aws"; then
        echo "[e] 'aws' command not found. Please install 'awscli' before running this script."
        exit 1
    fi
}

# Function to find all gp2 volumes within the given region
function find_gp2_volumes() {
    echo "[i] Start finding all gp2 volumes in ${region}"
    volume_ids=$(
        ${aws_cmd_path} ec2 describe-volumes \
        --region "${region}" \
        --filters Name=volume-type,Values=gp2 | \
        jq -r '.Volumes[].VolumeId'
    )

    echo "[i] List up all gp2 volumes in ${region}"
    echo "========================================="
    echo "$volume_ids"
    echo "========================================="
}

# Function to confirm the action before migration
function confirm_migration() {
    while true; do
        read -p "Do you want to proceed with the migration? (y/n): " choice
        case "$choice" in
            [yY])
                echo "[i] Starting volume migration..."
                return 0
                ;;
            [nN])
                echo "[i] Migration canceled by the user."
                exit 0
                ;;
            *)
                echo "[e] Invalid choice. Please enter 'y' or 'n'."
                ;;
        esac
    done
}

# Function to migrate a single gp2 volume to gp3
function migrate_volume_to_gp3() {
    local volume_id="$1"
    result=$(${aws_cmd_path} ec2 modify-volume \
        --region "${region}" \
        --volume-type gp3 \
        --volume-id "${volume_id}" | \
        jq -r '.VolumeModification.ModificationState'
    )

    if [ $? -eq 0 ] && [ "$result" == "modifying" ]; then
        echo "[i] Volume $volume_id changed to state 'modifying' successfully."
    else
        echo "[e] ERROR: Couldn't change volume $volume_id type to gp3!"
        exit 1
    fi
}

# Main function to run the entire process
function main() {
    check_jq_installed
    check_awscli_installed

    find_gp2_volumes
    confirm_migration

    echo "[i] Migrating all gp2 volumes to gp3"
    for volume_id in $volume_ids; do
        migrate_volume_to_gp3 "$volume_id"
    done

    echo "[i] All gp2 volumes have been migrated to gp3 successfully!"
}

# Call the main function to start the script
main
