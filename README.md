# wrp-go

[![Build Status](https://github.com/xmidt-org/wrp-go/actions/workflows/ci.yml/badge.svg)](https://github.com/xmidt-org/wrp-go/actions/workflows/ci.yml)
[![codecov](https://codecov.io/gh/xmidt-org/wrp-go/branch/main/graph/badge.svg?token=tWY4sd44iI)](https://codecov.io/gh/xmidt-org/wrp-go)
[![Go Report Card](https://goreportcard.com/badge/github.com/xmidt-org/wrp-go)](https://goreportcard.com/report/github.com/xmidt-org/wrp-go)
[![Apache V2 License](http://img.shields.io/badge/license-Apache%20V2-blue.svg)](https://github.com/xmidt-org/wrp-go/blob/main/LICENSE)
[![GitHub Release](https://img.shields.io/github/release/xmidt-org/wrp-go.svg)](https://github.com/xmidt-org/wrp-go/releases)
[![GoDoc](https://pkg.go.dev/badge/github.com/xmidt-org/wrp-go/v5)](https://pkg.go.dev/github.com/xmidt-org/wrp-go/v5)

wrp-go provides a Go library implementing the [Web Routing Protocol](https://github.com/xmidt-org/wrp-c/wiki/Web-Routing-Protocol)
structures and supporting utilities.

## Features

- **WRP Message Types**: Complete implementation of all WRP message structures
- **Multiple Encodings**: Support for JSON and MessagePack serialization formats
- **Device Identifiers**: Parsing and validation of device IDs and locators with support for:
  - MAC addresses (with automatic normalization)
  - UUIDs
  - Serial numbers
  - DNS names
  - Event names
  - Self-referencing locators
- **Zero-Copy Parsing**: Efficient string handling for device identifiers and locators
- **Validation**: Built-in validation for messages and locators
- **Transcoding**: Easy conversion between different encoding formats

## Table of Contents

- [Code of Conduct](#code-of-conduct)
- [Examples](#examples)
- [Contributing](#contributing)

## Code of Conduct

This project and everyone participating in it are governed by the [XMiDT Code Of Conduct](https://xmidt.io/code_of_conduct/). 
By participating, you agree to this Code.

## Examples

To use the wrp-go library, it first should be added as an import in the file you plan to use it.
Examples can be found in the [GoDoc](https://pkg.go.dev/github.com/xmidt-org/wrp-go/v5).

## Contributing

Refer to [CONTRIBUTING.md](CONTRIBUTING.md).
