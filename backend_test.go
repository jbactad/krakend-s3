package s3_test

import (
	"context"
	"errors"
	"testing"

	awsS3 "github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/golang/mock/gomock"
	s3 "github.com/jbactad/krakend-s3"
	"github.com/jbactad/krakend-s3/mocks"
	"github.com/luraproject/lura/v2/config"
	"github.com/luraproject/lura/v2/proxy"
	"github.com/stretchr/testify/assert"
)

//go:generate mockgen -destination mocks/lura_logger.go -package mocks github.com/luraproject/lura/v2/logging Logger
//go:generate mockgen -destination mocks/object_getter.go -package mocks -source backend.go

func TestBackendFactoryWithClient_backendProxyInvoked(t *testing.T) {
	ctrl := gomock.NewController(t)
	l := mocks.NewMockLogger(ctrl)
	cl := mocks.NewMockObjectGetter(ctrl)
	ctx := context.Background()

	type args struct {
		config *config.Backend
	}
	tests := []struct {
		name    string
		args    args
		wantErr assert.ErrorAssertionFunc
		want    *proxy.Response
		setup   func(logger *mocks.MockLogger, client *mocks.MockObjectGetter)
	}{
		{
			name: "valid config, should call s3 client to fetch file",
			args: args{
				config: &config.Backend{
					URLPattern: "/sample.json",
					ExtraConfig: map[string]interface{}{
						s3.Namespace: map[string]interface{}{
							"bucket": "bucket1",
						},
					},
				},
			},
			wantErr: assert.NoError,
			want:    nil,
			setup: func(logger *mocks.MockLogger, client *mocks.MockObjectGetter) {
				b := "bucket1"
				k := "sample.json"
				client.EXPECT().GetObject(
					ctx,
					gomock.Eq(
						&awsS3.GetObjectInput{
							Bucket: &b,
							Key:    &k,
						},
					),
				).Times(1)
			},
		},
	}
	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				if tt.setup != nil {
					tt.setup(l, cl)
				}

				b := s3.BackendFactoryWithClient(
					l, func(remote *config.Backend) proxy.Proxy {
						return proxy.NoopProxy
					},
					func(opts *s3.Options) s3.ObjectGetter {
						return cl
					},
				)
				p := b(tt.args.config)
				got, err := p(ctx, nil)
				if !tt.wantErr(t, err) {
					return
				}

				assert.Equal(t, tt.want, got)
			},
		)
	}
}

func TestBackendFactory_invalidConfig(t *testing.T) {
	ctrl := gomock.NewController(t)
	l := mocks.NewMockLogger(ctrl)

	oResp := &proxy.Response{
		Data:       map[string]interface{}{},
		IsComplete: true,
		Metadata:   proxy.Metadata{},
		Io:         nil,
	}

	type args struct {
		config *config.Backend
	}
	tests := []struct {
		name    string
		args    args
		wantErr assert.ErrorAssertionFunc
		want    *proxy.Response
		setup   func(logger *mocks.MockLogger)
	}{
		{
			name: "namespace not found in config, should return original proxy",
			args: args{
				config: &config.Backend{
					URLPattern:  "/some-endpoint",
					ExtraConfig: map[string]interface{}{},
				},
			},
			wantErr: assert.NoError,
			want:    oResp,
		},
		{
			name: "invalid config for namespace, should log error and return original proxy",
			args: args{
				config: &config.Backend{
					URLPattern: "/some-endpoint",
					ExtraConfig: map[string]interface{}{
						s3.Namespace: "",
					},
				},
			},
			wantErr: assert.NoError,
			want:    oResp,
			setup: func(logger *mocks.MockLogger) {
				logger.EXPECT().Error(
					"[BACKEND: /some-endpoint][S3]",
					errors.New("aws s3: invalid config"),
				)
			},
		},
		{
			name: "s3 bucket not defined, should log error and return original proxy",
			args: args{
				config: &config.Backend{
					URLPattern: "/some-endpoint",
					ExtraConfig: map[string]interface{}{
						s3.Namespace: map[string]interface{}{},
					},
				},
			},
			wantErr: assert.NoError,
			want:    oResp,
			setup: func(logger *mocks.MockLogger) {
				logger.EXPECT().Error(
					"[BACKEND: /some-endpoint][S3]",
					errors.New(`aws s3: invalid "bucket" defined`),
				)
			},
		},
		{
			name: "s3 bucket empty, should log error and return original proxy",
			args: args{
				config: &config.Backend{
					URLPattern: "/some-endpoint",
					ExtraConfig: map[string]interface{}{
						s3.Namespace: map[string]interface{}{
							"bucket": "",
						},
					},
				},
			},
			wantErr: assert.NoError,
			want:    oResp,
			setup: func(logger *mocks.MockLogger) {
				logger.EXPECT().Error(
					"[BACKEND: /some-endpoint][S3]",
					errors.New(`aws s3: invalid "bucket" defined`),
				)
			},
		},
		{
			name: "s3 bucket type invalid, should log error and return original proxy",
			args: args{
				config: &config.Backend{
					URLPattern: "/some-endpoint",
					ExtraConfig: map[string]interface{}{
						s3.Namespace: map[string]interface{}{
							"bucket": 1,
						},
					},
				},
			},
			wantErr: assert.NoError,
			want:    oResp,
			setup: func(logger *mocks.MockLogger) {
				logger.EXPECT().Error(
					"[BACKEND: /some-endpoint][S3]",
					errors.New(`aws s3: invalid "bucket" defined`),
				)
			},
		},
		{
			name: "s3 bucket empty, should log error and return original proxy",
			args: args{
				config: &config.Backend{
					URLPattern: "/some-endpoint",
					ExtraConfig: map[string]interface{}{
						s3.Namespace: map[string]interface{}{
							"key":    "",
							"secret": "some-secret",
						},
					},
				},
			},
			wantErr: assert.NoError,
			want:    oResp,
			setup: func(logger *mocks.MockLogger) {
				logger.EXPECT().Error(
					"[BACKEND: /some-endpoint][S3]",
					errors.New(`aws s3: invalid "bucket" defined`),
				)
			},
		},
	}
	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				if tt.setup != nil {
					tt.setup(l)
				}

				b := s3.BackendFactory(
					l, func(remote *config.Backend) proxy.Proxy {
						return func(ctx context.Context, request *proxy.Request) (*proxy.Response, error) {
							return oResp, nil
						}
					},
				)
				p := b(tt.args.config)
				got, err := p(nil, nil)
				if !tt.wantErr(t, err) {
					return
				}

				assert.Equal(t, tt.want, got)
			},
		)
	}
}
