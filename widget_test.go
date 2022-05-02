package trmon

import (
	"regexp"
	"testing"
)

func TestMainWidget_setUnit(t *testing.T) {
	type fields struct {
		Name          string
		Hosts         []*Host
		displayDownIF bool
		unit          Unit
		unitCalc      UnitCalc
		NarrowWidget  *NarrowWidget
	}
	type args struct {
		unit Unit
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
		{
			name:   "Unspecified value",
			fields: fields{},
			args: args{
				unit: 143131,
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &MainWidget{
				Name:          tt.fields.Name,
				Hosts:         tt.fields.Hosts,
				displayDownIF: tt.fields.displayDownIF,
				unit:          tt.fields.unit,
				unitCalc:      tt.fields.unitCalc,
				NarrowWidget:  tt.fields.NarrowWidget,
			}
			if err := m.setUnit(tt.args.unit); (err != nil) != tt.wantErr {
				t.Errorf("MainWidget.setUnit() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestNarrowWidget_setRegexp(t *testing.T) {
	type fields struct {
		Name   string
		regexp *regexp.Regexp
	}
	type args struct {
		expr string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
		{
			name:   "Empty String",
			fields: fields{},
			args: args{
				expr: "",
			},
			wantErr: false,
		},
		{
			name:   "Invalid RegulerExpr",
			fields: fields{},
			args: args{
				expr: "*",
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			n := &NarrowWidget{
				Name:   tt.fields.Name,
				regexp: tt.fields.regexp,
			}
			if err := n.setRegexp(tt.args.expr); (err != nil) != tt.wantErr {
				t.Errorf("NarrowWidget.setRegexp() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
