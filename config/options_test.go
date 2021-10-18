package config

import (
	"reflect"
	"testing"
)

func TestParseOptions(t *testing.T) {
	tests := []struct {
		name string
		want *Options
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ParseOptions(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ParseOptions() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSetEnv(t *testing.T) {
	type args struct {
		isDebug bool
	}
	tests := []struct {
		name string
		args args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
		})
	}
}

func Test_getCurrentAbPathByCaller(t *testing.T) {
	tests := []struct {
		name       string
		wantExPath string
		wantErr    bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotExPath, err := getCurrentAbPathByCaller()
			if (err != nil) != tt.wantErr {
				t.Errorf("getCurrentAbPathByCaller() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotExPath != tt.wantExPath {
				t.Errorf("getCurrentAbPathByCaller() gotExPath = %v, want %v", gotExPath, tt.wantExPath)
			}
		})
	}
}