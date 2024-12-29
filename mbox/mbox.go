package mbox

import (
	"io"
	"os"

	"github.com/rorycl/mailfinder/mail"

	mbox "github.com/ProtonMail/go-mbox"
)

// Mbox represents an mbox file on disk with related go-mbox reader and
// email position in the mbox file.
type Mbox struct {
	Path    string
	current int // current message being read
	file    *os.File
	reader  *mbox.Reader
}

// NewMbox sets up a new mbox for reading
func NewMbox(path string) (*Mbox, error) {
	m := Mbox{}
	var err error
	m.file, err = os.Open(path)
	if err != nil {
		return &m, err
	}
	m.Path = path
	m.current = -1
	m.reader = mbox.NewReader(m.file)
	return &m, err
}

// NextReader returns the next Mail from the reader until exhausted
func (m *Mbox) NextReader() (*mail.Mail, io.Reader, error) {
	m.current++
	thisMail := mail.Mail{
		Path: m.Path,
		No:   m.current,
	}
	reader, err := m.reader.NextMessage()
	if err != nil && err == io.EOF {
		m.file.Close()
	}
	return &thisMail, reader, err
}
