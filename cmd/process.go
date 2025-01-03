package cmd

import (
	"bytes"
	"fmt"
	"io"
	"regexp"

	"github.com/rorycl/mailfinder/finder"
	"github.com/rorycl/mailfinder/mail"
	"github.com/rorycl/mailfinder/maildir"
	"github.com/rorycl/mailfinder/mbox"
	"golang.org/x/sync/errgroup"
)

// readNextMail is a common interface for mbox, maildir reading
type readNextMail interface {
	NextReader() (*mail.Mail, io.Reader, error)
}

// workerNum is the default number of consumer workers; reset below
var workerNum int = 8

// skipParsingErrors allows the worker to continue after email parsing
// errors
var skipParsingErrors bool = false

// mailBytesId passes mail data from the reader to the worker
type mailBytesId struct {
	m   *mail.Mail
	buf *bytes.Buffer
	i   int // this email offset
}

// testingVerbose allows for some testing output
var testingVerbose bool = false

// workers process mail to see if they match the regex patterns provided
// on the reader chan, and if so write to the mutex protected mailbox
// writer mbw. The reader buf *bytes.Buffer is used because although
//
//	n, r, err := m.NextReader()
//
// in process returns an io.Reader, passing it between goroutines means
// the io.Reader has moved on by the time it is processed. Consequently
// the io.Reader contents are materialised in the buffer.
func workers(mbw *mbox.MboxWriter, patterns []*regexp.Regexp, reader <-chan mailBytesId) <-chan error {

	workerErrChan := make(chan error)
	g := new(errgroup.Group)
	for w := 0; w < workerNum; w++ {
		g.Go(func() error {
			for mbi := range reader {
				bodyBuf := &bytes.Buffer{}
				tee := io.TeeReader(mbi.buf, bodyBuf)
				ok, headers, err := finder.Finder(tee, patterns)
				if err != nil {
					thisErr := fmt.Errorf("worker error at %s offset %d:\n%w", mbi.m.Path, mbi.m.No, err)
					if skipParsingErrors {
						fmt.Println(thisErr)
						continue
					}
					workerErrChan <- thisErr
					close(workerErrChan)
					return thisErr
				}
				if !ok {
					mbi.buf = &bytes.Buffer{} // zero buf
					// bodyBuf = &bytes.Buffer{} // zero buf
					continue
				}
				if testingVerbose {
					fmt.Printf("match: mbox/mdir %d : %s (offset %d)\n", mbi.i, mbi.m.Path, mbi.m.No)
				}

				// mutex protected; checks for duplicate messages
				ok, err = mbw.Add(headers.From[0].Address, headers.Date, string(headers.MessageID), bodyBuf)
				if testingVerbose && !ok {
					fmt.Printf("duplicate message %s not written\n", headers.MessageID)
				}
				if err != nil {
					return err
				}
			}
			return nil
		})
	}
	go func() {
		workerErrChan <- g.Wait()
	}()
	return workerErrChan
}

// Process processes all mailboxes and maildirs in separate
// goroutines for each feeding the emails to the workers func over the
// reader chan.
func Process(options *Options) error {

	allMboxesAndMailDirs := []readNextMail{}
	for _, m := range options.Mboxes {
		b, err := mbox.NewMbox(m)
		if err != nil {
			return fmt.Errorf("register mbox error: %w", err)
		}
		allMboxesAndMailDirs = append(allMboxesAndMailDirs, b)
	}
	for _, m := range options.Maildirs {
		b, err := maildir.NewMailDir(m)
		if err != nil {
			return fmt.Errorf("register maildir error: %w", err)
		}
		allMboxesAndMailDirs = append(allMboxesAndMailDirs, b)
	}

	// set the number of workers (guarded against testing issues)
	if options.Workers > 0 {
		workerNum = options.Workers
	}

	// set whether email reading errors in workers should be skipped
	if options.skipParsingErrors {
		skipParsingErrors = true
	} else {
		skipParsingErrors = false
	}

	// output mbox writer
	mbw, err := mbox.NewMboxWriter(options.Args.OutputMbox)
	if err != nil {
		return err
	}
	defer mbw.Close()

	// reader is a chan for sending emails to workers
	reader := make(chan mailBytesId)

	// initiate email search/write workers
	workerErrChan := workers(mbw, options.regexes, reader)

	// Read each mbox/maildir in a separate goroutine, exiting on first
	// error. Errors from workers are signalled on the workerErrChan,
	// with the first worker error being reported after which the
	// workerErrChan is closed, causing other produer goroutines to
	// exit.
	g := new(errgroup.Group)
	for ii, mm := range allMboxesAndMailDirs {
		g.Go(func() error {
			i := ii
			m := mm
			for {
				// check for error or closed worker chan, exiting in
				// either case
				select {
				case err, ok := <-workerErrChan:
					if err != nil {
						return err
					}
					if !ok {
						return nil
					}
				default:
				}

				n, r, err := m.NextReader()
				if err != nil && err == io.EOF {
					break
				}
				if err != nil {
					return fmt.Errorf("read next mail error: %w", err)
				}
				b := bytes.Buffer{}
				_, err = b.ReadFrom(r)
				if err != nil {
					return fmt.Errorf("buffer error: %w", err)
				}
				reader <- mailBytesId{n, &b, i}
			}
			return nil
		})
	}
	err = g.Wait()
	if err != nil {
		return err
	}
	close(reader) // signal completion to workers

	// wait for workers to complete, possibly with error, if not
	// completed already
	err = <-workerErrChan
	return err
}
