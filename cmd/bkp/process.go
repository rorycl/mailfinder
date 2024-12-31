package cmd

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"time"

	"github.com/rorycl/mailfinder/finder"
	"github.com/rorycl/mailfinder/mail"
	"github.com/rorycl/mailfinder/mbox"
)

var workersNums int = 8

func process(options *Options) error {

	c := context.Background()
	ctx, cancel := context.WithCancel(c)

	// channels and related types
	type mailForChan struct {
		from string
		date time.Time
		buf  io.Writer
	}
	mailToWrite := make(chan mailForChan)

	type mailReader struct {
		m *mail.Mail
		r io.Reader
	}
	mailToRead := make(chan mailReader)

	errChan := make(chan error)

	// output mbox writer
	mbw, err := mbox.NewMboxWriter(options.Args.OutputMbox)
	if err != nil {
		return err
	}
	go func() {
		defer mbw.Close()

		for {
			select {
			case <-ctx.Done():
				return
			case m, isOpen := <-mailToWrite:
				err := mbw.Add(m.from, m.date, m.buf)
				if err != nil {
				}
			}
		}
	}()

	// consumers
	for i := 0; i < len(workersNums); i++ {
		go func() {
			for {
				select {
				case <-ctx.Done():
					return
				case mr, isOpen := <-mailToRead:
					if !isOpen {
						return
					}
					buf := &bytes.Buffer{}
					tee := io.TeeReader(mr.r, buf)
					ok, headers, err := finder.Finder(tee, patterns)
					if err != nil {
						errChan <- fmt.Errorf("finder error: %w", err)
					}
					if ok {
						mailToWrite <- mailForChan{
							from: headers.From[0].Address,
							date: headers.Date,
							buf:  buf,
						}
					}
				}
			}
		}()
	}

	// producers
	allMboxesAndMailDirs := []ReadNextMail{}
	for _, m := range options.Mboxes {
		b, err := mbox.NewMbox(m)
		if err != nil {
			return err
		}
		allMboxesAndMailDirs = append(allMboxesAndMailDirs, b)
	}
	for _, m := range options.Maildirs {
		b, err := mbox.NewMailDir(m)
		if err != nil {
			return err
		}
		allMboxesAndMailDirs = append(allMboxesAndMailDirs, b)
	}

	// process all mailboxes and maildirs
	for i, mm := range ms {
		for {
			m, r, err := mm.NextReader()
			if err != nil && err == io.EOF {
				break
			}
			if err != nil {
				cancel()
				return err
			}
			mailToRead <- mailReader{m, r}
		}
	}

	close(mailToRead)

}
