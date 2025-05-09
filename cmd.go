package main

import (
	"errors"
	"fmt"
	"os"
	"regexp"
	"runtime"

	"github.com/jessevdk/go-flags"
)

const version string = "0.0.11"

// Options are flag options
type Options struct {
	Maildirs    []string `short:"d" long:"maildir" description:"path to maildirs"`
	Mboxes      []string `short:"b" long:"mbox" description:"path to mboxes"`
	Regexes     []string `short:"r" long:"regex" description:"golang regular expressions for search"`
	Matchers    []string `short:"m" long:"matcher" description:"string expressions for search"`
	From        bool     `short:"f" long:"from" description:"also search email From header"`
	To          bool     `short:"t" long:"to" description:"also search email To header"`
	Cc          bool     `short:"c" long:"cc" description:"also search email Cc header"`
	Subject     bool     `short:"s" long:"subject" description:"also search email Subject header"`
	MessageID   bool     `short:"i" long:"messageid" description:"also search messageid header"`
	Headers     bool     `short:"a" long:"headers" description:"search email From, To, Cc, Subject and MessageID headers"`
	DontSkip    bool     `short:"k" long:"dontskip" description:"don't skip email parsing errors"`
	HeadersOnly bool     `short:"o" long:"headersonly" description:"don't search bodies"`
	Workers     int      `short:"w" long:"workers" description:"number of worker goroutines" default:"8"`
	// internal fields
	headers           []string         // rationalised headers to search
	regexes           []*regexp.Regexp // compiled search terms
	skipParsingErrors bool             // skip email parsing errors
	// output
	Args struct {
		OutputMbox string `description:"output mbox path (must not already exist)"`
	} `positional-args:"yes" required:"yes"`
}

// aggregateHeaders aggregates header options into options.headers
func (o *Options) aggregateHeaders() {
	a := map[string]bool{}
	if o.From {
		a["From"] = true
	}
	if o.To {
		a["To"] = true
	}
	if o.Cc {
		a["Cc"] = true
	}
	if o.Subject {
		a["Subject"] = true
	}
	if o.MessageID {
		a["MessageID"] = true
	}
	if o.Headers {
		a = map[string]bool{
			"From":      true,
			"To":        true,
			"Cc":        true,
			"Subject":   true,
			"MessageID": true,
		}
	}
	v := []string{}
	for k := range a {
		v = append(v, k)
	}
	o.headers = v
}

var cmdTpl string = `[options] OutputMbox

version %s

Find email in mbox and maildirs using one or more golang regular
expressions and/or string matchers. At least one mbox or maildir must be
specified. Searches can optionally be extended to some header fields
specified individually or by using the Headers option.

All regular expressions and string matchers provided must match.

(See https://yourbasic.org/golang/regexp-cheat-sheet/ for a primer on
golang's flavour of regular expressions.)

For boolean flags (such as From, To, Headers, etc.) only supply the flag
to include that item. For example, -s or --subject includes searching of
the subject lines of emails.

Mbox format files can also be xz, gz or bz2 compressed. Decompression
is transparent.

Each mailbox (mbox or maildir) is searched concurrently and searching
and output mailbox writing done by a number of workers, with the number
set by the -w/--workers switch.

Emails are de-duplicated by message id.

e.g. 

  mailfinder --headers -d maildir1 -b mbox2.xz -b mbox3 -r "fire.*safety" OutputMbox

or, to search by both regular expression and strings

  mailfinder --headers -d maildir1 -b mbox2.xz -b mbox3 -m 'Re: Friday' -r "fire.*safety"`

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
	if len(options.Regexes) == 0 && len(options.Matchers) == 0 {
		return nil, errors.New("no regular expressions or string matchers provided")
	}
	for i, r := range options.Regexes {
		rr, err := regexp.Compile(r)
		if err != nil {
			return nil, fmt.Errorf("regular expression %d did not compile: %s", i, err)
		}
		options.regexes = append(options.regexes, rr)
	}
	for _, r := range options.Matchers {
		if len(r) < 5 {
			return nil, fmt.Errorf("matcher %s is less than 5 characters in length", r)
		}
	}
	if options.Workers < 1 {
		return nil, errors.New("at least 1 worker is needed to process work")
	}
	if got, want := options.Workers, runtime.NumCPU()*4; got > want {
		return nil, fmt.Errorf("it is inadvisable to have workers of more than four times system cpus (%d)", runtime.NumCPU())
	}
	options.skipParsingErrors = !options.DontSkip
	if options.Args.OutputMbox == "" {
		return nil, errors.New("no output mbox path provided")
	}
	if checkFileExists(options.Args.OutputMbox) {
		return nil, fmt.Errorf("output mbox %s already exists", options.Args.OutputMbox)
	}

	// aggregate the headers
	options.aggregateHeaders()

	if options.HeadersOnly && len(options.headers) == 0 {
		return nil, errors.New("to use headersonly a header option must also be selected")
	}

	return &options, nil
}
