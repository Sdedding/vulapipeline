#!/bin/bash
set -e

export DISPLAY=:0

# Start an in-memory X server
Xvfb :0 -screen 0 1024x768x16 &
XVFB_PID=$!

# Launch the XFCE desktop
startxfce4 &
XFCE_PID=$!

# Serve the desktop over VNC
x11vnc -display :0 -nopw -forever -shared &
VNC_PID=$!

# Ensure dbus is running for the CLI tools
systemctl start dbus

# Expose via noVNC on all interfaces
websockify --web=/usr/share/novnc/ 0.0.0.0:6080 localhost:5900 &
NOVNC_PID=$!

trap 'kill $XVFB_PID $XFCE_PID $VNC_PID $NOVNC_PID' TERM INT
wait
