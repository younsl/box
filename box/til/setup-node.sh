# 1. Package Update
sudo apt install
sudo apt upgrade -y

# 2. Install docker
curl -fsSL https://get.docker.com -o get-docker.sh
sudo sh get-docker.sh
sudo docker ps && sudo docker version

# 3. Prometheus Node Exporter
## Download binary file
cd /tmp
curl -L https://github.com/prometheus/node_exporter/releases/download/v1.9.1/node_exporter-1.9.1.linux-amd64.tar.gz | tar xzf -

## Install
tar xzf node_exporter-*.tar.gz
sudo cp node_exporter-*/node_exporter /usr/local/bin/
sudo chown root:root /usr/local/bin/node_exporter
sudo chmod +x /usr/local/bin/node_exporter

## Create user for node_exporter
sudo useradd --no-create-home --shell /bin/false node_exporter

## Create systemd service file
sudo tee /etc/systemd/system/node_exporter.service > /dev/null <<EOF
[Unit]
Description=Node Exporter v1.9.1
Documentation=https://prometheus.io/docs/guides/node-exporter/
Wants=network-online.target
After=network-online.target

[Service]
User=node_exporter
Group=node_exporter
Type=simple
Restart=on-failure
RestartSec=5s
ExecStart=/usr/local/bin/node_exporter
SyslogIdentifier=node_exporter

[Install]
WantedBy=multi-user.target
EOF

## Start systemd
sudo systemctl daemon-reload
sudo systemctl enable node_exporter
sudo systemctl start node_exporter
