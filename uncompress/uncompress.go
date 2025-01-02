package uncompress

import (
	"compress/bzip2"
	"compress/gzip"
	"io"
	"os"

	"github.com/h2non/filetype"
	"github.com/ulikunitz/xz"
)

// uncompress is a struct representing the filename, extension and MIME
// of a file
type uncompress struct {
	Extension string
	MIME      string
}

// newUncompress determines the characteristics of a file
func newUncompress(f *os.File) (*uncompress, error) {
	head := make([]byte, 261)
	f.Read(head)

	kind, err := filetype.Match(head)
	if err != nil {
		return nil, err
	}
	u := &uncompress{
		Extension: kind.Extension,
		MIME:      kind.MIME.Value,
	}
	_, err = f.Seek(0, 0)
	return u, err
}

// isBzip determines if a file is a bzip file
func (u *uncompress) isBzip() bool {
	return u.MIME == "application/x-bzip2"
		return true
	}
	return false
}

// isXZ determines if a file is a xz file
func (u *uncompress) isXZ() bool {
	if u.MIME == "application/x-xz" {
		return true
	}
	return false
}

// isGzip determines if a file is a gzipped file
func (u *uncompress) isGzip() bool {
	if u.MIME == "application/gzip" {
		return true
	}
	return false
}

// bzipWrappedReader wraps a reader
func (u *uncompress) bzipWrappedReader(r io.Reader) io.Reader {
	return bzip2.NewReader(r)
}

// xzWrappedReader wraps a reader
func (u *uncompress) xzWrappedReader(r io.Reader) (io.Reader, error) {
	return xz.NewReader(r)
}

// gzipWrappedReader wraps a reader
func (u *uncompress) gzipWrappedReader(r io.Reader) (io.Reader, error) {
	return gzip.NewReader(r)
}

func NewReader(f *os.File) (io.Reader, error) {
	u, err := newUncompress(f)
	if err != nil {
		return nil, err
	}
	var r io.Reader
	r = io.Reader(f)
	if u.isBzip() {
		return u.bzipWrappedReader(f), nil
	}
	if u.isXZ() {
		return u.xzWrappedReader(f)
	}
	return r, nil
}
