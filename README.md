# krakend-s3
A Krakend plugin to serve S3 files directly from your gateway.

[//]: # (TODO: add bagdes for ci, coverage, godoc, go report)


## Installation

```bash
go get -u github.com/jbactad/krakend-newrelic-v2
```

## Quick Start

[//]: # (TODO: finish quick start guide)

## Configuring

NewRelic related configurations are all read from environment variables.
Refer to [newrelic go agent](https://pkg.go.dev/github.com/newrelic/go-agent/v3/newrelic@v3.18.1#ConfigFromEnvironment)
package to know more of which NewRelic options can be configure.

From krakend configuration file, these are the following options you can configure.

| Name | Type | Description                                           |
|------|------|-------------------------------------------------------|
| rate | int  | The rate the middlewares instrument your application. |


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
