#!/bin/bash
set -e
export CC=clang
apt install -y --no-install-recommends flit clang
git clone https://codeberg.org/vula/highctidh/
cd highctidh;
git checkout v1.0.2024092800
CC=$CC make deb
ls -al *
ls -al dist/
dpkg -i `pwd`/dist/python3-highctidh*.deb
echo "OK: highctidh installed on $(hostname)"
