package finder

import (
	"fmt"
	"os"
	"regexp"
	"testing"
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
		t.Run(fmt.Sprintf("test-%d", i), func(t *testing.T) {
			got, err := searchText(tt.contents, tt.patterns)
			if err != nil {
				t.Fatal(err)
			}
			if got, want := got, tt.ok; got != want {
				fmt.Errorf("got %t want %t", got, want)
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
		t.Run(fmt.Sprintf("test-%d", i), func(t *testing.T) {
			got, err := searchHTML(tt.contents, tt.patterns)
			if err != nil {
				t.Fatal(err)
			}
			if got, want := got, tt.ok; got != want {
				fmt.Errorf("got %t want %t", got, want)
			}
		})
	}
}

func TestFinder(t *testing.T) {

	f, err := os.Open("testdata/test.eml")
	if err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		patterns []*regexp.Regexp
		ok       bool
	}{
		{
			patterns: []*regexp.Regexp{
				regexp.MustCompile("test.*golang"),
				regexp.MustCompile("(?i)this section"),
			},
			ok: true,
		},
		{
			patterns: []*regexp.Regexp{
				regexp.MustCompile("This is not a test"),
			},
			ok: false,
		},
	}

	for i, tt := range tests {
		t.Run(fmt.Sprintf("test-%d", i), func(t *testing.T) {
			got, _, err := Finder(f, tt.patterns)
			if err != nil {
				t.Fatal(err)
			}
			if got, want := got, tt.ok; got != want {
				fmt.Errorf("got %t want %t", got, want)
			}
		})
		_, err = f.Seek(0, 0)
		if err != nil {
			t.Fatal(err)
		}
	}

}
