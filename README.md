# MMAP

[![GitHub release](https://img.shields.io/github/release/godcong/mmap.svg)](https://github.com/godcong/mmap/releases)
[![go.dev reference](https://img.shields.io/badge/go.dev-reference-007d9c?logo=go&logoColor=white&style=flat)](https://pkg.go.dev/github.com/godcong/mmap)
[![codecov](https://codecov.io/gh/godcong/mmap/branch/main/graph/badge.svg)](https://codecov.io/gh/godcong/mmap)
[![GoDoc](https://godoc.org/github.com/godcong/mmap?status.svg)](http://godoc.org/github.com/godcong/mmap)
[![License](https://img.shields.io/github/license/godcong/mmap.svg)](https://github.com/godcong/mmap/blob/main/LICENSE)
[![Go Report Card](https://goreportcard.com/badge/github.com/godcong/mmap)](https://goreportcard.com/report/github.com/godcong/mmap)

The `MMAP` package is a syscall interface to provide safe and efficient access to memory.

Supports for Darwin,Linux and Windows architectures.

## Installation

```
> go get github.com/godcong/mmap@latest
```

## Example

See [`examples`](https://github.com/godcong/mmap/blob/main/examples) folder

## Similar Packages

- github.com/godcong/mmap
- golang.org/x/exp/mmap
- github.com/riobard/godcong
- launchpad.net/gommap
- github.com/edsrzf/mmap-go

## Plan

- [ ] TCP transmits data through shared memory between threads.
- [ ] Turn off shared memory more elegantly.

## Memory Map Service(TODO) Flow
See [`ServiceMemMap`](docs/ServiceMemMap.mmd)

## License

This Project used MIT.