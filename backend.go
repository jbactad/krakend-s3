package s3

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/luraproject/lura/v2/config"
	"github.com/luraproject/lura/v2/logging"
	"github.com/luraproject/lura/v2/proxy"
)

const Namespace = "github.com/jbactad/krakend-s3"

var (
	errNoConfig      = errors.New("aws s3: no extra config defined")
	errInvalidBucket = errors.New(`aws s3: invalid "bucket" defined`)
	errInvalidConfig = errors.New("aws s3: invalid config")
)

type ObjectGetter interface {
	GetObject(ctx context.Context, params *s3.GetObjectInput, optFns ...func(*s3.Options)) (*s3.GetObjectOutput, error)
}

type Options struct {
	AWSConfig     aws.Config
	Bucket        string
	PathExtension string
}

func BackendFactory(logger logging.Logger, bf proxy.BackendFactory) proxy.BackendFactory {
	return BackendFactoryWithClient(
		logger, bf, func(opts *Options) ObjectGetter {
			return s3.NewFromConfig(opts.AWSConfig)
		},
	)
}

func BackendFactoryWithClient(
	logger logging.Logger,
	bf proxy.BackendFactory,
	clientFactory func(opts *Options) ObjectGetter,
) proxy.BackendFactory {
	return func(remote *config.Backend) proxy.Proxy {
		logPrefix := "[BACKEND: " + remote.URLPattern + "][S3]"
		opts, err := getOptions(remote)
		if err != nil {
			if err != errNoConfig {
				logger.Error(logPrefix, err)
			}

			return bf(remote)
		}

		cl := clientFactory(opts)

		ef := proxy.NewEntityFormatter(remote)

		return func(ctx context.Context, request *proxy.Request) (*proxy.Response, error) {
			k := strings.TrimPrefix(request.Path, "/")

			if len(opts.PathExtension) > 0 {
				k += "." + opts.PathExtension
			}

			obj, err := cl.GetObject(
				ctx, &s3.GetObjectInput{
					Bucket: &opts.Bucket,
					Key:    &k,
				},
			)
			if err != nil {
				return nil, err
			}

			data := map[string]interface{}{}
			cont, err := io.ReadAll(obj.Body)
			if err != nil {
				return nil, err
			}

			if err := json.Unmarshal(cont, &data); err != nil {
				return nil, err
			}

			response := proxy.Response{
				Data:       data,
				IsComplete: true,
				Metadata: proxy.Metadata{
					Headers:    map[string][]string{},
					StatusCode: 200,
				},
			}

			response = ef.Format(response)

			return &response, nil
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

	bucket, ok := v.(string)
	if !ok {
		return nil, errInvalidBucket
	}

	if bucket == "" {
		return nil, errInvalidBucket
	}

	opts := &Options{
		Bucket:    bucket,
		AWSConfig: aws.Config{},
	}

	if region, ok := cfg["region"].(string); ok {
		opts.AWSConfig.Region = region
	}

	if endpoint, ok := cfg["endpoint"].(string); ok && endpoint != "" {
		opts.AWSConfig.EndpointResolverWithOptions = aws.EndpointResolverWithOptionsFunc(
			func(service, r string, options ...interface{}) (aws.Endpoint, error) {
				return aws.Endpoint{
					URL:               endpoint,
					SigningRegion:     r,
					HostnameImmutable: true,
				}, nil
			},
		)
	}

	if maxRetries, ok := cfg["max_retries"].(int); ok {
		opts.AWSConfig.RetryMaxAttempts = maxRetries
	}

	if pathExtension, ok := cfg["path_extension"].(string); ok {
		opts.PathExtension = strings.TrimPrefix(pathExtension, ".")
	}

	return opts, nil
}
