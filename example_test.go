package s3_test

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
