package uncompress

import (
	"fmt"
	"io"
	"os"
	"testing"
)

func TestUncompress(t *testing.T) {

	tests := []struct {
		file                string
		isBzip              bool
		isXZ                bool
		uncompressedByteLen int
	}{
		{
			file:                "testdata/golang.mbox",
			isBzip:              false,
			isXZ:                false,
			uncompressedByteLen: 26097,
		},
		{
			file:                "testdata/golang.mbox.bz2",
			isBzip:              true,
			isXZ:                false,
			uncompressedByteLen: 26097,
		},
		{
			file:                "testdata/golang.mbox.xz",
			isBzip:              false,
			isXZ:                true,
			uncompressedByteLen: 26097,
		},
	}
	for i, tt := range tests {
		t.Run(fmt.Sprintf("test_%d", i), func(t *testing.T) {
			f, err := os.Open(tt.file)
			if err != nil {
				t.Fatal(err)
			}
			defer f.Close()

			u, err := newUncompress(f)
			if err != nil {
				t.Fatalf("file %s error %s", tt.file, err)
			}

			fmt.Printf("%#v\n", u)
			if got, want := u.isBzip(), tt.isBzip; got != want {
				t.Errorf("isBzip got %t want %t", got, want)
			}
			if got, want := u.isXZ(), tt.isXZ; got != want {
				t.Errorf("isXZ got %t want %t", got, want)
			}
		})
	}

	for i, tt := range tests {
		t.Run(fmt.Sprintf("testb_%d", i), func(t *testing.T) {
			f, err := os.Open(tt.file)
			if err != nil {
				t.Fatal(err)
			}
			defer f.Close()

			u, err := newUncompress(f)
			if err != nil {
				t.Fatal(err)
			}

			var r io.Reader
			r = io.Reader(f)
			if u.isBzip() {
				r = u.bzipWrappedReader(f)
			}
			if u.isXZ() {
				r, err = u.xzWrappedReader(f)
				if err != nil {
					t.Fatal(err)
				}
			}
			b, err := io.ReadAll(r)
			if err != nil {
				t.Fatal(err)
			}
			if got, want := len(b), tt.uncompressedByteLen; got != want {
				t.Errorf("byte len got %d want %d", got, want)
			}
		})
	}

	for i, tt := range tests {
		t.Run(fmt.Sprintf("testc_%d", i), func(t *testing.T) {
			f, err := os.Open(tt.file)
			if err != nil {
				t.Fatal(err)
			}
			defer f.Close()

			r, err := NewReader(f)
			if err != nil {
				t.Fatal(err)
			}

			b, err := io.ReadAll(r)
			if err != nil {
				t.Fatal(err)
			}
			if got, want := len(b), tt.uncompressedByteLen; got != want {
				t.Errorf("byte len got %d want %d", got, want)
			}
		})
	}
}
