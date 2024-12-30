package cmd

import (
	"errors"
	"fmt"
	"os"
	"regexp"

	"github.com/jessevdk/go-flags"
)

const version string = "0.0.1"

// Options are flag options
type Options struct {
	Maildirs []string `short:"d" long:"maildir" description:"path to one or more maildirs"`
	Mboxes   []string `short:"b" long:"mbox" description:"path to one or more mboxes"`
	Regexes  []string `short:"r" long:"regexes" description:"one or more golang regular expressions to search for"`
	regexes  []*regexp.Regexp
	Args     struct {
		OutputMbox string `description:"output mbox path (must be unique)"`
	} `positional-args:"yes" required:"yes"`
}

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

// parseOptions parses the command line options
func parseOptions() (*Options, error) {

	var options Options
	var parser = flags.NewParser(&options, flags.Default)
	parser.Usage = version

	if _, err := parser.Parse(); err != nil {
		return nil, err
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
