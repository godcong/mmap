// Copyright 2020 The go-mmap Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package examples_test

import (
	"fmt"
	"io"
	"log"
	"os"

	"github.com/godcong/mmap"
)

func ExampleOpen() {
	f, err := mmap.Open("example_mmap_test.go")
	if err != nil {
		log.Fatalf("could not mmap file: %+v", err)
	}
	defer f.Close()

	buf := make([]byte, 32)
	_, err = f.Read(buf)
	if err != nil {
		log.Fatalf("could not read into buffer: %+v", err)
	}

	fmt.Printf("%s\n", buf[:12])

	// Output:
	// // Copyright
}

func ExampleOpenFile_read() {
	f, err := mmap.OpenFile("example_mmap_test.go", os.O_RDONLY, 0o644)
	if err != nil {
		log.Fatalf("could not mmap file: %+v", err)
	}
	defer f.Close()

	buf := make([]byte, 32)
	_, err = f.ReadAt(buf, 0)
	if err != nil {
		log.Fatalf("could not read into buffer: %+v", err)
	}

	fmt.Printf("%s\n", buf[:12])

	// Output:
	// // Copyright
}

func ExampleOpenFile_readwrite() {
	tmp, err := os.CreateTemp("", "mmap-")
	if err != nil {
		log.Fatalf("could not create tmp file: %+v", err)
	}
	defer tmp.Close()
	defer os.Remove(tmp.Name())

	_, err = tmp.Write([]byte("hello world!"))
	if err != nil {
		log.Fatalf("could not write data: %+v", err)
	}

	err = tmp.Close()
	if err != nil {
		log.Fatalf("could not close file: %+v", err)
	}

	raw, err := os.ReadFile(tmp.Name())
	if err != nil {
		log.Fatalf("could not read back data: %+v", err)
	}

	fmt.Printf("%s\n", raw)

	rw, err := mmap.OpenFile(tmp.Name(), os.O_RDWR, 0o755)
	if err != nil {
		log.Fatalf("could not open mmap file: %+v", err)
	}
	defer rw.Close()

	_, err = rw.Write([]byte("bye!"))
	if err != nil {
		log.Fatalf("could not write to mmap file: %+v", err)
	}

	raw, err = os.ReadFile(tmp.Name())
	if err != nil {
		log.Fatalf("could not read back data: %+v", err)
	}

	fmt.Printf("%s\n", raw)

	// Output:
	// hello world!
	// bye!o world!
}

func ExampleOpenMem() {
	w, err := mmap.OpenMemS(mmap.MapMemKeyInvalid)
	if err != nil {
		log.Fatalf("could not create memory: %+v", err)
	}
	defer w.Close()

	n, err := w.Write([]byte("hello world!"))
	if err != nil {
		log.Fatalf("could not write to memory: %+v", err)
	}
	_, err = w.WriteAt([]byte("bye!"), 3)
	if err != nil {
		log.Fatalf("could not write at to memory: %+v", err)
	}
	r, err := mmap.OpenMem(w.ID(), n)
	if err != nil {
		log.Fatalf("could not open memory: %+v", err)
	}
	defer r.Close()
	rd, err := io.ReadAll(r)
	if err != nil {
		return
	}

	fmt.Printf("%s\n", rd)

	// Output:
	// helbye!orld!
}
