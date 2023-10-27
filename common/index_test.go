package common

import (
	"fmt"
	"testing"
)

func TestErr(t *testing.T) {
	var nilErr error = nil

	type args struct {
		rets []any
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
		{"T1", args{rets: []any{1, 2, 3}}, false},
		{"T2", args{rets: []any{1, "", nilErr}}, false},
		{"T3", args{rets: []any{1, "", fmt.Errorf("test error")}}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := Err(tt.args.rets...); (err != nil) != tt.wantErr {
				t.Errorf("Err() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
