# `vula-go`

This is a **highly experimental** reimplementation of vula in [go](https://go.dev/). It is not yet reviewed for correctness of the vula protocol or security considerations beyond basic interoperability.

The goal of this reimplementation is to create a portable version of vula that can be used as a drop-in replacement of the python version. Currently, the main business logic of the commands `discover`, `publish` and `organize` is implemented. To navigate the codebase, start at `internal/cmd/`. For example, open `internal/cmd/organize.go` and inspect the `Run` function, which is the entry point of the command.

The commands `discover`, `publish` and `organize` are commands that start services which export dbus interfaces. When they are started, one can use the python client to interact with them. For an example of a regular command implementation, see `internal/cmd/status.go`.

## Tests

The current set of tests (incompletely) track the doc tests of the python version. To run all available tests, run `go test ./...`.

## Build the binary

If `go` is available:

```sh
go build -o vula-go cmd/vula
```

One can use the `helper.sh` script to build in a container instead, if `podman` or `docker` are available:

```sh
./helper.sh build_simple
```

These produce the binary `./vula-go`, run it with e.g. `./vula-go --help`.

## Patch a running vula installation

The easiest way to try the go version right now is to use the `helper.sh` script to patch an existing running vula installation.

```sh
./helper.sh # see usage

# patch (if `go` is available)
./helper.sh patch

# patch (build in a container instead if `podman` or `docker` is available)
./helper.sh patch_simple

# revert
./helper.sh unpatch
```

## Run podman tests with go version

Go to `vula/podman` and run for example:

```sh
make gotest
```
