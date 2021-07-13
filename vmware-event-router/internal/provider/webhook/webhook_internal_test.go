// +build unit

package webhook

import (
	"testing"
)

func Test_validatePath(t *testing.T) {
	tests := []struct {
		name    string
		path    string
		want    string
		wantErr bool
	}{
		{name: "get default path", path: "", want: defaultPath, wantErr: false},
		{name: "invalid root path", path: "/", want: "", wantErr: true},
		{name: "remove slashes and make all lower", path: "/somePath/", want: "somepath", wantErr: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := validatePath(tt.path)
			if (err != nil) != tt.wantErr {
				t.Errorf("validatePath() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("validatePath() got = %v, want %v", got, tt.want)
			}
		})
	}
}
