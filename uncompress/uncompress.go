// package uncompress tries to determine the file type of a file and
// provides a reader to return an io.Reader wrapped by an uncompress
// reader (such as a bzip2, xz or gzip reader) depending on the
// MIME type determined by the file type check.
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

// newUncompress returns a new uncompress type which attempts to
// determing the file type.
func newUncompress(f *os.File) (*uncompress, error) {
	head := make([]byte, 261)
	_, err := f.Read(head)
	if err != nil {
		return nil, err
	}

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

// IsType reports if the file type is considered to be one of the
// supplied types described by typer.
func (u *uncompress) IsType(typer string) bool {
	switch typer {
	case "bzip", "bzip2", "bz2", "application/x-bzip2":
		return u.MIME == "application/x-bzip2"
	case "xz", "application/x-xz":
		return u.MIME == "application/x-xz"
	case "gzip", "gz":
		return u.MIME == "application/gzip"
	case "unknown":
		return u.MIME == ""
	default:
		return false
	}
}

// NewReader opens a file and attempts to determine its file type.
// Depending on the file type, it will return an io.Reader wrapped by a
// decompression reader.
//
// The decompressions type depends on the determined mime type. Note
// that the bzip reader returns just:
//
//	io.Reader
//
// whereas others return:
//
//	io.Reader, error
func NewReader(f *os.File) (io.Reader, error) {
	u, err := newUncompress(f)
	if err != nil {
		return nil, err
	}

	r := io.Reader(f)

	switch u.MIME {
	case "application/x-bzip2":
		return bzip2.NewReader(r), nil
	case "application/x-xz":
		return xz.NewReader(r)
	case "application/gzip":
		return gzip.NewReader(r)
	}
	return r, nil
}
