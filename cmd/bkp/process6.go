package cmd

import (
	"bytes"
	"fmt"
	"io"
	"sync"
	"time"

	"github.com/rorycl/mailfinder/finder"
	"github.com/rorycl/mailfinder/mail"
	"github.com/rorycl/mailfinder/maildir"
	"github.com/rorycl/mailfinder/mbox"
)

// mailForWrite has the information required to write an email to an
// mbox.
type mailForWrite struct {
	from string
	date time.Time
	buf  io.Reader
}

type readNextMail interface {
	NextReader() (*mail.Mail, io.Reader, error)
}

func mailWriter(options *Options, mailToWrite <-chan mailForWrite, errorChan chan<- error) {

	mbw, err := mbox.NewMboxWriter(options.Args.OutputMbox)
	if err != nil {
		errorChan <- err
		return
	}
	defer mbw.Close()

	for m := range mailToWrite {
		err := mbw.Add(m.from, m.date, m.buf)
		if err != nil {
			errorChan <- err
			return
		}
	}
}

func readMailBoxes(options *Options, allSources []readNextMail, errorChan chan<- error) <-chan mailForWrite {

	mailToWrite := make(chan mailForWrite)

	var wg sync.WaitGroup
	for ii, mmmm := range allSources {
		wg.Add(1)
		go func(i int, n readNextMail) {
			defer wg.Done()
			for {
				m, r, err := n.NextReader()
				if err != nil && err == io.EOF {
					break
				}
				if err != nil {
					errorChan <- err
					return
				}

				buf := &bytes.Buffer{}
				tee := io.TeeReader(r, buf)

				ok, headers, err := finder.Finder(tee, options.regexes)
				if err != nil {
					errorChan <- err
					return
				}
				if !ok {
					buf = &bytes.Buffer{} // zero buf
					continue
				}
				fmt.Printf("match: mbox/mdir %d : %s (offset %d)\n", i, m.Path, m.No)

				// err = mbw.Add(headers.From[0].Address, headers.Date, buf)
				mailToWrite <- mailForWrite{headers.From[0].Address, headers.Date, buf}
			}
			return
		}(ii, mmmm)
	}
	go func() {
		wg.Wait()
		close(mailToWrite)
	}()

	return mailToWrite
}

func process(options *Options) error {

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

	// setup mail writer
	errorChan := make(chan error)

	// drain the errorChan
	go func() {
		for e := range errorChan {
			panic(e)
		}
	}()

	mailToWrite := readMailBoxes(options, allMboxesAndMailDirs, errorChan)
	go mailWriter(options, mailToWrite, errorChan)

	return nil
}
