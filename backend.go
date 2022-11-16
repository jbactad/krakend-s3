package krakends3

import (
	"context"

	"github.com/luraproject/lura/v2/config"
	"github.com/luraproject/lura/v2/logging"
	"github.com/luraproject/lura/v2/proxy"
)

func BackendFactory(logger logging.Logger, bf proxy.BackendFactory) proxy.BackendFactory {
	return func(remote *config.Backend) proxy.Proxy {
		if _, ok := remote.ExtraConfig["github.com/jbactad/krakend-s3"]; !ok {
			return bf(remote)
		}
		return func(ctx context.Context, request *proxy.Request) (*proxy.Response, error) {
			return nil, nil
		}
	}
}
