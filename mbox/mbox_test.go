package mbox

import (
	"bufio"
	"errors"
	"io"
	"io/fs"
	"strings"
	"testing"
)

func TestMbox(t *testing.T) {
	md, err := NewMbox("testdata/golang.mbox")
	if err != nil {
		t.Fatal(err)
	}

	counter := 0
	firstFourLines := ""
	for {
		m, r, err := md.NextReader()
		if err != nil && err == io.EOF {
			break
		}
		if err != nil {
			t.Fatal(err)
		}
		counter++
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

	if got, want := counter, 2; got != want {
		t.Errorf("counter got %d want %d", got, want)
	}

	// check contents of last file
	lastFileHeader := `
testdata/golang.mbox
Return-path: <golang-nuts+bncBAABB5477SUAMGQE2GXYLLA@googlegroups.com>
Envelope-to: example@test.com
Delivery-date: Thu, 05 Oct 2023 19:35:22 +0000
Received: from mail-oo1-f59.google.com ([209.85.161.59])`

	if got, want := strings.TrimSpace(firstFourLines), strings.TrimSpace(lastFileHeader); got != want {
		t.Errorf("last file header error got\n%s\nwant\n%s", got, want)
	}
}

func TestMissingMbox(t *testing.T) {
	if _, err := NewMbox("testdata/null"); !errors.Is(err, fs.ErrNotExist) {
		t.Fatalf("expected fs ErrNotExistErr, got %s", err)
	}
}

func TestInvalidMbox(t *testing.T) {
	md, err := NewMbox("testdata/empty")
	if err != nil {
		t.Fatal(err)
	}
	counter := 0
	for {
		_, _, err := md.NextReader()
		if err != nil && err == io.EOF {
			break
		}
		if err != nil {
			t.Fatal(err)
		}
		counter++
	}
	if got, want := counter, 0; got != want {
		t.Errorf("counter got %d want %d", got, want)
	}
}
