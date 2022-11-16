package s3_test

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"strings"
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
			name: "s3 client returned a valid object, should parse and return content",
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
			want: &proxy.Response{
				Data: map[string]interface{}{
					"property1": "value1",
				},
				IsComplete: true,
				Metadata: proxy.Metadata{
					Headers:    map[string][]string{},
					StatusCode: 200,
				},
			},
			setup: func(logger *mocks.MockLogger, client *mocks.MockObjectGetter) {
				b := "bucket1"
				k := "sample.json"
				client.EXPECT().
					GetObject(
						ctx, gomock.Eq(
							&awsS3.GetObjectInput{
								Bucket: &b,
								Key:    &k,
							},
						),
					).
					Times(1).
					Return(
						&awsS3.GetObjectOutput{
							Body: io.NopCloser(
								strings.NewReader(
									`{
	"property1": "value1"
}`,
								),
							),
						}, nil,
					)
			},
		},
		{
			name: "s3 client returned an error, should return error",
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
			wantErr: func(t assert.TestingT, err error, i ...interface{}) bool {
				return assert.EqualValues(t, errors.New("something went wrong"), err)
			},
			want: nil,
			setup: func(logger *mocks.MockLogger, client *mocks.MockObjectGetter) {
				client.EXPECT().
					GetObject(gomock.Any(), gomock.Any()).
					Times(1).
					Return(
						nil, errors.New("something went wrong"),
					)
			},
		},
		{
			name: "error encountered while reading s3 response, should return error",
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
			wantErr: func(t assert.TestingT, err error, i ...interface{}) bool {
				return assert.EqualValues(t, errors.New("something went wrong"), err)
			},
			want: nil,
			setup: func(logger *mocks.MockLogger, client *mocks.MockObjectGetter) {
				client.EXPECT().
					GetObject(gomock.Any(), gomock.Any()).
					Times(1).
					Return(
						&awsS3.GetObjectOutput{
							Body: io.NopCloser(&faultyReader{error: errors.New("something went wrong")}),
						}, nil,
					)
			},
		},
		{
			name: "s3 returned invalid json, should return error",
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
			wantErr: func(t assert.TestingT, err error, i ...interface{}) bool {
				return assert.IsType(t, &json.SyntaxError{}, err)
			},
			want: nil,
			setup: func(logger *mocks.MockLogger, client *mocks.MockObjectGetter) {
				client.EXPECT().
					GetObject(gomock.Any(), gomock.Any()).
					Times(1).
					Return(
						&awsS3.GetObjectOutput{
							Body: io.NopCloser(strings.NewReader("")),
						}, nil,
					)
			},
		},
	}
	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				if tt.setup != nil {
					tt.setup(l, cl)
				}

				bf := proxy.BackendFactory(
					func(remote *config.Backend) proxy.Proxy {
						return func(ctx context.Context, request *proxy.Request) (*proxy.Response, error) {
							t.Error("this backend factory should not been called")
							return nil, nil
						}
					},
				)

				b := s3.BackendFactoryWithClient(
					l, bf,
					func(opts *s3.Options) s3.ObjectGetter {
						return cl
					},
				)
				p := b(tt.args.config)
				got, err := p(ctx, nil)
				if !tt.wantErr(t, err) {
					return
				}

				assert.EqualValues(t, tt.want, got)
			},
		)
	}
}

func TestBackendFactory_invalidConfig(t *testing.T) {
	ctrl := gomock.NewController(t)
	l := mocks.NewMockLogger(ctrl)

	expectedResp := &proxy.Response{
		Data:       map[string]interface{}{},
		IsComplete: true,
		Metadata:   proxy.Metadata{},
		Io:         nil,
	}

	type args struct {
		config *config.Backend
	}
	tests := []struct {
		name  string
		args  args
		setup func(logger *mocks.MockLogger)
	}{
		{
			name: "namespace not found in config, should return original proxy",
			args: args{
				config: &config.Backend{
					URLPattern:  "/some-endpoint",
					ExtraConfig: map[string]interface{}{},
				},
			},
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
							return expectedResp, nil
						}
					},
				)
				p := b(tt.args.config)
				got, _ := p(nil, nil)

				assert.Equal(t, expectedResp, got)
			},
		)
	}
}

type faultyReader struct {
	error error
}

func (f faultyReader) Read(_ []byte) (n int, err error) {
	return 0, f.error
}
