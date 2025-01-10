# mailfinder 

version 0.0.6 : 10 January 2025

A programme to search for emails in mbox or maildir format by (golang)
regular expressions, saving matched emails to an mbox. Each provided
mbox or maildir mailbox is searched concurrently. Email parsing errors
are optionally skipped.

This uses [mailboxoperator](https://github.com/rorycl/mailboxoperator)
for concurrent parsing of mailboxes. Due to mailboxoperator, searching
mbox files compressed with xz, gzip and bzip2 is supported.

```
Usage:
  mailfinder [options] OutputMbox

Find email in mbox and maildirs using one or more golang regular
expressions. At least one mbox or maildir must be specified. Searches
can optionally be extended to some header fields specified individually
or by using the Headers option.

All regular expressions must match.

(See https://yourbasic.org/golang/regexp-cheat-sheet/ for a primer on
golang's flavour of regular expressions.)

For boolean flags (such as From, To, Headers, etc.) only supply the flag
to include that item. For example, -s or --subject includes searching of
the subject lines of emails.

Mbox format files can also be xz, gz or bz2 compressed. Decompression
should be transparent.

Each mailbox (mbox or maildir) is searched concurrently and pattern
matching and writing done by a number of workers, with the number set by
the -w/--workers switch.

Emails are de-duplicated by message id.

version 0.0.6

e.g. mailfinder --headers -d maildir1 -b mbox2.xz -b mbox3 -r "fire.*safety"  OutputMbox

Application Options:
  -d, --maildir=    path to one or more maildirs
  -b, --mbox=       path to one or more mboxes
  -r, --regexes=    one or more golang regular expressions (required)
  -w, --workers=    number of worker goroutines (default: 8)
  -f, --from        also search email From header
  -t, --to          also search email To header
  -c, --cc          also search email Cc header
  -s, --subject     also search email Subject header
  -h, --headers     search email From, To, Cc and Subject headers
  -k, --dontskip    don't skip email parsing errors

Help Options:
  -h, --help        Show this help message

Arguments:
  OutputMbox:       output mbox path (must not already exist)

```

## License

This project is licensed under the [MIT Licence](LICENCE).
