package trmon

import (
	"os"
	"testing"
)

func TestRun(t *testing.T) {
	type args struct {
		community string
		interval  int
		expr      string
		isDebug   bool
		args      []string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
		{
			name: "no hosts",
			args: args{
				community: "my_conn",
				interval:  5,
				expr:      "",
				isDebug:   true,
				args:      []string{},
			},
			wantErr: true,
		},
		{
			name: "invalid community hosts",
			args: args{
				community: "invalid",
				interval:  5,
				expr:      "",
				isDebug:   true,
				args:      []string{"127.0.0.1"},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := Run(tt.args.community, tt.args.interval, tt.args.expr, tt.args.isDebug, os.Stderr, tt.args.args); (err != nil) != tt.wantErr {
				t.Errorf("Run() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}
