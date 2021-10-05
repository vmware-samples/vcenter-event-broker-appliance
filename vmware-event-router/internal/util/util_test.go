//go:build unit
// +build unit

package util

import (
	"testing"
)

func TestValidateAddress(t *testing.T) {
	type args struct {
		address string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "valid address",
			args: args{
				address: "0.0.0.0:8082",
			},
			wantErr: false,
		},
		{
			name: "no ip address specified",
			args: args{
				address: ":8082",
			},
			wantErr: true,
		},
		{
			name: "no port specified",
			args: args{
				address: "0.0.0.0",
			},
			wantErr: true,
		},
		{
			name: "empty string",
			args: args{
				address: "",
			},
			wantErr: true,
		},
		{
			name: "invalid string",
			args: args{
				address: ":",
			},
			wantErr: true,
		},
		{
			name: "URI provided",
			args: args{
				address: "https://0.0.0.0.8082",
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := ValidateAddress(tt.args.address); (err != nil) != tt.wantErr {
				t.Errorf("validateAddress() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
