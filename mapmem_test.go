package mmap_test

import (
	"bytes"
	"io"
	"os"
	"reflect"
	"testing"
	"unsafe"

	"github.com/godcong/mmap"
	"github.com/godcong/mmap/unsafex"
)

func TestMemFileWrite(t *testing.T) {
	log := mmap.Log().WithGroup("mapmem")
	display := func(id int, sz int) []byte {
		t.Helper()
		s, err := mmap.OpenMem(id, sz)
		if err != nil {
			log.Error("OpenMem", "err", err)
			return nil
		}
		defer s.Close()
		raw, err := io.ReadAll(s)
		if err != nil {
			t.Fatalf("could not read file %d: %+v", id, err)
		}
		return raw
	}

	for _, tc := range []struct {
		name  string
		flags int
	}{
		// {
		// 	name:  "write-only",
		// 	flags: Write,
		// },
		{
			name:  "read-write",
			flags: os.O_RDWR,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			w, err := mmap.OpenMem(mmap.MapMemKeyInvalid, len([]byte("hello world!\nbye.\n")))
			if err != nil {
				t.Fatalf("could not open file: %+v", err)
			}
			defer w.Close()
			var written int
			written, err = w.Write([]byte("hello world!\nbye.\n"))
			if err != nil {
				t.Fatalf("could not write: %+v", err)
			}
			log.Info("data write", "written", written, "id", w.ID())
			if got, want := display(w.ID(), len("hello world!\nbye.\n")), []byte("hello world!\nbye.\n"); !bytes.Equal(got, want) {
				t.Fatalf("invalid content:\ngot= %q\nwant=%q\n", got, want)
			}

			written, err = w.WriteAt([]byte("bye!\n"), 3)
			if err != nil {
				t.Fatalf("could not write-at: %+v", err)
			}

			log.Info("data write", "written", written, "id", w.ID())
			if got, want := display(w.ID(), len("hello world!\nbye.\n")), []byte("helbye!\nrld!\nbye.\n"); !bytes.Equal(got, want) {
				t.Fatalf("invalid content:\ngot= %q\nwant=%q\n", got, want)
			}

			var seeked int64
			seeked, err = w.Seek(0, io.SeekStart)
			if err != nil {
				t.Fatalf("could not seek to start: %+v", err)
			}
			log.Info("data seek", "seeked", seeked, "id", w.ID())

			written, err = w.Write([]byte("hello world!\nbye\n"))
			if err != nil {
				t.Fatalf("could not write: %+v", err)
			}
			log.Info("data write", "written", written, "id", w.ID())
			if got, want := display(w.ID(), len("hello world!\nbye.\n")), []byte("hello world!\nbye\n\n"); !bytes.Equal(got, want) {
				t.Fatalf("invalid content:\ngot= %q\nwant=%q\n", got, want)
			}

			seeked, err = w.Seek(5, io.SeekEnd)
			if err != nil {
				t.Fatalf("could not seek from end: %+v", err)
			}
			log.Info("data seek", "seeked", seeked, "id", w.ID())
			err = w.WriteByte('t')
			if err != nil {
				t.Fatalf("could not write-byte: %+v", err)
			}
			log.Info("data write byte")
			if got, want := display(w.ID(), len("hello world!\nbye.\n")), []byte("hello world!\ntye\n\n"); !bytes.Equal(got, want) {
				t.Fatalf("invalid content:\ngot= %q\nwant=%q\n", got, want)
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
			if got := unsafex.PointToBytes(tt.args.ptr, tt.args.n); !reflect.DeepEqual(got, tt.want) {
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
			if got := unsafex.BytesToPoint(tt.args.data); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("BytesToPoint() = %v, want %v", got, tt.want)
			}
		})
	}
}
