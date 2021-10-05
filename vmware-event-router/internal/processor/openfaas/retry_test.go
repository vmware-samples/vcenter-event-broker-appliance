//go:build unit
// +build unit

package openfaas

import (
	"context"
	"errors"
	"testing"
)

func Test_isSuccessful(t *testing.T) {
	type args struct {
		status int
		err    error
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "500 without error",
			args: args{status: 500, err: nil},
			want: false,
		},
		{
			name: "200 without error",
			args: args{status: 200, err: nil},
			want: true,
		},
		{
			name: "200 with error",
			args: args{status: 200, err: errors.New("failed")},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isSuccessful(tt.args.status, tt.args.err); got != tt.want {
				t.Errorf("isSuccessful() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_isRetryable(t *testing.T) {
	ctx := context.Background()
	cancelCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	type args struct {
		ctx  context.Context
		code int
		err  error
	}
	tests := []struct {
		name      string
		args      args
		cancelCtx bool // simulate ctx cancelled err
		want      bool
		wantErr   bool
	}{
		{
			name:      "429 retry",
			args:      args{ctx: ctx, code: 429, err: nil},
			cancelCtx: false,
			want:      true,
			wantErr:   false,
		},
		{
			name:      "503 retry",
			args:      args{ctx: ctx, code: 503, err: nil},
			cancelCtx: false,
			want:      true,
			wantErr:   false,
		},
		{
			name:      "200 no retry",
			args:      args{ctx: ctx, code: 200, err: nil},
			cancelCtx: false,
			want:      false,
			wantErr:   false,
		},
		{
			name:      "200 with ctx cancelled no retry",
			args:      args{ctx: cancelCtx, code: 200, err: nil},
			cancelCtx: true,
			want:      false,
			wantErr:   true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.cancelCtx == true {
				cancel()
			}
			got, err := isRetryable(tt.args.ctx, tt.args.code, tt.args.err)
			if (err != nil) != tt.wantErr {
				t.Errorf("isRetryable() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("isRetryable() got = %v, want %v", got, tt.want)
			}
		})
	}
}
