package cmd

import (
	"fmt"
	"os"
	"testing"
)

func TestOptionsFail(t *testing.T) {

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
			args: []string{"progname", "-d", "../maildir/testdata/example", "output.mbox"},
			err:  true,
		},
		{
			desc: "missing output mbox",
			args: []string{"progname", "-d", "../maildir/testdata/example", "-r", "hi"},
			err:  true,
		},
		{
			desc: "ok",
			args: []string{"progname", "-d", "../maildir/testdata/example", "-r", "hi", "output.mbox"},
			err:  false,
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
