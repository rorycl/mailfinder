package mbox

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"strings"
	"testing"
	"time"
)

func emailCreator(dateString, header string) (time.Time, io.Reader) {
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
	return date, r
}

func TestMboxWriter(t *testing.T) {
	file, err := ioutil.TempFile("", "mboxtest_*.mbox")
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
	date, r := emailCreator("Thu, 01 Jan 2015 01:01:01 +0100", "This is a simple test")

	err = m.Add(from, date, r)
	if err != nil {
		t.Fatal(err)
	}

	// write a second message to the mbox
	from = "test2@example.com"
	date, r = emailCreator("Fri, 02 Jan 2015 02:02:02 +0100", "This is another test")

	err = m.Add(from, date, r)
	if err != nil {
		t.Fatal(err)
	}

	err = m.Close()
	if err != nil {
		t.Fatal(err)
	}

	// comment this out to check the mbox
	_ = os.Remove(f)

}

func TestFailMboxWriter(t *testing.T) {
	file, err := ioutil.TempFile("", "mboxtest_*.mbox")
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
