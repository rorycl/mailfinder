package cmd

import (
	"fmt"
	"os"
	"regexp"
	"testing"
)

func TestProcessFailMboxes(t *testing.T) {

	options := Options{
		Maildirs: []string{"maildir/testdata/example/"},
		Mboxes:   []string{"mbox/testdata/golang.mbox", "mbox/testdata/gonuts.mbox"},
		regexes:  []*regexp.Regexp{regexp.MustCompile("golang")},
	}
	fn := "/tmp/testOutput.mbox"
	_ = os.Remove(fn)
	options.Args.OutputMbox = fn

	err := Process(&options)
	if err == nil {
		t.Fatal(err)
	}
	fmt.Println(err)

}

func TestProcessOK(t *testing.T) {

	testingVerbose = true

	options := Options{
		Maildirs: []string{"../maildir/testdata/example/"},
		Mboxes:   []string{"../mbox/testdata/golang.mbox", "../mbox/testdata/gonuts.mbox"},
		regexes:  []*regexp.Regexp{regexp.MustCompile("(?i)(golang|openssl)")},
	}
	fn := "/tmp/testOutput.mbox"
	_ = os.Remove(fn)
	options.Args.OutputMbox = fn

	err := Process(&options)
	if err != nil {
		t.Fatal(err)
	}
}
