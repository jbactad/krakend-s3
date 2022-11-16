package s3

import (
	"context"
	"errors"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/luraproject/lura/v2/config"
	"github.com/luraproject/lura/v2/logging"
	"github.com/luraproject/lura/v2/proxy"
)

var (
	errNoConfig      = errors.New("aws s3: no extra config defined")
	errInvalidBucket = errors.New(`aws s3: invalid "bucket" defined`)
	errInvalidConfig = errors.New("aws s3: invalid config")
)

const Namespace = "github.com/jbactad/krakend-s3"

type Options struct {
	AWS *aws.Config
}

func BackendFactory(logger logging.Logger, bf proxy.BackendFactory) proxy.BackendFactory {
	return func(remote *config.Backend) proxy.Proxy {
		logPrefix := "[BACKEND: " + remote.URLPattern + "][S3]"
		_, err := getOptions(remote)
		if err != nil {
			if err != errNoConfig {
				logger.Error(logPrefix, err)
			}

			return bf(remote)
		}

		if _, ok := remote.ExtraConfig[Namespace]; !ok {
			logger.Error(logPrefix, errNoConfig)
			return bf(remote)
		}

		return func(ctx context.Context, request *proxy.Request) (*proxy.Response, error) {
			return nil, nil
		}
	}
}

func getOptions(remote *config.Backend) (*Options, error) {
	v, ok := remote.ExtraConfig[Namespace]
	if !ok {
		return nil, errNoConfig
	}

	cfg, ok := v.(map[string]interface{})
	if !ok {
		return nil, errInvalidConfig
	}

	v, ok = cfg["bucket"]
	if !ok {
		return nil, errInvalidBucket
	}

	key, ok := v.(string)
	if !ok {
		return nil, errInvalidBucket
	}

	if key == "" {
		return nil, errInvalidBucket
	}

	opts := &Options{

	}

	return opts, nil
}
