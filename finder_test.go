package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"regexp"
	"testing"

	"github.com/mnako/letters"
)

func TestSearchText(t *testing.T) {

	tests := []struct {
		contents string
		patterns []*regexp.Regexp
		ok       bool
	}{
		{
			contents: "abc",
			patterns: []*regexp.Regexp{
				regexp.MustCompile("ab"),
			},
			ok: true,
		},
		{
			contents: "abc\ndef",
			patterns: []*regexp.Regexp{
				regexp.MustCompile(".bc"),
				regexp.MustCompile("d[en]."),
			},
			ok: true,
		},
		{
			contents: "ABC\ndef",
			patterns: []*regexp.Regexp{
				regexp.MustCompile("(?i).bc"),
				regexp.MustCompile("d[en]."),
			},
			ok: true,
		},
		{
			contents: "abc\ndef",
			patterns: []*regexp.Regexp{
				regexp.MustCompile("ab"),
				regexp.MustCompile("z"),
			},
			ok: false,
		},
		{
			contents: "abc\ndef",
			patterns: []*regexp.Regexp{
				regexp.MustCompile("g"),
			},
			ok: false,
		},
	}

	for i, tt := range tests {

		mm := matchRegexpCount{}
		t.Run(fmt.Sprintf("test-%d", i), func(t *testing.T) {
			f := Finder{searchers: tt.patterns}
			got := f.searchText(tt.contents, mm)
			if got, want := got, tt.ok; got != want {
				t.Errorf("got %t want %t", got, want)
			}
		})
	}
}

func TestSearchHTML(t *testing.T) {

	tests := []struct {
		contents string
		patterns []*regexp.Regexp
		ok       bool
	}{
		{
			contents: "<h1>hello</h1><p>there</p>",
			patterns: []*regexp.Regexp{
				regexp.MustCompile("hello"),
				regexp.MustCompile("there"),
			},
			ok: true,
		},
		{
			contents: "<h1>hello</h1><p>there</p>",
			patterns: []*regexp.Regexp{
				regexp.MustCompile("h1"),
			},
			ok: false,
		},
	}

	for i, tt := range tests {
		mm := matchRegexpCount{}
		f := Finder{searchers: tt.patterns}
		t.Run(fmt.Sprintf("test-%d", i), func(t *testing.T) {
			got := f.searchHTML(tt.contents, mm)
			if got, want := got, tt.ok; got != want {
				t.Errorf("got %t want %t", got, want)
			}
		})
	}
}

func TestSearchHeaders(t *testing.T) {

	tests := []struct {
		desc      string
		emailFile string
		patterns  []*regexp.Regexp
		keys      []string
		num       int
	}{
		{
			desc:      "to ok",
			emailFile: "testdata/error.eml",
			patterns: []*regexp.Regexp{
				regexp.MustCompile("exampleto"),
			},
			keys: []string{"To"},
			num:  1,
		},
		{
			desc:      "to not ok",
			emailFile: "testdata/sync.eml",
			patterns: []*regexp.Regexp{
				regexp.MustCompile("notExample"),
			},
			keys: []string{"To"},
			num:  0,
		},
		{
			desc:      "to and from ok",
			emailFile: "testdata/error.eml",
			patterns: []*regexp.Regexp{
				regexp.MustCompile("examplefrom"),
				regexp.MustCompile("exampleto"),
			},
			keys: []string{"To", "From"},
			num:  2,
		},
		{
			desc:      "to from and subject ok",
			emailFile: "testdata/error.eml",
			patterns: []*regexp.Regexp{
				regexp.MustCompile("exampleto"),
				regexp.MustCompile("examplefrom"),
				regexp.MustCompile("^error email$"),
			},
			keys: []string{"To", "From", "Subject"},
			num:  3,
		},
		{
			desc:      "to from and subject partially ok",
			emailFile: "testdata/error.eml",
			patterns: []*regexp.Regexp{
				regexp.MustCompile("exampleto"),
				regexp.MustCompile("examplefrom"),
				regexp.MustCompile("^error email not match$"),
			},
			keys: []string{"To", "From", "Subject"},
			num:  2,
		},
	}
	for _, tt := range tests {
		file, err := os.Open(tt.emailFile)
		if err != nil {
			t.Fatal(err)
		}
		defer file.Close()
		f := Finder{searchers: tt.patterns, headerKeys: tt.keys}
		email, err := letters.ParseEmail(file)
		if err != nil {
			t.Fatal(err)
		}
		t.Run(fmt.Sprintf("test_%s", tt.desc), func(t *testing.T) {
			if got, want := len(f.searchHeaders(email.Headers)), tt.num; got != want {
				t.Errorf("got %d matches want %d", got, want)
			}
		})
	}
}

