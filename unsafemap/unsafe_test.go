package unsafemap

import (
	"reflect"
	"testing"
)

func TestPtrToBytes(t *testing.T) {
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
				ptr: BytesToPoint(data),
				n:   len(data),
			},
			want: data,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := PointToBytes(tt.args.ptr, tt.args.n); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("PtrToBytes() = %v, want %v", got, tt.want)
			}
		})
	}
}
