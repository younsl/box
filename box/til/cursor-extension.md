# Cursor extension export and import

On the old machine:

```bash
cursor --list-extensions > cursor-extensions.list
```

On the new machine:

```bash
cursor vscode-extensions.list | xargs -L 1 cursor --install-extension
```
