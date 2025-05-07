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

// matchRegexpCount counts the numer of hits for a regexp
type matchRegexpCount map[*regexp.Regexp]int

// searchText searches only for text content
func (f *Finder) searchText(content string, matchMap matchRegexpCount) bool {
	for _, s := range f.searchers {
		if _, ok := matchMap[s]; ok {
			continue
		}
		if s.MatchString(content) {
			matchMap[s]++
			if len(f.searchers) == len(matchMap) {
				break
			}
		}
	}
	return len(matchMap) == len(f.searchers)
}

// searchEnrichedText redirects to searchHTML
func (f *Finder) searchEnrichedText(content string, matchMap matchRegexpCount) bool {
	return f.searchHTML(content, matchMap)
}

// searchHTML converts html to text and then searches via searchText
func (f *Finder) searchHTML(content string, matchMap matchRegexpCount) bool {
	plainText := html2text.HTML2Text(content)
	return f.searchText(plainText, matchMap)
}

// searchHeaders counts matches against searchers amongst the supplied
// header strings to search
func (f *Finder) searchHeaders(headers email.Headers) matchRegexpCount {
	matchMap := matchRegexpCount{}
	if len(f.headerKeys) == 0 {
		return matchMap
	}

	findFromAddresses := func(addresses ...*mail.Address) {
		for _, a := range addresses {
			for _, s := range f.searchers {
				if len(f.searchers) == len(matchMap) {
					return
				}
				if _, ok := matchMap[s]; ok { // continue if already a match
					continue
				}
				if s.MatchString(a.Name) || s.MatchString(a.Address) {
					matchMap[s]++
				}
			}
		}
	}

	findFromString := func(str string) {
		for _, s := range f.searchers {
			if len(f.searchers) == len(matchMap) {
				return
			}
			if _, ok := matchMap[s]; ok { // continue if already a match
				continue
			}
			if s.MatchString(str) {
				matchMap[s]++
			}
		}
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
	return matchMap
}

// Finder is a struct with settings for performing mail finding
type Finder struct {
	searchers         []*regexp.Regexp
	headerKeys        []string
	mboxWriter        *mbox.MboxWriter
	skipParsingErrors bool
	processed         int
	found             int
	foundMutex        sync.Mutex
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
func NewFinder(outputMbox string, searchers []*regexp.Regexp, headerKeys ...string) (*Finder, error) {
	if searchers == nil {
		return nil, errors.New("no regexps provided")
	}
	mbw, err := mbox.NewMboxWriter(outputMbox)
	if err != nil {
		return nil, fmt.Errorf("NewFinder error: %w", err)
	}
	f := Finder{
		searchers:  searchers,
		headerKeys: headerKeys,
		mboxWriter: mbw,
	}
	return &f, nil
}

// Finder searches parts of an email for the given regexp, including (where
// provided) searching the provided email headers set out in headerKeys.
func (f *Finder) Operate(r io.Reader) error {

	buf := &bytes.Buffer{}
	tee := io.TeeReader(r, buf)

	// only consider attachments with a "text" content disposition
	onlyProcessTextAttachments := func(fe *email.File) error {
		if !strings.HasPrefix(fe.ContentInfo.Type, "text/") {
			return nil
		}
		var err error
		fe.Data, err = io.ReadAll(fe.Reader)
		return err
	}

	emailParser := letters.NewParser(
		parser.WithCustomFileFunc(
			onlyProcessTextAttachments,
		),
	)
	email, err := emailParser.Parse(tee)
	if err != nil {
		if !f.skipParsingErrors {
			return err
		}
		fmt.Println(err)
		return nil
	}

	// search headers
	matchMap := f.searchHeaders(email.Headers)

	// search content
	var ok bool = false
	if email.Text != "" {
		ok = f.searchText(email.Text, matchMap)
	}
	if email.EnrichedText != "" && !ok {
		ok = f.searchEnrichedText(email.EnrichedText, matchMap)
	}
	if email.HTML != "" && !ok {
		ok = f.searchHTML(email.HTML, matchMap)
	}
	if len(email.Files) > 0 && !ok {
		for _, fi := range email.Files {
			switch fi.ContentInfo.Type {
			case "text/plain":
				ok = f.searchText(string(fi.Data), matchMap)
			case "text/enriched":
				ok = f.searchEnrichedText(string(fi.Data), matchMap)
			case "text/html":
				ok = f.searchHTML(string(fi.Data), matchMap)
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
