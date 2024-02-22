package mmap

import (
	"reflect"
	"testing"
	"unsafe"
)

func TestCreateMem(t *testing.T) {
	mem, err := OpenMemFile(0, PROT_READ|PROT_WRITE, 1024)
	if err != nil {
		t.Fatal("create mem failed:", err)
	}
	type args struct {
		id   int
		prot int
		size int
	}
	tests := []struct {
		name    string
		args    args
		want    *MemFile
		wantErr bool
	}{
		{
			name: "",
			args: args{
				id:   0,
				prot: 0,
				size: 0,
			},
			want:    mem,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mem.Write([]byte("hello"))
		})
	}
}

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

func TestPointToBytes(t *testing.T) {
	data := []byte(`hello`)
	type args struct {
		ptr *byte
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
				ptr: (*byte)(unsafe.Pointer(&data[:1][0])),
				n:   len(data),
			},
			want: data,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := PointToBytes(tt.args.ptr, tt.args.n); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("PointToBytes() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBytesToPoint(t *testing.T) {
	data := []byte(`hello`)
	type args struct {
		data []byte
	}

	tests := []struct {
		name string
		args args
		want *byte
	}{
		{
			name: "",
			args: args{
				data: data,
			},
			want: (*byte)(unsafe.Pointer(&data[:1][0])),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := BytesToPoint(tt.args.data); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("BytesToPoint() = %v, want %v", got, tt.want)
			}
		})
	}
}
