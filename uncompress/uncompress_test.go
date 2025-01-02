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
		typer               string
		uncompressedByteLen int
	}{
		{
			file:                "testdata/golang.mbox",
			typer:               "unknown",
			uncompressedByteLen: 26097,
		},
		{
			file:                "testdata/golang.mbox.bz2",
			typer:               "bzip2",
			uncompressedByteLen: 26097,
		},
		{
			file:                "testdata/golang.mbox.xz",
			typer:               "xz",
			uncompressedByteLen: 26097,
		},
		{
			file:                "testdata/golang.mbox.gz",
			typer:               "gzip",
			uncompressedByteLen: 26097,
		},
	}

	// check types
	for i, tt := range tests {
		t.Run(fmt.Sprintf("test_type_%d", i), func(t *testing.T) {
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
			if got, want := u.IsType(tt.typer), true; got != want {
				t.Errorf("isBzip got %t want %t (for %s)", got, want, u.MIME)
			}
		})
	}

	// check uncompress
	for i, tt := range tests {
		t.Run(fmt.Sprintf("test_uncompress_%d", i), func(t *testing.T) {
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
