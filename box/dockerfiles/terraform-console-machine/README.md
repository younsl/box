
# terraform-console-machine

The console container image to apply and plan Terraform locally.  
Container image for terraform administrator (e.g. DevOps Engineer or SRE).

&nbsp;

## How to use

Change volumes in docker-compose.yml file.

Run the container using docker compose.

```bash
# Build a docker images
$ docker build -t local-worker-worker .

# Run a docker container using docker compose
$ docker-compose up -d
```
