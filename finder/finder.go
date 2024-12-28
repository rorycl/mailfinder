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
	return searchText(content, searchers)
}

func searchHTML(content string, searchers []*regexp.Regexp) (bool, error) {
	plainText := html2text.HTML2Text(content)
	return searchText(plainText, searchers)
}

// Finder searches the
func Finder(r io.Reader, searchers []*regexp.Regexp) (bool, error) {
	if len(searchers) == 0 {
		return false, errors.New("no regular expressions received")
	}
	email, err := letters.ParseEmail(r)
	if err != nil {
		return false, err
	}
	switch {
	case email.Text != "":
		return searchText(email.Text, searchers)
	case email.EnrichedText != "":
		return searchEnrichedText(email.EnrichedText, searchers)
	case email.HTML != "":
		return searchHTML(email.HTML, searchers)
	}
	return false, nil
}
