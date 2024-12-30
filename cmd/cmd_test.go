package cmd

import (
	"fmt"
	"os"
	"testing"
)

func TestOptions(t *testing.T) {

	mBox, err := os.CreateTemp("", "test_options_*.mbox")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(mBox.Name())

	mDir, err := os.MkdirTemp("", "test_options_*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(mDir)

	outMbox, err := os.CreateTemp("", "test_options_*.mbox")
	if err != nil {
		t.Fatal(err)
	}
	outMboxName := outMbox.Name()
	_ = os.Remove(outMboxName)

	os.Args = []string{
		"progname",
		"-d",
		mDir,
		"-b",
		mBox.Name(),
		"-r",
		"(abc|def)",
		outMboxName,
	}

	o, err := parseOptions()
	if err != nil {
		t.Fatal(err)
	}
	if got, want := len(o.regexes), 1; got != want {
		t.Errorf("number of regexes got %d want %d", got, want)
	}
	fmt.Printf("%#v\n", o)

}
