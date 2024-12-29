package finder

import (
	"errors"
	"io"
	"regexp"

	"github.com/k3a/html2text"
	"github.com/mnako/letters"
)

func searchText(content string, searchers []*regexp.Regexp) (bool, error) {
	seen := 0
	for _, s := range searchers {
		if s.MatchString(content) {
			seen++
			if seen == len(searchers) {
				return true, nil
			}
			continue
		}
	}
	return false, nil
}

func searchEnrichedText(content string, searchers []*regexp.Regexp) (bool, error) {
	return searchHTML(content, searchers)
}

func searchHTML(content string, searchers []*regexp.Regexp) (bool, error) {
	plainText := html2text.HTML2Text(content)
	return searchText(plainText, searchers)
}

// Finder searches the
func Finder(r io.Reader, searchers []*regexp.Regexp) (bool, *letters.Headers, error) {
	if len(searchers) == 0 {
		return false, nil, errors.New("no regular expressions received")
	}
	email, err := letters.ParseEmail(r)
	if err != nil {
		return false, nil, err
	}
	switch {
	case email.Text != "":
		ok, err := searchText(email.Text, searchers)
		return ok, &email.Headers, err
	case email.EnrichedText != "":
		ok, err := searchEnrichedText(email.EnrichedText, searchers)
		return ok, &email.Headers, err
	case email.HTML != "":
		ok, err := searchEnrichedText(email.HTML, searchers)
		return ok, &email.Headers, err
	}
	return false, nil, nil
}
