# MageAI 0.9.73 Custom Docker Image

This is a custom Docker image based on [MageAI 0.9.73](https://github.com/mage-ai/mage-ai/releases/tag/0.9.73) with additional Python packages for enhanced functionality.

## Additional Packages

This custom image includes the following additional packages not found in the original MageAI image:

- **sendgrid==6.12.4** - Email delivery service integration
- **gspread-formatting==1.2.1** - Google Sheets formatting capabilities

## Base Image Features

Built on the official MageAI 0.9.73 image with:

- Python 3.10 (bookworm)
- Spark integration with sparkmagic
- R environment with package management
- Microsoft SQL Server ODBC drivers
- NFS support
- Graphviz for visualizations
- All standard MageAI integrations

## Usage

```bash
# Build the image
docker build -t mageai:0.9.73-custom.1 .

# Run the container
docker run --name mageai -p 6789:6789 -p 7789:7789 mageai:0.9.73-custom.1
```

## Ports

- `6789` - Main MageAI web interface
- `7789` - Additional MageAI service port

## Environment Variables

- `MAGE_DATA_DIR` - Data directory path (default: `/home/src/mage_data`)
- `PYTHONPATH` - Python path includes `/home/src`

## Package Verification

To verify that the custom packages are properly installed in the container, run the following command:

```bash
# Check if sendgrid and gspread-formatting packages are installed
docker run --rm mageai:0.9.73-custom.1 pip list | egrep 'sendgrid|gspread-formatting'
```

Expected output:

```bash
gspread               5.7.2
gspread-formatting    1.2.1
sendgrid              6.12.4
```

Alternative verification inside a running container:

```bash
# Enter the container
docker exec -it <container_name> /bin/bash

# Check installed packages
pip list | egrep 'sendgrid|gspread-formatting'
```

## Custom Package Use Cases

### SendGrid Integration

The included `sendgrid` package enables email notifications and delivery features within MageAI pipelines.

### Google Sheets Formatting

The `gspread-formatting` package allows advanced formatting options when working with Google Sheets data sources.