func TestFinder(t *testing.T) {

	tests := []struct {
		file             string
		patterns         []*regexp.Regexp
		processed, found int
	}{
		{
			file: "testdata/test_txt.eml",
			patterns: []*regexp.Regexp{
				regexp.MustCompile("test.*golang"),
				regexp.MustCompile("(?i)this section"),
			},
			processed: 1,
			found:     1,
		},
		{
			file: "testdata/test_html.eml",
			patterns: []*regexp.Regexp{
				regexp.MustCompile("test.*golang"),
				regexp.MustCompile("(?i)this section"),
			},
			processed: 1,
			found:     1,
		},
		{
			file: "testdata/test_enriched.eml",
			patterns: []*regexp.Regexp{
				regexp.MustCompile("test.*golang"),
				regexp.MustCompile("(?i)this section"),
			},
			processed: 1,
			found:     1,
		},
		{
			file: "testdata/test_txt.eml",
			patterns: []*regexp.Regexp{
				regexp.MustCompile("This is not a test"),
			},
			processed: 1,
			found:     0,
		},
	}

	for i, tt := range tests {
		t.Run(fmt.Sprintf("test-%d", i), func(t *testing.T) {
			mailFile, err := os.Open(tt.file)
			if err != nil {
				t.Fatal(err)
			}
			defer mailFile.Close()

			outFile, err := ioutil.TempFile("", "finder_")
			if err != nil {
				t.Fatal(err)
			}
			outFileName := outFile.Name()
			_ = os.Remove(outFileName)

			f, err := NewFinder(outFileName, tt.patterns)
			if err != nil {
				t.Fatal(err)
			}
			err = f.Operate(mailFile)
			if err != nil {
				t.Fatal(err)
			}

			if got, want := f.processed, tt.processed; got != want {
				t.Errorf("processed got %d want %d", got, want)
			}
			if got, want := f.found, tt.found; got != want {
				t.Errorf("found got %d want %d", got, want)
			}
			_ = os.Remove(outFileName)
		})
	}

}

func TestHeaderAndBodyFinder(t *testing.T) {

	tests := []struct {
		file             string
		patterns         []*regexp.Regexp
		keys             []string
		processed, found int
	}{
		{
			file: "testdata/test_txt.eml",
			patterns: []*regexp.Regexp{
				regexp.MustCompile("test.*golang"),
				regexp.MustCompile("(?i)this section"),
				regexp.MustCompile("thisexample.*gmail.com"),
			},
			keys:      []string{"To", "From", "Subject"},
			processed: 1,
			found:     1,
		},
		{
			file: "testdata/test_txt.eml",
			patterns: []*regexp.Regexp{
				regexp.MustCompile("test.*golang"),
				regexp.MustCompile("(?i)this section"),
				regexp.MustCompile("can't match this"),
			},
			keys:      []string{"To", "From", "Subject"},
			processed: 1,
			found:     0,
		},
		{
			file: "testdata/test_html.eml",
			patterns: []*regexp.Regexp{
				regexp.MustCompile("test.*golang"),
				regexp.MustCompile("(?i)this section"),
				regexp.MustCompile("(?i)example user"),
			},
			keys:      []string{"To"},
			processed: 1,
			found:     1,
		},
		{
			file: "testdata/test_html.eml",
			patterns: []*regexp.Regexp{
				regexp.MustCompile("test.*golang"),
				regexp.MustCompile("(?i)this section"),
				regexp.MustCompile("(?i)test.*golang"),
			},
			keys:      []string{"Subject"},
			processed: 1,
			found:     1,
		},
		{
			file: "testdata/test_html.eml",
			patterns: []*regexp.Regexp{
				regexp.MustCompile("test.*golang"),
				regexp.MustCompile("(?i)this section"),
				regexp.MustCompile("not an example"),
			},
			keys:      []string{"To"},
			processed: 1,
			found:     0,
		},
		{
			file: "testdata/test_enriched.eml",
			patterns: []*regexp.Regexp{
				regexp.MustCompile("test.*golang"),
				regexp.MustCompile("(?i)this section"),
			},
			keys:      []string{},
			processed: 1,
			found:     1,
		},
		{
			file: "testdata/test_txt.eml",
			patterns: []*regexp.Regexp{
				regexp.MustCompile("This is not a test"),
			},
			keys:      []string{},
			processed: 1,
			found:     0,
		},
	}

	for i, tt := range tests {
		t.Run(fmt.Sprintf("test-%d", i), func(t *testing.T) {
			mailFile, err := os.Open(tt.file)
			if err != nil {
				t.Fatal(err)
			}
			defer mailFile.Close()

			outFile, err := ioutil.TempFile("", "finder_")
			if err != nil {
				t.Fatal(err)
			}
			outFileName := outFile.Name()
			_ = os.Remove(outFileName)

			f, err := NewFinder(outFileName, tt.patterns, tt.keys...)
			if err != nil {
				t.Fatal(err)
			}
			err = f.Operate(mailFile)
			if err != nil {
				t.Fatal(err)
			}

			if got, want := f.processed, tt.processed; got != want {
				t.Errorf("processed got %d want %d", got, want)
			}
			if got, want := f.found, tt.found; got != want {
				t.Errorf("found got %d want %d", got, want)
			}
			_ = os.Remove(outFileName)
		})
	}
}
