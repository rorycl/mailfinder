package cmd

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"sync"
	"time"

	"github.com/rorycl/mailfinder/finder"
	"github.com/rorycl/mailfinder/mail"
	"github.com/rorycl/mailfinder/maildir"
	"github.com/rorycl/mailfinder/mbox"
)

type ReadNextMail interface {
	NextReader() (*mail.Mail, io.Reader, error)
}

// workerNum is the number of consumer workers
var workerNum int = 8

// mailForWrite has the information required to write an email to an
// mbox.
type mailForWrite struct {
	from string
	date time.Time
	buf  io.Reader
}

// mailReader contains the information for filtering and reading emails.
type mailReader struct {
	m *mail.Mail
	r io.Reader
}

// ProcessMailboxes processes all mailboxes and maildirs in separate
// goroutines for each, putting results on a mailReader channel.
// Downstream consumers are signalled for early exit with a
// withcancel context.Context.
func processMailboxes(ctx context.Context, options *Options, errChan chan<- error) <-chan mailReader {

	type readNextMail interface {
		NextReader() (*mail.Mail, io.Reader, error)
	}

	mailToRead := make(chan mailReader)

	allMboxesAndMailDirs := []readNextMail{}
	for _, m := range options.Mboxes {
		b, err := mbox.NewMbox(m)
		if err != nil {
			errChan <- fmt.Errorf("register mbox error: %w", err)
			return mailToRead
		}
		allMboxesAndMailDirs = append(allMboxesAndMailDirs, b)
	}
	for _, m := range options.Maildirs {
		b, err := maildir.NewMailDir(m)
		if err != nil {
			errChan <- fmt.Errorf("register maildir error: %w", err)
			return mailToRead
		}
		allMboxesAndMailDirs = append(allMboxesAndMailDirs, b)
	}

	// process each mbox/maildir in a goroutine
	var wg sync.WaitGroup
	wg.Add(len(allMboxesAndMailDirs))

	for _, mm := range allMboxesAndMailDirs {
		go func() {
			defer wg.Done()
			for {
				// return early on cancelled context
				select {
				case <-ctx.Done():
					return
				default:
				}
				// pick the next mail off the reader
				m, r, err := mm.NextReader()
				if err != nil && err == io.EOF {
					break
				}
				if err != nil {
					fmt.Println(err)
					errChan <- fmt.Errorf("process mailboxes error: %w", err)
					return
				}
				mailToRead <- mailReader{m, r}
			}
		}()
	}

	// wait for the mailbox/maildir processes to finish reading,
	// then close the mailReader channel
	go func() {
		wg.Wait()
		close(mailToRead)
	}()

	return mailToRead
}

// consumeMail consumes email from the mailToRead chan sending email
// that has passed regex filtering to the mailForWrite chan.
func consumeMail(
	ctx context.Context,
	options *Options,
	errChan chan<- error,
	mailToRead <-chan mailReader,
) <-chan mailForWrite {

	mailToWrite := make(chan mailForWrite)

	var wg sync.WaitGroup
	wg.Add(workerNum)
	for i := 0; i < workerNum; i++ {
		go func() {
			defer wg.Done()
			for {
				select {
				case <-ctx.Done():
					return
				case mr, isOpen := <-mailToRead:
					if !isOpen {
						return
					}
					// copy reader to send to both the finder and
					// possibly (if the finder returns true) to the
					// mailToWrite chan
					buf := &bytes.Buffer{}
					tee := io.TeeReader(mr.r, buf)
					ok, headers, err := finder.Finder(tee, options.regexes)
					if err != nil {
						fmt.Println(err)
						errChan <- fmt.Errorf("finder error: %w", err)
						return
					}
					if !ok {
						continue
					}
					mailToWrite <- mailForWrite{
						from: headers.From[0].Address,
						date: headers.Date,
						buf:  buf,
					}
				}
			}
		}()
	}
	go func() {
		wg.Wait()
		close(mailToWrite)
	}()
	return mailToWrite
}

// mboxWrite writes messages received on the mailToWrite chan to the
// mbox specified in options. This func is not concurrent safe.
func mboxWrite(ctx context.Context, options *Options, mailToWrite <-chan mailForWrite, errChan chan<- error) {

	// output mbox writer
	mbw, err := mbox.NewMboxWriter(options.Args.OutputMbox)
	if err != nil {
		errChan <- err
		return
	}
	defer mbw.Close()

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			select {
			case <-ctx.Done():
				return
			case m, isOpen := <-mailToWrite:
				fmt.Println("qqqqqqqqqqq")
				if !isOpen {
					return
				}
				err := mbw.Add(m.from, m.date, m.buf)
				if err != nil {
					errChan <- fmt.Errorf("mbox writer error %w", err)
					return
				}
			}
		}
	}()
	go func() {
		wg.Wait()
	}()
}

// process joins the processMailboxes, consumeMail and mboxWrite
// functions to 1. process the mboxes and maildirs to provide readers
// for each; 2. process each mail to see if it meets the filtering
// requirements 3. write emails that pass filtering to write these to an
// output mailbox. processMailboxes (1) and consumeMail (2) are
// concurrent while mboxWrite is not concurrent safe. The function
// returns a read-only error chan.
func process(options *Options) <-chan error {

	c := context.Background()
	ctx, cancel := context.WithCancel(c)

	errChan := make(chan error)
	defer close(errChan)
	returnErrChan := make(chan error)

	mailReader := processMailboxes(ctx, options, errChan)
	mailForWrite := consumeMail(ctx, options, errChan, mailReader)

	// drain the error channel, exiting on first error
	go func() {
		for err := range errChan {
			cancel()
			returnErrChan <- err
			close(returnErrChan)
			return
		}
	}()

	// wait for writing to close
	go mboxWrite(ctx, options, mailForWrite, errChan)

	return returnErrChan
}
