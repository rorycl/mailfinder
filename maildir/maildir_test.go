package maildir

import (
	"bufio"
	"errors"
	"io"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestMailDir(t *testing.T) {
	md, err := NewMailDir("testdata/example")
	if err != nil {
		t.Fatal(err)
	}
	stats := map[string]int{
		"cur": 4,
		"new": 2,
	}
	if !cmp.Equal(md.stats, stats) {
		t.Errorf("stats not as expected %s", cmp.Diff(md.stats, stats))
	}

	if got, want := md.TotalEmails(), 6; got != want {
		t.Errorf("total emails got %d want %d", got, want)
	}

	files := `
testdata/example/cur/1735238277.2023287_11.rory-t470s:2,S
testdata/example/cur/1735238277.2023287_5.rory-t470s:2,S
testdata/example/cur/1735238277.2023287_7.rory-t470s:2,S
testdata/example/cur/1735238277.2023287_9.rory-t470s:2,S
testdata/example/new/1735238277.2023287_1.rory-t470s
testdata/example/new/1735238277.2023287_3.rory-t470s
`
	fo := ""
	for _, f := range md.Contents {
		fo += f.Path + "\n"
	}

	if got, want := strings.TrimSpace(fo), strings.TrimSpace(files); got != want {
		t.Errorf("files error got\n%s\nwant\n%s", got, want)
	}

	counter := 0
	for {
		_, err := md.Next()
		if err != nil && err == io.EOF {
			break
		}
		if err != nil {
			t.Fatal(err)
		}
		counter++
	}

	if got, want := counter, 6; got != want {
		t.Errorf("count got %d want %d", got, want)
	}

	md.Reset()

	firstFourLines := ""
	for {
		m, r, err := md.NextReader()
		if err != nil && err == io.EOF {
			break
		}
		if err != nil {
			t.Fatal(err)
		}
		firstFourLines = m.Path + "\n"
		b := bufio.NewReader(r)
		for i := 0; i < 4; i++ {
			line, _, err := b.ReadLine()
			if err != nil {
				t.Fatal(err)
			}
			firstFourLines += string(line) + "\n"
		}
	}

	// check contents of last file
	lastFileHeader := `
testdata/example/new/1735238277.2023287_3.rory-t470s
From: bent at clark.net (Ben Taylor)
Date: Wed, 1 Mar 2000 14:52:56 -0500 (EST)
Subject: Post-compile RSA error with 1.2.2, Solaris 7, OpenSSL 0.9.5
In-Reply-To: <Pine.GSO.4.05.10003011417440.7189-100000@shell.clark.net>`

	if got, want := strings.TrimSpace(firstFourLines), strings.TrimSpace(lastFileHeader); got != want {
		t.Errorf("last file header error got\n%s\nwant\n%s", got, want)
	}
}

func TestMailDirEmpty(t *testing.T) {
	if _, err := NewMailDir("testdata/empty"); !errors.Is(err, EmptyMailDir) {
		t.Fatal("expected empty error")
	}
}
