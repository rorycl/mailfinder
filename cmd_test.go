package main

import (
	"errors"
	"fmt"
	"os"
	"slices"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestCmdOptions(t *testing.T) {

	tests := []struct {
		desc string
		args []string
		err  bool
	}{

		{
			desc: "no maildirs or mboxes",
			args: []string{"progname", "output.mbox"},
			err:  true,
		},
		{
			desc: "missing mbox",
			args: []string{"progname", "-b", "not-valid", "output.mbox"},
			err:  true,
		},
		{
			desc: "missing maildir",
			args: []string{"progname", "-d", "invalid-mdir", "output.mbox"},
			err:  true,
		},
		{
			desc: "missing regex",
			args: []string{"progname", "-d", "testdata/maildir/example", "output.mbox"},
			err:  true,
		},
		{
			desc: "matcher not long enough",
			args: []string{"progname", "-d", "testdata/maildir/example", "-m", "hi", "output.mbox"},
			err:  true,
		},
		{
			desc: "matcher ok",
			args: []string{"progname", "-d", "testdata/maildir/example", "-m", "hi there", "output.mbox"},
			err:  false,
		},
		{
			desc: "missing output mbox",
			args: []string{"progname", "-d", "testdata/maildir/example", "-r", "hi"},
			err:  true,
		},
		{
			desc: "ok",
			args: []string{"progname", "-d", "testdata/maildir/example", "-r", "hi", "output.mbox"},
			err:  false,
		},
		{
			desc: "not ok headersonly",
			args: []string{"progname", "-o", "-d", "testdata/maildir/example", "-r", "hi", "output.mbox"},
			err:  true,
		},
		{
			desc: "ok headersonly",
			args: []string{"progname", "-a", "-o", "-d", "testdata/maildir/example", "-r", "hi", "output.mbox"},
			err:  false,
		},
		{
			desc: "not enough workers",
			args: []string{"progname", "-d", "testdata/maildir/example", "-w", "0", "-r", "hi", "output.mbox"},
			err:  true,
		},
		{
			desc: "ok workers",
			args: []string{"progname", "-d", "testdata/maildir/example", "-w", "12", "-r", "hi", "output.mbox"},
			err:  false,
		},
		{
			desc: "too many workers",
			args: []string{"progname", "-d", "testdata/maildir/example", "-w", "17", "-r", "hi", "output.mbox"},
			err:  true,
		},
		{
			desc: "ok matchers and regex",
			args: []string{"progname", "-d", "testdata/maildir/example", "-w", "12", "-r", "hi", "-m", "there", "output.mbox"},
			err:  false,
		},
		{
			desc: "ok matchers, regex and from",
			args: []string{"progname", "-d", "testdata/maildir/example", "--from", "-w", "12", "-r", "hi", "-m", "there", "output.mbox"},
			err:  false,
		},
		{
			desc: "ok matchers, regex and datefrom",
			args: []string{"progname", "-d", "testdata/maildir/example", "--datefrom", "2006-01-02", "-w", "12", "-r", "hi", "-m", "there", "output.mbox"},
			err:  false,
		},
		{
			desc: "ok matchers, regex and dateto",
			args: []string{"progname", "-d", "testdata/maildir/example", "--dateto", "2006-01-02", "-w", "12", "-r", "hi", "-m", "there", "output.mbox"},
			err:  false,
		},
		{
			desc: "fail dateto before datefrom",
			args: []string{"progname", "-d", "testdata/maildir/example", "--dateto", "2006-01-02", "--datefrom", "2006-01-03", "-m", "there", "output.mbox"},
			err:  true,
		},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("test_%s", tt.desc), func(t *testing.T) {
			os.Args = tt.args
			_, err := ParseOptions()
			if got, want := (err != nil), tt.err; got != want {
				t.Errorf("want err %t for %s", tt.err, tt.desc)
			}
			fmt.Println(err)
		})
	}
}

func TestParserError(t *testing.T) {
	os.Args = []string{"progname"} // no args
	_, err := ParseOptions()
	if err == nil {
		t.Fatal("expected parsing error")
	}
	var pe ParserError
	if !errors.As(err, &pe) {
		t.Errorf("expected error %T %s to be of type ParserError", err, err)
	} else {
		_ = pe.Error()
	}
}

func TestOptionsFromMailBoxes(t *testing.T) {

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

	o, err := ParseOptions()
	if err != nil {
		t.Fatal(err)
	}
	if got, want := len(o.regexes), 1; got != want {
		t.Errorf("number of regexes got %d want %d", got, want)
	}
	fmt.Printf("%#v\n", o)

	// clean up
	_ = os.Remove(outMboxName)

}

func TestHeaderOptions(t *testing.T) {

	tests := []struct {
		name    string
		options *CmdOptions
		results []string
	}{
		{
			name:    "empty",
			options: &CmdOptions{},
			results: []string{},
		},
		{
			name:    "from",
			options: &CmdOptions{From: true},
			results: []string{"From"},
		},
		{
			name:    "from and subject",
			options: &CmdOptions{From: true, Subject: true},
			results: []string{"From", "Subject"},
		},
		{
			name:    "from, subject and messageID",
			options: &CmdOptions{From: true, Subject: true, MessageID: true},
			results: []string{"From", "Subject", "MessageID"},
		},
		{
			name:    "headers",
			options: &CmdOptions{Headers: true},
			results: []string{"From", "To", "Cc", "Subject", "MessageID"},
		},
		{
			name:    "headers and from",
			options: &CmdOptions{From: true, Headers: true},
			results: []string{"From", "To", "Cc", "Subject", "MessageID"},
		},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("test_%s", tt.name), func(t *testing.T) {
			headers := tt.options.aggregateHeaders()
			slices.Sort(headers)
			slices.Sort(tt.results)
			got, want := headers, tt.results
			if !cmp.Equal(got, want) {
				t.Errorf("header aggregation error %s", cmp.Diff(got, want))
			}
		})
	}
}

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

	o, err := ParseOptions()
	if err != nil {
		t.Fatal(err)
	}
	if got, want := len(o.regexes), 1; got != want {
		t.Errorf("number of regexes got %d want %d", got, want)
	}
	fmt.Printf("%#v\n", o)

}

func TestOptionsSkip(t *testing.T) {

	tests := []struct {
		desc string
		args []string
		err  bool
		skip bool
	}{

		{
			desc: "no skip",
			args: []string{"progname", "-k", "-d", "testdata/maildir/example", "-r", "hi", "output.mbox"},
			err:  false,
			skip: false,
		},
		{
			desc: "skip",
			args: []string{"progname", "-d", "testdata/maildir/example", "-r", "hi", "output.mbox"},
			err:  false,
			skip: true,
		},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("test_%s", tt.desc), func(t *testing.T) {
			os.Args = tt.args
			options, err := ParseOptions()
			if got, want := (err != nil), tt.err; got != want {
				t.Errorf("want err %t for %s", tt.err, tt.desc)
			}
			fmt.Println(err)
			if got, want := options.skipParsingErrors, tt.skip; got != want {
				t.Errorf("skipParsingErrors got %t want %t", got, want)
			}
		})
	}
}
