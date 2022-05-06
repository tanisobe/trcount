package trmon

import (
	"os"
	"testing"

	"github.com/jroimartin/gocui"
)

func TestApp_initHosts(t *testing.T) {
	type fields struct {
		hosts []*Host
		gui   *gocui.Gui
		log   *Logger
	}
	type args struct {
		hostnames []string
		community string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
		{
			name: "no hosts",
			fields: fields{
				hosts: make([]*Host, 0),
				gui:   &gocui.Gui{},
				log:   NewLogger(true, os.Stdout),
			},
			args: args{
				hostnames: []string{},
				community: "my_comm",
			},
			wantErr: true,
		},
		{
			name: "invalid community",
			fields: fields{
				hosts: make([]*Host, 0),
				gui:   &gocui.Gui{},
				log:   NewLogger(true, os.Stdout),
			},
			args: args{
				hostnames: []string{"127.0.0.1"},
				community: "mogear",
			},
			wantErr: true,
		},
		{
			name: "unreachable host",
			fields: fields{
				hosts: make([]*Host, 0),
				gui:   &gocui.Gui{},
				log:   NewLogger(true, os.Stdout),
			},
			args: args{
				hostnames: []string{"127.0.0.254"},
				community: "my_comm",
			},
			wantErr: true,
		},
		{
			name: "valid hosts",
			fields: fields{
				hosts: make([]*Host, 0),
				gui:   &gocui.Gui{},
				log:   NewLogger(true, os.Stdout),
			},
			args: args{
				hostnames: []string{"127.0.0.1", "127.0.0.1"},
				community: "my_comm",
			},
			wantErr: false,
		},
		{
			name: "valid and invalid hosts",
			fields: fields{
				hosts: make([]*Host, 0),
				gui:   &gocui.Gui{},
				log:   NewLogger(true, os.Stdout),
			},
			args: args{
				hostnames: []string{"127.0.0.1", "invalid-host"},
				community: "my_comm",
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := &App{
				hosts: tt.fields.hosts,
				gui:   tt.fields.gui,
				log:   tt.fields.log,
			}
			if err := a.initHosts(tt.args.hostnames, tt.args.community); (err != nil) != tt.wantErr {
				t.Errorf("App.initHosts() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
