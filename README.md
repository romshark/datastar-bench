# Datastar Benchmark

This is a collection of *synthetic* benchmarks supposed to test
[D*'s](https://data-star.dev) practical performance limits.

**Disclaimer:** Please, keep in mind, that if you're doing something like
that in production - you're most likely doing it wrong!

## Prerequisites

- Latest version of [Go](https://go.dev)

## Running

Run `go run .` and open `0.0.0.0:8080`.

Or compile to deployable binary using `go build`.

## Developing

Run `make dev` and open `0.0.0.0:7331`.

If you open `0.0.0.0:8080` you won't get automatic tab reloading.
