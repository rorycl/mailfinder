package main

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"regexp"
	"sync"

	"github.com/rorycl/mailfinder/finder"
	"github.com/rorycl/mailfinder/mail"
	"github.com/rorycl/mailfinder/maildir"
	"github.com/rorycl/mailfinder/mbox"
)

type ReadNextMail interface {
	NextReader() (*mail.Mail, io.Reader, error)
}

type mailBytesId struct {
	m   *mail.Mail
	buf *bytes.Buffer
	i   int
}

func workers(mbw *mbox.MboxWriter, patterns []*regexp.Regexp, reader <-chan mailBytesId) <-chan bool {

	done := make(chan bool)

	var wg sync.WaitGroup
	wg.Add(8)
	for w := 0; w < 8; w++ {
		go func() {
			defer wg.Done()
			for {
				select {
				case mbi, open := <-reader:
					if !open {
						return
					}
					ok, headers, err := finder.Finder(mbi.buf, patterns)
					if err != nil {
						fmt.Println("finder err", err)
						os.Exit(1)
					}
					if !ok {
						mbi.buf = &bytes.Buffer{} // zero buf
						return
					}
					fmt.Printf("match: mbox/mdir %d : %s (offset %d)\n", mbi.i, mbi.m.Path, mbi.m.No)

					// mutex protected
					err = mbw.Add(headers.From[0].Address, headers.Date, mbi.buf)
					if err != nil {
						fmt.Println(err)
						os.Exit(1)
					}
				}
			}
		}()
	}
	go func() {
		wg.Wait()
		done <- true
	}()
	return done
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

	reader := make(chan mailBytesId)
	done := workers(mbw, patterns, reader)

	var wg sync.WaitGroup
	wg.Add(len(ms))
	for i, m := range ms {
		go func(i int, m ReadNextMail) {
			defer wg.Done()
			for {
				n, r, err := m.NextReader()
				if err != nil && err == io.EOF {
					break
				}
				if err != nil {
					fmt.Println("next error", err)
					os.Exit(1)
				}
				b := bytes.Buffer{}
				_, err = b.ReadFrom(r)
				if err != nil {
					fmt.Println("buffer error ", err)
				}
				reader <- mailBytesId{n, &b, i}
			}
		}(i, m)
	}
	wg.Wait()
	close(reader)

	<-done
}
