# mailfinder 

version 0.0.1 : 01 January 2025

A programme to search concurrently for emails in mbox or maildir format
by (golang) regular expressions, saving the output to a unix mbox.

```
./mailfinder -h

Usage:
  mailfinder [options] OutputMbox

Find email in mbox and maildirs using golang regular expressions.
Note that at least one mbox or maildir must be specified, together with
at least one regular expression.

version 0.0.1

e.g. mailfinder -d maildir -b mbox1 -b mbox2 -r "fire.*safety"  OutputMbox

Application Options:
  -d, --maildir=    path to one or more maildirs
  -b, --mbox=       path to one or more mboxes
  -r, --regexes=    one or more golang regular expressions (required)

Help Options:
  -h, --help        Show this help message

Arguments:
  OutputMbox:       output mbox path (must be unique)

```

## License

This project is licensed under the [MIT Licence](LICENCE).
