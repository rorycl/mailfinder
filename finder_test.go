package main

import (
	"fmt"
	"os"
	"regexp"
	"testing"

	"github.com/rorycl/letters"
	"github.com/rorycl/letters/parser"
)

func TestSearchText(t *testing.T) {

	tests := []struct {
		contents  string
		searchers []*regexp.Regexp
		matchers  []string
		ok        bool
	}{
		{
			contents: "abc",
			searchers: []*regexp.Regexp{
				regexp.MustCompile("ab"),
			},
			ok: true,
		},
		{
			contents: "abc\ndef",
			searchers: []*regexp.Regexp{
				regexp.MustCompile(".bc"),
				regexp.MustCompile("d[en]."),
			},
			ok: true,
		},
		{
			contents: "ABC\ndef",
			searchers: []*regexp.Regexp{
				regexp.MustCompile("(?i).bc"),
				regexp.MustCompile("d[en]."),
			},
			ok: true,
		},
		{
			contents: "ABC\ndef",
			searchers: []*regexp.Regexp{
				regexp.MustCompile("(?i).bc"),
				regexp.MustCompile("d[en]."),
			},
			matchers: []string{
				"ABC",
			},
			ok: true,
		},
		{
			contents: "ABC\ndef",
			searchers: []*regexp.Regexp{
				regexp.MustCompile("(?i).bc"),
				regexp.MustCompile("d[en]."),
			},
			matchers: []string{
				"xyz",
			},
			ok: false,
		},
		{
			contents: "abc\ndef",
			searchers: []*regexp.Regexp{
				regexp.MustCompile("ab"),
				regexp.MustCompile("z"),
			},
			ok: false,
		},
		{
			contents: "abc\ndef",
			searchers: []*regexp.Regexp{
				regexp.MustCompile("g"),
			},
			ok: false,
		},
		{
			contents: "abc\ndef",
			matchers: []string{
				"abc",
				"def",
			},
			ok: true,
		},
	}

	for i, tt := range tests {

		t.Run(fmt.Sprintf("test-%d", i), func(t *testing.T) {
			mc, err := newMatchCounter(len(tt.searchers) + len(tt.matchers))
			if err != nil {
				t.Fatal(err)
			}
			f := Finder{searchers: tt.searchers, matchers: tt.matchers}
			got := f.searchText(tt.contents, mc)
			if got, want := got, tt.ok; got != want {
				t.Errorf("got %t want %t", got, want)
			}
		})
	}
}

