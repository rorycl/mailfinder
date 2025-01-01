package mbox

import (
	"fmt"
	"io"
	"os"
	"strings"
	"testing"
	"time"
)

var id int = 99

func emailCreator(dateString, header string) (time.Time, io.Reader, string) {
	// taken from https://github.com/ProtonMail/go-mbox/blob/master/writer_test.go
	emailTpl := `Date: %s
		
		%s.
		
		And, by the way, this is how a "From" line is escaped in mboxo format:
		
		From Herp Derp with love.
		
		Bye.`
	date, err := time.Parse(time.RFC1123Z, dateString)
	if err != nil {
		panic(err)
	}
	emailTpl = strings.ReplaceAll(emailTpl, "		", "")
	email := fmt.Sprintf(emailTpl, date.Format(time.RFC1123Z), header)
	r := strings.NewReader(email)
	id++
	messageId := fmt.Sprintf("%d", id)
	return date, r, messageId
}

func TestMboxWriter(t *testing.T) {
	file, err := os.CreateTemp("", "mboxtest_*.mbox")
	if err != nil {
		t.Fatal(err)
	}
	f := file.Name()
	file.Close()
	_ = os.Remove(f)

	m, err := NewMboxWriter(f)
	if err != nil {
		t.Fatal(err)
	}

	// write a first message to the mbox
	from := "test1@example.com"
	date, r, id := emailCreator("Thu, 01 Jan 2015 01:01:01 +0100", "This is a simple test")

	ok, err := m.Add(from, date, id, r)
	if err != nil || ok != true {
		t.Fatal(err, ok)
	}

	// write a second message to the mbox
	from = "test2@example.com"
	date, r, id = emailCreator("Fri, 02 Jan 2015 02:02:02 +0100", "This is another test")

	ok, err = m.Add(from, date, id, r)
	if err != nil || ok != true {
		t.Fatal(err, ok)
	}

	// fail to write a third message to the mbox due to duplicate id
	// (reuse id)
	from = "test2@example.com"
	date, r, _ = emailCreator("Fri, 02 Jan 2015 02:02:02 +0100", "This is another test")

	ok, err = m.Add(from, date, id, r)
	if err != nil || ok != false {
		t.Fatal(err, ok)
	}

	err = m.Close()
	if err != nil {
		t.Fatal(err)
	}

	// comment this out to check the mbox
	_ = os.Remove(f)

}

func TestFailMboxWriter(t *testing.T) {
	file, err := os.CreateTemp("", "mboxtest_*.mbox")
	if err != nil {
		t.Fatal(err)
	}
	_, err = NewMboxWriter(file.Name())
	if err == nil {
		t.Fatalf("expected NewMboxWriter to fail with existing file for %s", file.Name())
	}
	file.Close()
	_ = os.Remove(file.Name())
}
