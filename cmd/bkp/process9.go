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

	"golang.org/x/sync/errgroup"
)

type ReadNextMail interface {
	NextReader() (*mail.Mail, io.Reader, error)
}

type mailBytesId struct {
	m   *mail.Mail
	buf *bytes.Buffer
	i   int
}

func workers(mbw *mbox.MboxWriter, patterns []*regexp.Regexp, reader <-chan mailBytesId) <-chan error {

	workerErrChan := make(chan error)

	g := new(errgroup.Group)
	for w := 0; w < 8; w++ {
		g.Go(func() error {
			for {
				select {
				case mbi, open := <-reader:
					if !open {
						return nil
					}
					ok, headers, err := finder.Finder(mbi.buf, patterns)
					if err != nil {
						return err
					}
					if !ok {
						mbi.buf = &bytes.Buffer{} // zero buf
						return nil
					}
					fmt.Printf("match: mbox/mdir %d : %s (offset %d)\n", mbi.i, mbi.m.Path, mbi.m.No)

					// mutex protected
					err = mbw.Add(headers.From[0].Address, headers.Date, mbi.buf)
					if err != nil {
						return err
					}
				}
			}
		})
	}
	go func() {
		workerErrChan <- g.Wait()
	}()
	return workerErrChan
}

func process() error {

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

	reader := make(chan mailBytesId)
	workerErrChan := workers(mbw, patterns, reader)

	g := new(errgroup.Group)

	for ii, mm := range ms {
		g.Go(func() error {
			i := ii
			m := mm
			for {
				n, r, err := m.NextReader()
				if err != nil && err == io.EOF {
					break
				}
				if err != nil {
					return fmt.Errorf("next error", err)
				}
				b := bytes.Buffer{}
				_, err = b.ReadFrom(r)
				if err != nil {
					return fmt.Errorf("buffer error ", err)
				}
				reader <- mailBytesId{n, &b, i}
			}
			return nil
		})
	}
	err = g.Wait()
	close(reader)
	if err != nil {
		return err
	}

	err = <-workerErrChan
	return err
}

func main() {

	err := process()
	fmt.Println("main", err)

}