func TestSearchHTML(t *testing.T) {

	tests := []struct {
		contents  string
		searchers []*regexp.Regexp
		matchers  []string
		ok        bool
		useFunc   func(f *Finder, content string, mc *matchCounter) bool
	}{
		{
			contents: "<h1>hello</h1><p>there</p>",
			searchers: []*regexp.Regexp{
				regexp.MustCompile("hello"),
				regexp.MustCompile("there"),
			},
			ok:      true,
			useFunc: (*Finder).searchHTML,
		},
		{
			contents: "<h1>hello</h1><p>there</p>",
			searchers: []*regexp.Regexp{
				regexp.MustCompile("h1"),
			},
			ok:      false,
			useFunc: (*Finder).searchHTML,
		},
		{
			contents: "<h1>hello</h1><p>there</p>",
			searchers: []*regexp.Regexp{
				regexp.MustCompile("hello"),
				regexp.MustCompile("there"),
			},
			ok:      true,
			useFunc: (*Finder).searchEnrichedText,
		},
		{
			contents: "<h1>hello</h1><p>there</p>",
			searchers: []*regexp.Regexp{
				regexp.MustCompile("hello"),
			},
			matchers: []string{
				"hello",
				"there",
			},
			ok:      true,
			useFunc: (*Finder).searchEnrichedText,
		},
	}

	for i, tt := range tests {
		t.Run(fmt.Sprintf("test-%d", i), func(t *testing.T) {
			mc, err := newMatchCounter(len(tt.searchers) + len(tt.matchers))
			if err != nil {
				t.Fatal(err)
			}
			f := &Finder{searchers: tt.searchers, matchers: tt.matchers}
			got := tt.useFunc(f, tt.contents, mc)
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
		searchers []*regexp.Regexp
		matchers  []string
		keys      []string
		num       int
	}{
		{
			desc:      "to ok",
			emailFile: "testdata/error.eml",
			searchers: []*regexp.Regexp{
				regexp.MustCompile("exampleto"),
			},
			keys: []string{"To"},
			num:  1,
		},
		{
			desc:      "to not ok",
			emailFile: "testdata/sync.eml",
			searchers: []*regexp.Regexp{
				regexp.MustCompile("notExample"),
			},
			keys: []string{"To"},
			num:  0,
		},
		{
			desc:      "to and from ok",
			emailFile: "testdata/error.eml",
			searchers: []*regexp.Regexp{
				regexp.MustCompile("examplefrom"),
				regexp.MustCompile("exampleto"),
			},
			keys: []string{"To", "From"},
			num:  2,
		},
		{
			desc:      "to from and subject ok",
			emailFile: "testdata/error.eml",
			searchers: []*regexp.Regexp{
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
			searchers: []*regexp.Regexp{
				regexp.MustCompile("exampleto"),
				regexp.MustCompile("examplefrom"),
				regexp.MustCompile("^error email not match$"),
			},
			keys: []string{"To", "From", "Subject"},
			num:  2,
		},
		{
			desc:      "to from and subject partially ok search and match",
			emailFile: "testdata/error.eml",
			searchers: []*regexp.Regexp{
				regexp.MustCompile("exampleto"),
				regexp.MustCompile("^error email not match$"),
			},
			matchers: []string{
				"this is an error",
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

		opt := parser.WithHeadersOnly()
		p := letters.NewParser(opt, parser.WithoutAttachments())
		parsedEmail, err := p.Parse(file)
		if err != nil {
			t.Fatal(err)
		}

		t.Run(fmt.Sprintf("test_%s", tt.desc), func(t *testing.T) {
			f := &Finder{searchers: tt.searchers, matchers: tt.matchers, headerKeys: tt.keys}
			mc := f.searchHeaders(parsedEmail.Headers)
			if got, want := mc.got, tt.num; got != want {
				t.Errorf("got %d matches want %d", got, want)
			}
		})
	}
}

func TestFinder(t *testing.T) {

	tests := []struct {
		file             string
		searchers        []*regexp.Regexp
		matchers         []string
		processed, found int
	}{
		{
			file: "testdata/test_txt.eml",
			searchers: []*regexp.Regexp{
				regexp.MustCompile("test.*golang"),
				regexp.MustCompile("(?i)this section"),
			},
			processed: 1,
			found:     1,
		},
		{
			file: "testdata/test_html.eml",
			searchers: []*regexp.Regexp{
				regexp.MustCompile("test.*golang"),
				regexp.MustCompile("(?i)this section"),
			},
			processed: 1,
			found:     1,
		},
		{
			file: "testdata/test_enriched.eml",
			searchers: []*regexp.Regexp{
				regexp.MustCompile("test.*golang"),
				regexp.MustCompile("(?i)this section"),
			},
			processed: 1,
			found:     1,
		},
		{
			file: "testdata/test_enriched.eml",
			searchers: []*regexp.Regexp{
				regexp.MustCompile("(?i)this section"),
			},
			processed: 1,
			found:     1,
		},
		{
			file: "testdata/test_txt.eml",
			searchers: []*regexp.Regexp{
				regexp.MustCompile("This is not a test"),
			},
			processed: 1,
			found:     0,
		},
		{
			file: "testdata/test_multipart.eml",
			searchers: []*regexp.Regexp{
				regexp.MustCompile("1:body:A"), // body text
			},
			processed: 1,
			found:     1,
		},
		{
			file: "testdata/test_multipart.eml",
			searchers: []*regexp.Regexp{
				regexp.MustCompile("<div dir="), // body html in markup
			},
			processed: 1,
			found:     0, // fail
		},
		{
			file: "testdata/test_multipart.eml",
			searchers: []*regexp.Regexp{
				regexp.MustCompile("2:body:B"), // body html after stripping
			},
			processed: 1,
			found:     1,
		},
		{
			file: "testdata/test_multipart.eml",
			searchers: []*regexp.Regexp{
				regexp.MustCompile("1.txt:C"), // text/plain attachment, base64 encoded
			},
			processed: 1,
			found:     1,
		},
		{
			file: "testdata/test_multipart.eml",
			searchers: []*regexp.Regexp{
				regexp.MustCompile("2.html:D"), // text/html attachment, base64 encoded
			},
			processed: 1,
			found:     1,
		},
		{
			file: "testdata/test_multipart.eml",
			searchers: []*regexp.Regexp{
				regexp.MustCompile("1:body:A"), // body text
				regexp.MustCompile("2:body:B"), // body html after stripping
				regexp.MustCompile("1.txt:C"),  // text/plain attachment, base64 encoded
				regexp.MustCompile("2.html:D"), // text/html attachment, base64 encoded
			},
			processed: 1,
			found:     1,
		},
		{
			file: "testdata/test_multipart.eml",
			searchers: []*regexp.Regexp{
				regexp.MustCompile("1:body:A"), // body text
				regexp.MustCompile("2:body:B"), // body html after stripping
				regexp.MustCompile("1.txt:C"),  // text/plain attachment, base64 encoded
				regexp.MustCompile("2.html:D"), // text/html attachment, base64 encoded
			},
			matchers: []string{
				"A multipart test.",
				"1:body:A",
			},
			processed: 1,
			found:     1,
		},
	}

	for i, tt := range tests {
		t.Run(fmt.Sprintf("test-%d", i), func(t *testing.T) {
			mailFile, err := os.Open(tt.file)
			if err != nil {
				t.Fatal(err)
			}
			defer mailFile.Close()

			outFile, err := os.CreateTemp("", "finder_")
			if err != nil {
				t.Fatal(err)
			}
			outFileName := outFile.Name()
			_ = os.Remove(outFileName)

			f, err := NewFinder(outFileName, tt.searchers, tt.matchers)
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
		searchers        []*regexp.Regexp
		matchers         []string
		keys             []string
		processed, found int
	}{
		{
			file: "testdata/test_txt.eml",
			searchers: []*regexp.Regexp{
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
			searchers: []*regexp.Regexp{
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
			searchers: []*regexp.Regexp{
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
			searchers: []*regexp.Regexp{
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
			searchers: []*regexp.Regexp{
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
			searchers: []*regexp.Regexp{
				regexp.MustCompile("test.*golang"),
				regexp.MustCompile("(?i)this section"),
			},
			keys:      []string{},
			processed: 1,
			found:     1,
		},
		{
			file: "testdata/test_txt.eml",
			searchers: []*regexp.Regexp{
				regexp.MustCompile("This is not a test"),
			},
			keys:      []string{},
			processed: 1,
			found:     0,
		},
		{
			file: "testdata/test_txt.eml",
			searchers: []*regexp.Regexp{
				// note use of QuoteMeta to escape "+" character
				regexp.MustCompile(regexp.QuoteMeta(`CAPQX7QTZxwWh31YxJQd+DcLCm0qTRxCErYwAYRnd-FiFk=hdrQ@mail.gmail.com`)),
			},
			keys:      []string{"MessageID"},
			processed: 1,
			found:     1,
		},
		{
			file: "testdata/test_txt.eml",
			searchers: []*regexp.Regexp{
				// note use of QuoteMeta to escape "+" character
				regexp.MustCompile(regexp.QuoteMeta(`CAPQX7QTZxwWh31YxJQd+DcLCm0qTRxCErYwAYRnd-FiFk=hdrQ@mail.gmail.com`)),
			},
			matchers: []string{
				// redundant but useful for testing
				"CAPQX7QTZxwWh31YxJQd+DcLCm0qTRxCErYwAYRnd-FiFk=hdrQ@mail.gmail.com",
			},
			keys:      []string{"MessageID"},
			processed: 1,
			found:     1,
		},
	}

	for i, tt := range tests {
		t.Run(fmt.Sprintf("test-%d", i), func(t *testing.T) {
			mailFile, err := os.Open(tt.file)
			if err != nil {
				t.Fatal(err)
			}
			defer mailFile.Close()

			outFile, err := os.CreateTemp("", "finder_")
			if err != nil {
				t.Fatal(err)
			}
			outFileName := outFile.Name()
			_ = os.Remove(outFileName)

			f, err := NewFinder(outFileName, tt.searchers, tt.matchers, tt.keys...)
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

func TestSummary(t *testing.T) {
	tests := []struct {
		file             string
		processed, found int // for info in this case
	}{
		{
			file:      "testdata/test_txt.eml",
			processed: 1,
			found:     1,
		},
		{
			file:      "./testdata/error.eml",
			processed: 1,
			found:     0,
		},
	}
	outFile, err := os.CreateTemp("", "finder_")
	if err != nil {
		t.Fatal(err)
	}
	outFileName := outFile.Name()
	_ = os.Remove(outFileName)

	searchers := []*regexp.Regexp{
		regexp.MustCompile("example"),
		regexp.MustCompile("(?i)this section"),
	}
	keys := []string{"To"}

	f, err := NewFinder(outFileName, searchers, []string{}, keys...)
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(outFileName)

	for i, tt := range tests {
		t.Run(fmt.Sprintf("test-%d", i), func(t *testing.T) {
			mailFile, err := os.Open(tt.file)
			if err != nil {
				t.Fatal(err)
			}
			defer mailFile.Close()

			err = f.Operate(mailFile)
			if err != nil {
				t.Fatal(err)
			}

		})
	}
	if got, want := f.Summary(), "processed 2 found 1 emails"; got != want {
		t.Errorf("got\n%s\nwant\n%s", got, want)
	}
}
