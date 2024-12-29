// package mail represents a mail on disk in a maildir or as part of an
// mbox.
package mail

// Mail represents a Mail file on disk
type Mail struct {
	Path string
	No   int // the item number in the maildir or mbox
}
