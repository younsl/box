# EKS Security Enhancement

A presentation about improving EKS Security configuration.

## Usage

Save the current Marp presentation as a `.pdf` file.

> [!NOTE]
> If your presentation references local files, add the `--allow-local-files` option to enable file references.

```bash
# Convert slide deck into PDF
docker run --rm --init -v $PWD:/home/marp/app/ -e LANG=$LANG marpteam/marp-cli slide-deck.md --pdf --allow-local-files
```