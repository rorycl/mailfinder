package main

import (
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

}

func TestHeaderOptions(t *testing.T) {

	tests := []struct {
		name    string
		options *Options
		results []string
	}{
		{
			name:    "empty",
			options: &Options{},
			results: []string{},
		},
		{
			name:    "from",
			options: &Options{From: true},
			results: []string{"From"},
		},
		{
			name:    "from and subject",
			options: &Options{From: true, Subject: true},
			results: []string{"From", "Subject"},
		},
		{
			name:    "headers",
			options: &Options{Headers: true},
			results: []string{"From", "To", "Cc", "Subject"},
		},
		{
			name:    "headers and from",
			options: &Options{From: true, Headers: true},
			results: []string{"From", "To", "Cc", "Subject"},
		},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("test_%s", tt.name), func(t *testing.T) {
			aggregateHeader(tt.options)
			slices.Sort(tt.options.headers)
			slices.Sort(tt.results)
			got, want := tt.options.headers, tt.results
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
