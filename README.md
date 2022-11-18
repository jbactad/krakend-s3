# krakend-s3

A Krakend plugin to serve S3 files directly from your gateway.

[![Go Reference](https://pkg.go.dev/badge/github.com/jbactad/krakend-s3.svg)](https://pkg.go.dev/github.com/jbactad/krakend-s3)
![Go](https://github.com/jbactad/krakend-s3/actions/workflows/go.yml/badge.svg)
[![codecov](https://codecov.io/gh/jbactad/krakend-s3/branch/main/graph/badge.svg?token=OEX805T5L8)](https://codecov.io/gh/jbactad/krakend-s3)
[![Go Report Card](https://goreportcard.com/badge/github.com/jbactad/krakend-s3)](https://goreportcard.com/report/github.com/jbactad/krakend-s3)

## Installation

```bash
go get -u github.com/jbactad/krakend-s3
```

## Quick Start

The s3 backend handler can be enabled in the backend layer of Krakend as shown below.

```go
package main

import (
	"context"
	"net/http"

	s3 "github.com/jbactad/krakend-s3"
	"github.com/luraproject/lura/v2/config"
	"github.com/luraproject/lura/v2/logging"
	"github.com/luraproject/lura/v2/proxy"

	"github.com/gin-gonic/gin"
	router "github.com/luraproject/lura/v2/router/gin"
	serverhttp "github.com/luraproject/lura/v2/transport/http/server"
	server "github.com/luraproject/lura/v2/transport/http/server/plugin"
)

func ExampleRegister() {
	cfg := config.ServiceConfig{
		Endpoints: []*config.EndpointConfig{
			{
				Backend: []*config.Backend{
					{
						URLPattern: "sample",
						ExtraConfig: map[string]interface{}{
							"github.com/jbactad/krakend-s3": map[string]interface{}{
								"bucket": "test-bucket",
							},
						},
					},
				},
			},
		},
	}
	logger := logging.NoOp

	backendFactory := proxy.HTTPProxyFactory(&http.Client{})

	// Wrap backendFactory with backend factory with s3 support.
	backendFactory = s3.BackendFactory(logger, backendFactory)

	pf := proxy.NewDefaultFactory(backendFactory, logger)

	handlerFactory := router.CustomErrorEndpointHandler(logger, serverhttp.DefaultToHTTPError)

	engine := gin.New()

	// setup the krakend router
	routerFactory := router.NewFactory(
		router.Config{
			Engine:         engine,
			ProxyFactory:   pf,
			Logger:         logger,
			RunServer:      router.RunServerFunc(server.New(logger, serverhttp.RunServer)),
			HandlerFactory: handlerFactory,
		},
	)

	// start the engines
	logger.Info("Starting the KrakenD instance")
	routerFactory.NewWithContext(context.Background()).Run(cfg)
}
```

Then in your gateway's config file make sure to add `github.com/jbactad/krakend-s3` in the
one of your endpoint backend's `extra_config` section.

```json
{
  "$schema": "https://www.krakend.io/schema/v3.json",
  "version": 3.0,
  "endpoints": [
    {
      "endpoint": "/sample",
      "backend": [
        {
          "url_pattern": "/sample-file-path",
          "extra_config": {
            "github_com/jbactad/krakend-s3": {
              "bucket": "test-bucket-name",
              "region": "eu-west-1",
              "max_retries": 5,
              "path_extension": "json"
            }
          }
        }
      ]
    }
  ]
}
```

## Configuring

The `url_pattern` of your `backend` configuration will be used as the key of the object to fetch from s3 bucket.

Because Krakend expects your backend to be rest apis,
refer from adding the file extension directly from the `url_pattern`.
Use the provided `path_extension` instead.

For example, the config below will use `sample-file-path.json` as the object key
and will return the object located at `s3://test-bucket-name/sample-file-path.json.`

```json
{
  "endpoint": "/sample",
  "backend": [
    {
      "url_pattern": "/sample-file-path",
      "extra_config": {
        "github_com/jbactad/krakend-s3": {
          "bucket": "test-bucket-name",
          "region": "eu-west-1",
          "max_retries": 5,
          "path_extension": "json"
        }
      }
    }
  ]
}
```

From krakend configuration file, these are the following options you can configure.

| Name           | Type | Required | Description                                                                      |
|----------------|------|:---------|----------------------------------------------------------------------------------|
| bucket         | int  | true     | The s3 bucket to fetch the object from.                                          |
| region         | int  | false    | The s3 region to use when fetching the object from the bucket.                   |
| endpoint       | int  | false    | The aws endpoint to use when fetching the object from s3.                        |
| max_retries    | int  | false    | Maximum number of retries to make if a failure occurred while fetching the file. |
| path_extension | int  | false    | Suffix to use when generating the key of the object. i.e. (json)                 |

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
