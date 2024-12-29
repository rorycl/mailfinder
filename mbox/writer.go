package mbox

import (
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"time"

	protonMbox "github.com/ProtonMail/go-mbox"
)

// MboxWriter wraps a protonMbox.Writer
type MboxWriter struct {
	Writer *protonMbox.Writer
}

// NewMboxWriter wraps a proton mbox writer with some file checking
func NewMboxWriter(path string) (*MboxWriter, error) {
	if _, err := os.Stat(path); !errors.Is(err, fs.ErrNotExist) {
		return nil, fmt.Errorf("file %s already exists", path)
	}
	f, err := os.Create(path)
	if err != nil {
		return nil, err
	}
	return &MboxWriter{protonMbox.NewWriter(f)}, nil
}

// Add creates an mbox entry with the email From, Date and contents (as
// an io.Reader). See protonMbox.CreateMessage for more detail.
func (m *MboxWriter) Add(from string, date time.Time, r io.Reader) error {
	w, err := m.Writer.CreateMessage(from, date)
	if err != nil {
		return fmt.Errorf("mboxwriter create error: %w", err)
	}
	_, err = io.Copy(w, r)
	if err != nil {
		return fmt.Errorf("mboxwriter write error: %w", err)
	}
	return nil
}

// Close calls protonMbox.Close()
func (m *MboxWriter) Close() error {
	return m.Writer.Close()
}
