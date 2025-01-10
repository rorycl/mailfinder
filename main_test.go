//go:build linux
// +build linux

package main

import (
	"io"
	"os"
	"testing"
)

func TestMain(t *testing.T) {

	// override main Exit
	var result int = 0
	Exit := func(n int) {
		result = n
	}
	exit = Exit

	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	tf, err := os.CreateTemp("", "main_")
	if err != nil {
		t.Fatal(err)
	}
	tfName := tf.Name()
	_ = os.Remove(tfName)

	os.Args = []string{
		"--headers",
		"-d", "testdata/maildir/example",
		"-b", "testdata/mbox/testdata/golang.mbox.bz2",
		"-b", "testdata/mbox/testdata/gonuts.mbox",
		"-r", "(?i)golang",
		"-r", `(go install.*21\.0@latest|CVE-2023-39323)`,
		tfName,
	}

	// run main
	main()

	w.Close()
	output, _ := io.ReadAll(r)
	os.Stdout = oldStdout

	if got, want := result, 0; got != want {
		t.Errorf("got %d want %d exit result", got, want)
	}

	if got, want := string(output), "processed 9 found 2 emails\n"; got != want {
		t.Errorf("got\n%swant\n%s", got, want)
	}
	_ = os.Remove(tfName)

}

func TestMainFail(t *testing.T) {

	// override main Exit
	var result int = 0
	Exit := func(n int) {
		result = n
	}
	exit = Exit

	os.Args = []string{
		"--headers",
		"-d", "testdata/maildir/example",
		"-b", "testdata/mbox/testdata/golang.mbox.bz2",
		"-b", "testdata/mbox/testdata/gonuts.mbox",
		"-r", "hi",
		"/dev/null", // only on linux?
	}
	main()
	if got, want := result, 1; got != want {
		t.Errorf("got %d want %d exit result", got, want)
	}
}
