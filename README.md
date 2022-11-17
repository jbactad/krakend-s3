# krakend-s3

A Krakend plugin to serve S3 files directly from your gateway.

[![Go Reference](https://pkg.go.dev/badge/github.com/jbactad/krakend-s3.svg)](https://pkg.go.dev/github.com/jbactad/krakend-s3)
![Go](https://github.com/jbactad/krakend-s3/actions/workflows/go.yml/badge.svg)
[![codecov](https://codecov.io/gh/jbactad/krakend-s3/branch/main/graph/badge.svg?token=OEX805T5L8)](https://codecov.io/gh/jbactad/krakend-s3)
[![Go Report Card](https://goreportcard.com/badge/github.com/jbactad/krakend-s3)](https://goreportcard.com/report/github.com/jbactad/krakend-s3)

## Installation

```bash
go get -u github.com/jbactad/krakend-newrelic-v2
```

## Quick Start

[//]: # (TODO: finish quick start guide)

## Configuring

From krakend configuration file, these are the following options you can configure.

| Name           | Type | Description                                                                      |
|----------------|------|----------------------------------------------------------------------------------|
| bucket         | int  | The s3 bucket to fetch the file from.                                            |
| region         | int  | The s3 region to use when fetching the file from the bucket.                     |
| max_retries    | int  | Maximum number of retries to make if a failure occurred while fetching the file. |
| path_extension | int  | Suffix to use when generating the file path. i.e. (json)                         |

## Development

### Requirements

To start development, make sure you have the following dependencies installed in your development environment.

- golang >=v1.17

### Setup

Run the following to install the necessary tools to run tests.

```bash
make setup
```

### Generate mocks for unit tests

```bash
make gen
```

### Running unit tests

To run the unit tests, execute the following command.

```bash
make test-unit
```
