package krakends3_test

import (
	"context"
	"testing"

	krakends3 "github.com/jbactad/krakend-s3"
	"github.com/luraproject/lura/v2/config"
	"github.com/luraproject/lura/v2/logging"
	"github.com/luraproject/lura/v2/proxy"
	"github.com/stretchr/testify/assert"
)

func TestBackendFactory_invalidConfig(t *testing.T) {
	l := logging.NoOp
	oResp := &proxy.Response{
		Data:       map[string]interface{}{},
		IsComplete: true,
		Metadata:   proxy.Metadata{},
		Io:         nil,
	}
	bf := proxy.BackendFactory(
		func(remote *config.Backend) proxy.Proxy {
			return func(ctx context.Context, request *proxy.Request) (*proxy.Response, error) {
				return oResp, nil
			}
		},
	)

	type fields struct {
		logger logging.Logger
	}
	type args struct {
		config *config.Backend
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr assert.ErrorAssertionFunc
		want    *proxy.Response
	}{
		{
			name: "namespace not found in config, should invoke original proxy",
			fields: fields{
				logger: l,
			},
			args: args{
				config: &config.Backend{
					ExtraConfig: map[string]interface{}{},
				},
			},
			wantErr: assert.NoError,
			want:    oResp,
		},
	}
	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				b := krakends3.BackendFactory(tt.fields.logger, bf)
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
