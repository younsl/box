# Cursor extension export and import

On the old machine:

```bash
cursor --list-extensions > extensions.list
```

On the new machine:

```bash
cursor extensions.list | xargs -L 1 cursor --install-extension
```
