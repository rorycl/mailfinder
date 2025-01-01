package finder

import (
	"errors"
	"fmt"
	"io"
	"net/mail"
	"regexp"

	"github.com/k3a/html2text"
	"github.com/mnako/letters"
)

// matchRegexpCount counts the numer of hits for a regexp
type matchRegexpCount map[*regexp.Regexp]int

// searchText searches only for text content
func searchText(content string, searchers []*regexp.Regexp, matchMap matchRegexpCount) (bool, error) {

	for _, s := range searchers {
		if _, ok := matchMap[s]; ok {
			continue
		}
		if s.MatchString(content) {
			matchMap[s]++
			if len(searchers) == len(matchMap) {
				break
			}
		}
	}
	return (len(matchMap) == len(searchers)), nil
}

// searchEnrichedText redirects to searchHTML
func searchEnrichedText(content string, searchers []*regexp.Regexp, matchMap matchRegexpCount) (bool, error) {
	return searchHTML(content, searchers, matchMap)
}

// searchHTML converts html to text and then searches via searchText
func searchHTML(content string, searchers []*regexp.Regexp, matchMap matchRegexpCount) (bool, error) {
	plainText := html2text.HTML2Text(content)
	return searchText(plainText, searchers, matchMap)
}

// searchHeaders counts matches against searchers amongst the supplied
// header strings to search
func searchHeaders(headers letters.Headers, searchers []*regexp.Regexp, keys ...string) matchRegexpCount {
	matchMap := matchRegexpCount{}
	if len(keys) == 0 {
		return matchMap
	}

	findFromAddresses := func(addresses ...*mail.Address) {
		for _, a := range addresses {
			for _, s := range searchers {
				if len(searchers) == len(matchMap) {
					return
				}
				if _, ok := matchMap[s]; ok {
					continue
				}
				fmt.Println(a.Name, a.Address, s)
				if s.MatchString(a.Name) || s.MatchString(a.Address) {
					matchMap[s]++
				}
			}
		}
	}

	findFromString := func(str string) {
		for _, s := range searchers {
			if len(searchers) == len(matchMap) {
				return
			}
			if _, ok := matchMap[s]; ok {
				continue
			}
			if s.MatchString(str) {
				matchMap[s]++
			}
		}
	}

	for _, k := range keys {
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
		}
	}
	return matchMap
}

// Finder searches parts of an email for the given regexp, including (where
// provided) searching the provided email headers set out in headerKeys.
func Finder(r io.Reader, searchers []*regexp.Regexp, headerKeys ...string) (bool, *letters.Headers, error) {
	if len(searchers) == 0 {
		return false, nil, errors.New("no regular expressions received")
	}
	email, err := letters.ParseEmail(r)
	if err != nil {
		return false, nil, err
	}

	// search headers
	matchMap := searchHeaders(email.Headers, searchers, headerKeys...)

	// search content
	switch {
	case email.Text != "":
		ok, err := searchText(email.Text, searchers, matchMap)
		return ok, &email.Headers, err
	case email.EnrichedText != "":
		ok, err := searchEnrichedText(email.EnrichedText, searchers, matchMap)
		return ok, &email.Headers, err
	case email.HTML != "":
		ok, err := searchEnrichedText(email.HTML, searchers, matchMap)
		return ok, &email.Headers, err
	}
	return false, nil, nil
}
