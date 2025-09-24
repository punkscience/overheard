#!/bin/bash

# This script installs the 'overheard' application as a systemd service.

# --- Configuration ---
APP_NAME="overheard"
APP_BINARY_PATH="/usr/local/bin/$APP_NAME"
SERVICE_NAME="$APP_NAME.service"
SERVICE_FILE_PATH="/etc/systemd/system/$SERVICE_NAME"
USER=${SUDO_USER:-$(whoami)}

# --- Check for root privileges ---
if [ "$EUID" -ne 0 ]; then
  echo "Please run with sudo"
  exit
fi

# --- Build the application ---
echo "Building the application..."
go build -o $APP_NAME /home/darryl/projects/overheard/cmd/overheard

# --- Copy the binary to the system directory ---
echo "Installing the application binary to $APP_BINARY_PATH..."
cp $APP_NAME $APP_BINARY_PATH

# --- Create the systemd service file ---
echo "Creating the systemd service file at $SERVICE_FILE_PATH..."
cat > $SERVICE_FILE_PATH << EOL
[Unit]
Description=Overheard - A command-line utility for scheduling and recording audio from internet radio streams.
After=network.target

[Service]
Type=simple
User=$USER
ExecStart=$APP_BINARY_PATH record
Restart=on-failure

[Install]
WantedBy=multi-user.target
EOL

# --- Reload the systemd daemon, enable and start the service ---
echo "Reloading systemd, enabling and starting the service..."
systemctl daemon-reload
systemctl enable $SERVICE_NAME
systemctl start $SERVICE_NAME

echo "Installation complete. The '$APP_NAME' service is now running."
echo "You can check the status of the service by running: systemctl status $APP_NAME"
