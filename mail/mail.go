// package mail represents a mail on disk in a maildir or as part of an
// mbox.
package mail

import "fmt"

// Mail represents a Mail file on disk
type Mail struct {
	Path string
	No   int // the item number in the maildir or mbox
}

// String is a string representation of Mail for debugging
func (m Mail) String() string {
	tpl := "path %s item %d"
	return fmt.Sprintf(tpl, m.Path, m.No)
}
