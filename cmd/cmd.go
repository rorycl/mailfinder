package cmd

import (
	"errors"
	"fmt"
	"os"
	"regexp"

	"github.com/jessevdk/go-flags"
)

const version string = "0.0.2"

// Options are flag options
type Options struct {
	Maildirs []string         `short:"d" long:"maildir" description:"path to one or more maildirs"`
	Mboxes   []string         `short:"b" long:"mbox" description:"path to one or more mboxes"`
	Regexes  []string         `short:"r" long:"regexes" description:"one or more golang regular expressions (required)"`
	From     bool             `short:"f" long:"from" description:"also search email From header"`
	To       bool             `short:"t" long:"to" description:"also search email To header"`
	Cc       bool             `short:"c" long:"cc" description:"also search email Cc header"`
	Subject  bool             `short:"s" long:"subject" description:"also search email Subject header"`
	Headers  bool             `short:"h" long:"headers" description:"search email From, To, Cc and Subject headers"`
	headers  []string         // rationalised headers to search
	regexes  []*regexp.Regexp // compiled search terms
	Args     struct {
		OutputMbox string `description:"output mbox path (must be unique)"`
	} `positional-args:"yes" required:"yes"`
}

var cmdTpl string = `[options] OutputMbox

Find email in mbox and maildirs using golang regular expressions. At
least one mbox or maildir must be specified, together with at least one
regular expression. Searches can optionally be extended to some header
fields specified individually or by using the Headers option.

All regular expressions must match.

version %s

e.g. mailfinder --headers -d maildir -b mbox1 -b mbox2 -r "fire.*safety" `

// checkFileExists checks if a file exists
func checkFileExists(path string) bool {
	p, err := os.Stat(path)
	if err != nil && errors.Is(err, os.ErrNotExist) {
		return false
	}
	if p.IsDir() {
		return false
	}
	return true
}

// checkDirExists checks if a directory exists
func checkDirExists(path string) bool {
	p, err := os.Stat(path)
	if err != nil && errors.Is(err, os.ErrNotExist) {
		return false
	}
	if !p.IsDir() {
		return false
	}
	return true
}

// ParserError indicates a parser error
type ParserError struct {
	err error
}

func (p ParserError) Error() string {
	return fmt.Sprintf("%v", p.err)
}

// aggregateHeaders aggregates header options into options.headers
func aggregateHeader(options *Options) {
	a := map[string]bool{}
	if options.From {
		a["From"] = true
	}
	if options.To {
		a["To"] = true
	}
	if options.Cc {
		a["Cc"] = true
	}
	if options.Subject {
		a["Subject"] = true
	}
	if options.Headers {
		a = map[string]bool{
			"From":    true,
			"To":      true,
			"Cc":      true,
			"Subject": true,
		}
	}
	v := []string{}
	for k := range a {
		v = append(v, k)
	}
	options.headers = v
}

// ParseOptions parses the command line options
func ParseOptions() (*Options, error) {

	var options Options
	var parser = flags.NewParser(&options, flags.Default)
	parser.Usage = fmt.Sprintf(cmdTpl, version)

	if _, err := parser.Parse(); err != nil {
		return nil, ParserError{err}
	}
	if (len(options.Maildirs) + len(options.Mboxes)) == 0 {
		return nil, errors.New("no maildirs or mboxes found")
	}
	for _, d := range options.Maildirs {
		if !checkDirExists(d) {
			return nil, fmt.Errorf("maildir %s does not exist", d)
		}
	}
	for _, m := range options.Mboxes {
		if !checkFileExists(m) {
			return nil, fmt.Errorf("mbox %s does not exist", m)
		}
	}
	if len(options.Regexes) == 0 {
		return nil, errors.New("no regular expressions provided")
	}
	for i, r := range options.Regexes {
		rr, err := regexp.Compile(r)
		if err != nil {
			return nil, fmt.Errorf("regular expression %d did not compile: %s", i, err)
		}
		options.regexes = append(options.regexes, rr)
	}
	if options.Args.OutputMbox == "" {
		return nil, errors.New("no output mbox path provided")
	}
	if checkFileExists(options.Args.OutputMbox) {
		return nil, fmt.Errorf("output mbox %s already exists", options.Args.OutputMbox)
	}
	return &options, nil
}
