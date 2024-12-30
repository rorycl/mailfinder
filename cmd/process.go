package cmd

import (
	"bytes"
	"fmt"
	"io"

	"github.com/rorycl/mailfinder/finder"
	"github.com/rorycl/mailfinder/mail"
	"github.com/rorycl/mailfinder/maildir"
	"github.com/rorycl/mailfinder/mbox"
)

type ReadNextMail interface {
	NextReader() (*mail.Mail, io.Reader, error)
}

func mailWriter(o *Options) error {

}

func process(options *Options) error {

	type readNextMail interface {
		NextReader() (*mail.Mail, io.Reader, error)
	}

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

	mbw, err := mbox.NewMboxWriter(options.Args.OutputMbox)
	if err != nil {
		return err
	}
	defer mbw.Close()

	for i, mm := range allMboxesAndMailDirs {
		for {
			m, r, err := mm.NextReader()
			if err != nil && err == io.EOF {
				break
			}
			if err != nil {
				return err
			}

			buf := &bytes.Buffer{}
			tee := io.TeeReader(r, buf)

			ok, headers, err := finder.Finder(tee, options.regexes)
			if err != nil {
				return err
			}
			if !ok {
				buf = &bytes.Buffer{}
				continue // hopefully the gc will clean up buf
			}
			fmt.Printf("match: mbox/mdir %d : %s (offset %d)\n", i, m.Path, m.No)

			err = mbw.Add(headers.From[0].Address, headers.Date, buf)
			if err != nil {
				return err
			}
		}
	}
	return nil
}
