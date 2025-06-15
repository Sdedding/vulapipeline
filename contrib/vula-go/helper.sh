#!/usr/bin/env sh

# should be POSIX compliant

installed() {
  command -v "$1" >/dev/null 2>&1
}

help() {
  cat <<EOF
usage: $(basename "$0") [help|build_simple|patch|patch_simple|unpatch]

  help          print this message
  build_simple  build the binary with podman or docker
  patch         patch vula systemd service files, reload daemon and services
  patch_simple  same as patch but builds the binary in a container
  unpatch       unpatch vula systemd service files, reload daemon and services
  test          run extensive tests

Note: this script can "patch" an existing vula installation.

EOF
}

build_simple() {
  base_dir=$(CDPATH="" cd -- "$(dirname -- "$0")" && pwd -P)
  if "$base_dir/vula-go" --help >/dev/null 2>&1; then
    return
  fi
  if installed "podman"; then
    podman run --rm -v "$base_dir:/root/vula-go" -w /root/vula-go golang:1.24.3 go build -o "/root/vula-go/vula-go" "/root/vula-go/cmd/vula"
  elif intalled "docker"; then
    docker run --rm -v "$base_dir:/root/vula-go" -w /root/vula-go golang:1.24.3 go build -o "/root/vula-go/vula-go" "/root/vula-go/cmd/vula"
  else
    echo "Command podman or docker needs to be installed and available..." >&2
    exit 1
  fi
}

patch() {
  base_dir=$(CDPATH="" cd -- "$(dirname -- "$0")" && pwd -P)
  if ! installed "go"; then
    echo "Command go needs to be installed and available..." >&2
    echo "If you want to build in a container, try patch_simple instead." >&2
    exit 1
  fi
  go build -o "/usr/local/bin/vula-go" "$base_dir/cmd/vula"
  chown root:root /usr/local/bin/vula-go
  chmod 755 /usr/local/bin/vula-go

  systemctl stop vula-organize vula-discover vula-publish

  for service in organize discover publish; do
    sed -i "s|^ExecStart=.*$|ExecStart=vula-go -v $service|" "/etc/systemd/system/vula-$service.service"
  done

  systemctl daemon-reload
  systemctl start vula-organize vula-discover vula-publish
}

patch_simple() {
  base_dir=$(CDPATH="" cd -- "$(dirname -- "$0")" && pwd -P)
  if ! "$base_dir/vula-go" --help >/dev/null 2>&1; then
    build_simple
  fi
  cp "$base_dir/vula-go" /usr/local/bin/vula-go
  chown root:root /usr/local/bin/vula-go
  chmod 755 /usr/local/bin/vula-go

  systemctl stop vula-organize vula-discover vula-publish

  for service in organize discover publish; do
    sed -i "s|^ExecStart=.*$|ExecStart=vula-go -v $service|" "/etc/systemd/system/vula-$service.service"
  done

  systemctl daemon-reload
  systemctl start vula-organize vula-discover vula-publish
}

unpatch() {
  systemctl stop vula-discover vula-publish vula-organize

  rm /usr/local/bin/vula-go
  for service in organize discover publish; do
    sed -i "s|^ExecStart=.*$|ExecStart=vula $service|" "/etc/systemd/system/vula-$service.service"
  done

  systemctl daemon-reload
  systemctl start vula-organize vula-discover vula-publish
}

test() {
  base_dir=$(CDPATH="" cd -- "$(dirname -- "$0")" && pwd -P)
  go test ./... && go test -race ./... &&
    CC=clang go test -msan ./... && go test -asan ./... &&
    go run honnef.co/go/tools/cmd/staticcheck@latest ./... &&
    go run github.com/kisielk/errcheck@latest ./... &&
    go run github.com/securego/gosec/v2/cmd/gosec@latest ./... &&
    go run golang.org/x/vuln/cmd/govulncheck@latest ./...
}

# main logic
if [ "$#" -lt 1 ] || [ "$#" -gt 1 ]; then
  echo "Invalid number of arguments provided. Requires exactly 1 argument..." >&2
  help
  exit 1
fi

commands="dirname cat"
for command in $commands; do
  if ! installed "$command"; then
    echo "Command $command needs to be installed and available..." >&2
    exit 1
  fi
done

case "$1" in
"build_simple")
  build_simple
  ;;
"patch")
  patch
  ;;
"patch_simple")
  patch_simple
  ;;
"unpatch")
  unpatch
  ;;
"test")
  test
  ;;
*)
  help
  ;;
esac

