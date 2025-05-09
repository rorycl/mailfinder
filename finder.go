package main

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"net/mail"
	"regexp"
	"strings"
	"sync"

	"github.com/k3a/html2text"
	"github.com/rorycl/letters"
	"github.com/rorycl/letters/email"
	"github.com/rorycl/letters/parser"
	"github.com/rorycl/mailboxoperator/mbox"
)

// matchCounter reports if an email matches the search criteria provided
// (regexes and matchers) by checking that there is at least one count
// for each  counts the numer of hits for a regexp or matcher. This
// should not be used in a concurrent context.
type matchCounter struct {
	matchMap map[string]int
	need     int
	got      int
}

// newMatchCounter returns a new *matchCounter or error if the needed
// number of matches is less than 1. "need" should equal the combined
// length of finder.matchers and finder.searchers.
func newMatchCounter(need int) (*matchCounter, error) {
	if need < 1 {
		return nil, fmt.Errorf("matchCounter initialised with need %d", need)
	}
	return &matchCounter{
		matchMap: map[string]int{},
		need:     need,
	}, nil
}

// found reports if this email has met all matchers and searchers
func (m *matchCounter) found() bool {
	return m.got == m.need
}

// search searches the provided content with the finder.matchers string
// expressions and finder.searchers regular expressions. If each
// match/search criteria is met, return true, else false.
//
// Consider in future not using regular expressions whose String()
// exactly matches any string matchers since the latter is faster.
func (m *matchCounter) search(content string, f *Finder) bool {
	if m.need == 0 {
		panic("matchCounter not initialised before use")
	}
	if m.found() {
		return true
	}

	// search string expressions
	for _, str := range f.matchers {
		if _, ok := m.matchMap[str]; ok {
			continue
		}
		if strings.Contains(content, str) {
			m.matchMap[str]++
			m.got++
			if m.found() {
				return true
			}
		}
	}

	// search regular expressions.
	// Note that regular expressions are recorded in the matchMap with
	// an "r#" prefix to try avoid inadvertent overlaps with string
	// matcher keys.
	for _, r := range f.searchers {
		if _, ok := m.matchMap["r#"+r.String()]; ok {
			continue
		}
		if r.MatchString(content) {
			m.matchMap["r#"+r.String()]++
			m.got++
			if m.found() {
				return true
			}
		}
	}
	return m.found()
}

// searchText searches only for text content
func (f *Finder) searchText(content string, mc *matchCounter) bool {
	return mc.search(content, f)
}

// searchEnrichedText redirects to searchHTML
func (f *Finder) searchEnrichedText(content string, mc *matchCounter) bool {
	return f.searchHTML(content, mc)
}

// searchHTML converts html to text and then searches via searchText
func (f *Finder) searchHTML(content string, mc *matchCounter) bool {
	plainText := html2text.HTML2Text(content)
	return f.searchText(plainText, mc)
}

// searchHeaders counts matches against searchers amongst the supplied
// header strings to search
func (f *Finder) searchHeaders(headers email.Headers) *matchCounter {

	mc, err := newMatchCounter(len(f.searchers) + len(f.matchers))
	if err != nil {
		panic(fmt.Sprintf("invalid initialisation of NewMatchCounter %s", err))
	}

	if len(f.headerKeys) == 0 {
		return mc
	}

	findFromAddresses := func(addresses ...*mail.Address) {
		for _, a := range addresses {
			if mc.search(a.Name, f) {
				return
			}
			if mc.search(a.Address, f) {
				return
			}
		}
	}

	findFromString := func(str string) {
		mc.search(str, f)
	}

	for _, k := range f.headerKeys {
		switch k {
		case "Sender":
			findFromAddresses(headers.Sender)
		case "From":
			findFromAddresses(headers.From...)
		case "ReplyTo":
			findFromAddresses(headers.ReplyTo...)
		case "To":
			findFromAddresses(headers.To...)
		case "Cc":
			findFromAddresses(headers.Cc...)
		case "Subject":
			findFromString(headers.Subject)
		case "MessageID":
			findFromString(headers.MessageID)
		}
	}
	return mc
}

