package trmon

import (
	"os"
	"reflect"
	"testing"
	"time"
)

func TestNewHost(t *testing.T) {
	type args struct {
		hostname  string
		community string
		logger    *Logger
	}
	tests := []struct {
		name    string
		args    args
		want    *Host
		wantErr bool
	}{
		{
			name: "can't connect snmp target",
			args: args{
				hostname:  "localhost",
				community: "",
				logger:    NewLogger(true, os.Stdout),
			},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewHost(tt.args.hostname, tt.args.community, tt.args.logger)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewHost() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewHost() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCounter_update(t *testing.T) {
	type fields struct {
		name       string
		Last       int64
		Before     int64
		LastTime   time.Time
		BeforeTime time.Time
		Diff       int64
		Rate       int64
		log        *Logger
	}
	type args struct {
		v int64
		t time.Time
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   fields
	}{
		// TODO: Add test cases.
		{
			name: "zero devide",
			fields: fields{
				name:       "hoge",
				Last:       0,
				Before:     0,
				LastTime:   time.Date(2000, time.December, 10, 10, 1, 1, 0, time.UTC),
				BeforeTime: time.Date(2000, time.December, 10, 10, 1, 0, 50, time.UTC),
				Diff:       0,
				Rate:       0,
				log:        NewLogger(true, os.Stdout),
			},
			args: args{
				v: 100,
				t: time.Date(2000, time.December, 10, 10, 1, 1, 30, time.UTC),
			},
			want: fields{
				name:       "hoge",
				Last:       100,
				Before:     0,
				LastTime:   time.Date(2000, time.December, 10, 10, 1, 1, 0, time.UTC),
				BeforeTime: time.Date(2000, time.December, 10, 10, 1, 1, 0, time.UTC),
				Diff:       100,
				Rate:       0,
			},
		},
		{
			name: "minus diff",
			fields: fields{
				name:       "hoge",
				Last:       70,
				Before:     10,
				LastTime:   time.Date(2000, time.December, 10, 10, 1, 6, 6, time.UTC),
				BeforeTime: time.Date(2000, time.December, 10, 10, 1, 1, 1, time.UTC),
				Diff:       40,
				Rate:       6,
				log:        NewLogger(true, os.Stdout),
			},
			args: args{
				v: 10,
				t: time.Date(2000, time.December, 10, 10, 1, 16, 6, time.UTC),
			},
			want: fields{
				name:       "hoge",
				Last:       10,
				Before:     70,
				LastTime:   time.Date(2000, time.December, 10, 10, 1, 16, 6, time.UTC),
				BeforeTime: time.Date(2000, time.December, 10, 10, 1, 6, 6, time.UTC),
				Diff:       0,
				Rate:       0,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Counter{
				name:       tt.fields.name,
				Last:       tt.fields.Last,
				Before:     tt.fields.Before,
				LastTime:   tt.fields.LastTime,
				BeforeTime: tt.fields.BeforeTime,
				Diff:       tt.fields.Diff,
				Rate:       tt.fields.Rate,
				log:        tt.fields.log,
			}
			c.update(tt.args.v, tt.args.t)
			if !reflect.DeepEqual(c.Rate, tt.want.Rate) {
				t.Errorf("c.update().Rate = %v, want %v", c.Rate, tt.want.Rate)
			}
		})
	}
}
