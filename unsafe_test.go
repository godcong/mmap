package mmap

import (
	"reflect"
	"testing"
	"unsafe"
)

func TestPtrToBytes(t *testing.T) {
	data := []byte(`hello`)
	type args struct {
		ptr uintptr
		n   int
	}
	tests := []struct {
		name string
		args args
		want []byte
	}{
		{
			name: "",
			args: args{
				ptr: uintptr(unsafe.Pointer(&data[:1][0])),
				n:   len(data),
			},
			want: data,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := PtrToBytes(tt.args.ptr, tt.args.n); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("PtrToBytes() = %v, want %v", got, tt.want)
			}
		})
	}
}
