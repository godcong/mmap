//go:build linux || darwin || freebsd

package mmap

import (
	"os"
	"testing"
)

func Test_getPageSize(t *testing.T) {
	type args struct {
		size int
	}
	tests := []struct {
		name string
		args args
		want int
	}{
		{
			name: "",
			args: args{
				size: 0,
			},
			want: os.Getpagesize(),
		},
		{
			name: "",
			args: args{
				size: 4,
			},
			want: 4,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getPageSize(tt.args.size); got != tt.want {
				t.Errorf("getPageSize() = %v, want %v", got, tt.want)
			}
		})
	}
}