// Finder is a struct with settings for performing mail finding
type Finder struct {
	searchers         []*regexp.Regexp
	matchers          []string
	headerKeys        []string
	mboxWriter        *mbox.MboxWriter
	headersOnly       bool
	skipParsingErrors bool
	processed         int
	found             int
	foundMutex        sync.Mutex
	emailParser       *parser.Parser
}

// addFound records processing numbers
func (f *Finder) addFound(b bool) {
	f.foundMutex.Lock()
	defer f.foundMutex.Unlock()
	f.processed++
	if b {
		f.found++
	}
}

// Summary prints a summary of the found emails
func (f *Finder) Summary() string {
	return fmt.Sprintf("processed %d found %d emails", f.processed, f.found)
}

// NewFinder creates a new Finder.
func NewFinder(po *ProgramOptions) (*Finder, error) {
	if (len(po.regexes) + len(po.matchers)) == 0 {
		return nil, errors.New("no regexps or matchers provided")
	}
	if po.headersOnly && len(po.headers) == 0 {
		return nil, errors.New("no headers provided for headersOnly search")
	}

	mbw, err := mbox.NewMboxWriter(po.outputMbox)
	if err != nil {
		return nil, fmt.Errorf("NewFinder error: %w", err)
	}

	f := Finder{
		searchers:         po.regexes,
		matchers:          po.matchers,
		headersOnly:       po.headersOnly,
		headerKeys:        po.headers,
		mboxWriter:        mbw,
		skipParsingErrors: po.skipParsingErrors,
	}
	if f.headersOnly {
		f.emailParser = letters.NewParser(
			parser.WithHeadersOnly(),
		)
	} else {
		// only consider attachments with a "text" content disposition
		onlyProcessTextAttachments := func(fe *email.File) error {
			if !strings.HasPrefix(fe.ContentInfo.Type, "text/") {
				return nil
			}
			var err error
			fe.Data, err = io.ReadAll(fe.Reader)
			return err
		}
		f.emailParser = letters.NewParser(
			parser.WithCustomFileFunc(
				onlyProcessTextAttachments,
			),
		)
	}
	return &f, nil
}

// Finder searches parts of an email for the given regexp, including (where
// provided) searching the provided email headers set out in headerKeys.
// Operate fulfills the "Operator" interface required by
// mailboxoperator.
func (f *Finder) Operate(r io.Reader) error {

	buf := &bytes.Buffer{}
	tee := io.TeeReader(r, buf)

	email, err := f.emailParser.Parse(tee)
	if err != nil {
		if !f.skipParsingErrors {
			return err
		}
		fmt.Println(err)
		return nil
	}

	// search headers
	mc := f.searchHeaders(email.Headers)
	ok := mc.found()

	if f.headersOnly {
		// drain tee
		_, err = io.ReadAll(tee)
		if err != nil {
			return fmt.Errorf("tee draining error %w", err)
		}
	}

	// search content
	if email.Text != "" && !ok {
		ok = f.searchText(email.Text, mc)
	}
	if email.EnrichedText != "" && !ok {
		ok = f.searchEnrichedText(email.EnrichedText, mc)
	}
	if email.HTML != "" && !ok {
		ok = f.searchHTML(email.HTML, mc)
	}
	if len(email.Files) > 0 && !ok {
		for _, fi := range email.Files {
			switch fi.ContentInfo.Type {
			case "text/plain":
				ok = f.searchText(string(fi.Data), mc)
				if ok {
					break
				}
			case "text/enriched":
				ok = f.searchEnrichedText(string(fi.Data), mc)
				if ok {
					break
				}
			case "text/html":
				ok = f.searchHTML(string(fi.Data), mc)
				if ok {
					break
				}
			}
		}
	}
	if !ok {
		f.addFound(false)
		return nil
	}
	f.addFound(true)
	_, err = f.mboxWriter.Add(
		email.Headers.From[0].Address,
		email.Headers.Date,
		string(email.Headers.MessageID),
		buf,
	)
	return err
}
