package main

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"regexp"

	"github.com/rorycl/mailfinder/finder"
	"github.com/rorycl/mailfinder/mail"
	"github.com/rorycl/mailfinder/maildir"
	"github.com/rorycl/mailfinder/mbox"
)

type ReadNextMail interface {
	NextReader() (*mail.Mail, io.Reader, error)
}

func main() {

	mboxes := []string{"mbox/testdata/golang.mbox", "mbox/testdata/gonuts.mbox"}
	mdirs := []string{"maildir/testdata/example/"}

	patterns := []*regexp.Regexp{
		// regexp.MustCompile("bypass.*restrictions"),
		// regexp.MustCompile("IIRC some broken Unices"),
		regexp.MustCompile("golang"),
	}

	m1, err := mbox.NewMbox(mboxes[0])
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	m2, err := mbox.NewMbox(mboxes[1])
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	m3, err := maildir.NewMailDir(mdirs[0])
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	ms := []ReadNextMail{m1, m2, m3}

	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	fn := "/tmp/testOutput.mbox"
	_ = os.Remove(fn)
	mbw, err := mbox.NewMboxWriter(fn)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	defer mbw.Close()

	for i, mm := range ms {
		for {
			m, r, err := mm.NextReader()
			if err != nil && err == io.EOF {
				break
			}
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}

			buf := &bytes.Buffer{}
			tee := io.TeeReader(r, buf)

			ok, headers, err := finder.Finder(tee, patterns)
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
			if !ok {
				continue // hopefully the gc will clean up buf
			}
			fmt.Printf("match: mbox/mdir %d : %s (offset %d)\n", i, m.Path, m.No)

			err = mbw.Add(headers.From[0].Address, headers.Date, buf)
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
		}
	}
}
