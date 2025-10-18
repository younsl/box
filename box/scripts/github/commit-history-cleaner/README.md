# commit-history-cleaner

This script is used to clean all commit history of a specific repository.

## Usage

Download cleaner script in the root directory of the repository.

```bash
wget -O commit-history-cleaner.sh https://raw.githubusercontent.com/younsl/o/main/box/scripts/github/commit-history-cleaner/commit-history-cleaner.sh
```

Run `sh` command to execute the script.

```bash
sh commit-history-cleaner.sh
```

> [!WARNING]
> This script will delete and recreate the 'main' branch to clean all commit history. Double check this execution before running. This action is irreversible.

Enter `y` key to continue.

```bash
This script will delete the 'main' branch and create the 'latest_branch' branch.
Do you want to continue? (yY/n) y
```