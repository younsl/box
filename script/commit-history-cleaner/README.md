
# scripts

## Overview

Delete and Recreate Branch Script to clean all commit history.

This script is designed to delete the main branch and create a new branch called latest_branch. The script ensures that any errors encountered during execution will immediately terminate the process.

> **Important**  
> Please be aware that branch deletion is irreversible, and there is no built-in recovery mechanism in this script. Make sure you have a backup or are certain of your actions before proceeding with branch deletion.

## Prerequisites

Git should be installed and properly configured on your system.

## Usage

Run the script by executing the following command in the terminal:

```bash
./commit_history_cleaner.sh
```

or

```bash
sh commit_history_cleaner.sh
```

The script will prompt you with a confirmation message:

```bash
This script will delete the main branch and create the latest_branch branch.
Do you want to continue? (y/n)
```

Type `y` or `n` to proceed or abort, respectively, and press Enter.

If you choose to continue (`y`), the script will perform the following actions:

- Checkout to a new orphan branch named latest_branch.
- Add all files to the new branch.
- Commit the changes with the message "Initial commit".
- Delete the main branch (if it exists).
- Rename the current branch to main.
Force push the main branch to the remote repository.
- Display a completion message.

If you choose to abort (`n`), the script will display an "Aborted" message and exit.

## Notes

- The script assumes that the remote repository is already set up and configured with a remote named origin.

- Use this script with caution, as it permanently deletes the main branch and replaces it with the latest_branch.

Feel free to modify this script according to your specific requirements or repository setup.
