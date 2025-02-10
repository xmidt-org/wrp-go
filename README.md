# wrp-go

wrp-go provides a library implementing the [Web Routing Protocol](https://github.com/xmidt-org/wrp-c/wiki/Web-Routing-Protocol) 
structures and supporting utilities.

[![Build Status](https://github.com/xmidt-org/wrp-go/actions/workflows/ci.yml/badge.svg)](https://github.com/xmidt-org/wrp-go/actions/workflows/ci.yml)
[![codecov](https://codecov.io/gh/xmidt-org/wrp-go/graph/badge.svg?token=tWY4sd44iI)](https://codecov.io/gh/xmidt-org/wrp-go)
[![Go Report Card](https://goreportcard.com/badge/github.com/xmidt-org/wrp-go)](https://goreportcard.com/report/github.com/xmidt-org/wrp-go)
[![Apache V2 License](http://img.shields.io/badge/license-Apache%20V2-blue.svg)](https://github.com/xmidt-org/wrp-go/blob/main/LICENSE)
[![GitHub Release](https://img.shields.io/github/release/xmidt-org/wrp-go.svg)](CHANGELOG.md)
[![GoDoc](https://pkg.go.dev/badge/github.com/xmidt-org/wrp-go/v3)](https://pkg.go.dev/github.com/xmidt-org/wrp-go/v3)

## Table of Contents

- [Code of Conduct](#code-of-conduct)
- [Validators](#validators)
- [Examples](#examples)
- [Contributing](#contributing)

## Code of Conduct

This project and everyone participating in it are governed by the [XMiDT Code Of Conduct](https://xmidt.io/code_of_conduct/). 
By participating, you agree to this Code.

## Validators

To setup application wrp validators, visit the example `ExampleMetaValidator` [GoDoc](https://pkg.go.dev/github.com/xmidt-org/wrp-go/v3/wrpvalidator#example-MetaValidator).

Application config example:
```yaml
# wrpValidators defines the wrp validators used to validate incoming wrp messages.
# (Optional)
# Available validator types: always_invalid, always_valid, utf8, msg_type, source, destination, simple_res_req, simple_event, spans
# Available validator levels: info, warning, error
# Validators can be disabled with `disable: true`, it is false by default
wrpValidators:
  - type: utf8
    level: warning
    disable: true
  - type: source
    level: error
```

## Examples 

To use the wrp-go library, it first should be added as an import in the file you plan to use it.
Examples can be found at the top of the [GoDoc](https://godoc.org/github.com/xmidt-org/wrp-go).

## Contributing

Refer to [CONTRIBUTING.md](CONTRIBUTING.md).
