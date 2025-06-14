#!/bin/bash
set -e
apt update
apt install -y --no-install-recommends \
    python3 python3-pip python3-venv \
    xfce4 x11vnc xvfb novnc websockify xterm dbus-x11
