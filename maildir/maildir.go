// package maildir provides a very simple way of accessing the emails in an
// email Maildir.
//
// for example:
//
//	md, _ := NewMailDir("openssh")
//	_ = md.List()
//	for _, m := range md.Contents {
//		fmt.Println(m)
//	}
//	for {
//		m, r, err := md.NextReader()
//		if err != nil && err == io.EOF {
//			break
//		}
//		fmt.Println(m.Path)
//	   // do something with the io.Reader in r
//	}
package maildir

import (
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
)

// mailDirContents are the normal subdirectory names of an enclosing
// maildir; see https://en.wikipedia.org/wiki/Maildir#Specifications
var mailDirContents []string = []string{"cur", "new", "tmp"}

// EmptyMailDir is a sentinel error
var EmptyMailDir error = errors.New("maildir is empty")

// Mail represents a Mail file on disk
type Mail struct {
	Directory string
	Name      string
	Path      string
}

// String is a string representation of Mail for debugging
func (m Mail) String() string {
	tpl := "dir %s name %s path %s"
	return fmt.Sprintf(tpl, m.Directory, m.Name, m.Path)
}

// Maildir represents the outer directory of a set of maildir
// subdirectories (expected to be mailDirContents) and a listing of the
// Mail items (if any) in each subdirectory.
type MailDir struct {
	Path     string
	Contents []*Mail
	stats    map[string]int
	current  int // current message being read
}

// NewMailDir sets up a mail directory for listing the contents.
func NewMailDir(path string) (*MailDir, error) {
	m := MailDir{}
	_, err := os.Stat(path)
	if errors.Is(err, fs.ErrNotExist) {
		return &m, err
	}
	m.Path = path
	m.stats = map[string]int{}
	m.current = -1
	err = m.list()
	if m.TotalEmails() == 0 {
		return &m, EmptyMailDir
	}
	return &m, err
}

// list lists all of the contents of the mail directory's
// subdirectories.
func (m *MailDir) list() error {
	for _, md := range mailDirContents {
		dirPathForGlob := filepath.Join(m.Path, md, "*")
		contents, err := filepath.Glob(dirPathForGlob)
		if err != nil {
			return err
		}
		if len(contents) == 0 {
			continue
		}
		m.stats[md] = len(contents)
		for _, c := range contents {
			mail := &Mail{
				Directory: md,
				Name:      filepath.Base(c),
				Path:      filepath.Join(c),
			}
			m.Contents = append(m.Contents, mail)
		}
	}
	return nil
}

// TotalEmails provides a total of all mails in all Maildir subdirectories
func (m *MailDir) TotalEmails() int {
	n := 0
	for _, v := range m.stats {
		n += v
	}
	return n
}

// Next returns the next Mail in MailDir.contents or io.EOF when
// the contents are exhausted.
func (m *MailDir) Next() (*Mail, error) {
	m.current++
	if m.current > len(m.Contents)-1 {
		return nil, io.EOF
	}
	return m.Contents[m.current], nil
}

// NextReader returns the next Mail in MailDir.contents as  Mail
// metadata and io.Reader unless the contents are exhausted.
func (m *MailDir) NextReader() (*Mail, io.Reader, error) {
	m.current++
	if m.current > len(m.Contents)-1 {
		return nil, nil, io.EOF
	}
	f, err := os.Open(m.Contents[m.current].Path)
	if err != nil {
		return nil, nil, fmt.Errorf("file opening error %w", err)
	}
	return m.Contents[m.current], io.Reader(f), nil
}

// Reset sets the MailDir internal pointer back to -1 to re-read the
// contents of the directories for Next() or NextReader().
func (m *MailDir) Reset() {
	m.current = -1
}
