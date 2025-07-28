# Flatnotes

![image](https://github.com/user-attachments/assets/c8f91ab2-9678-47e6-a9ad-827bca7320ff)

Local note-taking application using [Flatnotes](https://github.com/dullage/flatnotes) with Docker volume persistence.

## Quick Start

```bash
make run     # Start the application (creates volume and prompts for credentials)
make start   # Start existing container
make stop    # Stop container
make clean   # Remove container and volume (deletes all data)
make logs    # View container logs
```

## Volume Management

```bash
make volume-create  # Create Docker volume
make volume-remove  # Remove Docker volume (WARNING: deletes all data)
```

### Inspect Volume

```bash
docker volume inspect flatnotes-data  # View volume details and local storage path
```

## Data Persistence

This setup uses Docker volumes to store data persistently on your local machine. The `flatnotes-data` volume is automatically created and mounted to `/data` inside the container, ensuring your notes persist across container restarts and updates.

Access the application at http://localhost:8080 after running `make run`.
