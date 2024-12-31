package cmd

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"sync"
	"time"

	"github.com/rorycl/mailfinder/finder"
	"github.com/rorycl/mailfinder/mail"
	"github.com/rorycl/mailfinder/maildir"
	"github.com/rorycl/mailfinder/mbox"
)

type readNextMail interface {
	NextReader() (*mail.Mail, io.Reader, error)
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

	// mailForWrite has the information required to write an email to an
	// mbox.
	type mailForWrite struct {
		from string
		date time.Time
		buf  io.Reader
	}

	// mailbox output writer
	mboxWriter := make(chan mailForWrite)
	fn := options.Args.OutputMbox
	_ = os.Remove(fn)
	mbw, err := mbox.NewMboxWriter(fn)
	if err != nil {
		return err
	}
	defer mbw.Close()

	// mailbox writing in goroutine loop to only allow one write at a
	// time
	go func() {
		for m := range mboxWriter {
			err := mbw.Add(m.from, m.date, m.buf)
			if err != nil {
				panic(err)
			}
		}
	}()

	// iterate over different mboxes and maildirs to find regexes,
	// writing found mail to the output mailbox via mboxWriter.
	var wg sync.WaitGroup
	wg.Add(len(allMboxesAndMailDirs))
	for i, mm := range allMboxesAndMailDirs {
		go func() {
			defer wg.Done()
			for {
				m, r, err := mm.NextReader()
				if err != nil && err == io.EOF {
					break
				}
				if err != nil {
					panic(err)
				}

				buf := &bytes.Buffer{}
				tee := io.TeeReader(r, buf)

				ok, headers, err := finder.Finder(tee, options.regexes)
				if err != nil {
					panic(err)
				}
				if !ok {
					buf = &bytes.Buffer{} // zero
					continue
				}
				fmt.Printf("match: mbox/mdir %d : %s (offset %d)\n", i, m.Path, m.No)

				mboxWriter <- mailForWrite{
					from: headers.From[0].Address,
					date: headers.Date,
					buf:  buf,
				}

			}
		}()
	}
	wg.Wait()
	return nil
}
