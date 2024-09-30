# UK Controller Plugin SRD Tools

Tools relating to the Standard Routes Document and AIRAC Cycles for the UK Controller Plugin.

## Using

The CLI for the tool is documented in-app. You can access this using the `--help` flag to list all commands, or with a specific command to get specific information about that command:

```shell
./srd-tools-linux-amd64 --help
```

An `.env` file must be provided for commands that require database access (import and download). An example file is present in this repo.

## Building

This project is built in `Golang`. If you've got `asdf` installed, you can install the correct version by simply running `asdf install`.

We currently build for Linux on both AMD64 and ARM64, as well as MacOS on Apple Silicon.

## Testing

### Automated

All automated testing is performed using `go test`.

You will need `docker` installed, as we use `testcontainers` in order to spin up local testing containers for the database.

### Manually

If you wish to run the full process locally, you will need a MySQL database (or docker container thereof) running on your local machine.

## Benchmarking

There are a number of benchmarks that may be run (they do not run as part of the test suite by default) to check the performance of the import.

You can run these using a command like the following: `/go test -benchmem -run=^$ -tags integration -bench ^BenchmarkImport$ github.com/VATSIM-UK/ukcp-srd-tools/internal/srd`

You will need a local copy of the SRD for these tests to run.

### Benchmarking Results

Below is the output of running the benchmarks for a simple export of the 2409 SRD cycle:

```shell
❯ go test -benchmem -run=^$ -tags integration -bench ^BenchmarkImport$ github.com/VATSIM-UK/ukcp-srd-tools/internal/srd
goos: darwin
goarch: arm64
pkg: github.com/VATSIM-UK/ukcp-srd-tools/internal/srd
cpu: Apple M1 Max
BenchmarkImport-10             1        23460575042 ns/op       5394721968 B/op  4413233 allocs/op
--- BENCH: BenchmarkImport-10
    import_test.go:631: Heap allocated for import: 32629360
```

## Releasing

Creating a release on GitHub will automatically build binaries and attach them to the release.
