package mbox

import (
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"sync"
	"time"

	protonMbox "github.com/ProtonMail/go-mbox"
)

// MboxWriter wraps a protonMbox.Writer
type MboxWriter struct {
	Writer *protonMbox.Writer
	ids    map[string]struct{}
	sync.Mutex
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
	m := &MboxWriter{
		Writer: protonMbox.NewWriter(f),
		ids:    map[string]struct{}{},
	}
	return m, nil
}

// Add creates an mbox entry with the email From, Date and messageId and
// contents (as an io.Reader). See protonMbox.CreateMessage for more
// detail. Messages with ids that have already been written are not
// written again. Add returns a bool indicating if the message was
// written or error.
func (m *MboxWriter) Add(from string, date time.Time, messageId string, r io.Reader) (bool, error) {
	m.Lock()
	defer m.Unlock()
	if _, ok := m.ids[messageId]; ok {
		return false, nil
	}
	w, err := m.Writer.CreateMessage(from, date)
	if err != nil {
		return false, fmt.Errorf("mboxwriter create error: %w", err)
	}
	_, err = io.Copy(w, r)
	if err != nil {
		return false, fmt.Errorf("mboxwriter write error: %w", err)
	}
	m.ids[messageId] = struct{}{}
	return true, nil
}

// Close calls protonMbox.Close()
func (m *MboxWriter) Close() error {
	return m.Writer.Close()
}
