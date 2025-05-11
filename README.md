# mailfinder
Search emails in mbox or maildir directories.

version 0.0.13 : 11 May 2025 : simplify invocation of finder

Update to use
[github.com/rorycl/letters](https://github.com/rorycl/letters), which
offers speed improvements, and to search text* content-type inline and
attached files.

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

version 0.0.13

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

  mailfinder --headers -d maildir1 -b mbox2.xz -b mbox3 -m 'Re: Friday' -r "fire.*safety" OutputMbox

Application Options:
  -d, --maildir=     path to maildirs
  -b, --mbox=        path to mboxes
  -r, --regex=       golang regular expressions for search
  -m, --matcher=     string expressions for search
      --from         also search email From header
      --to           also search email To header
      --cc           also search email Cc header
      --subject      also search email Subject header
      --messageid    also search messageid header
  -a, --headers      search email From, To, Cc, Subject and MessageID headers
  -k, --dontskip     don't skip email parsing errors
  -o, --headersonly  don't search bodies
      --datefrom=    inclusive date from which to search (2006-01-02 format)
      --dateto=      inclusive date to which to search (2006-01-02 format)
  -w, --workers=     number of worker goroutines (default: 8)

Help Options:
  -h, --help         Show this help message

Arguments:
  OutputMbox:        output mbox path (must not already exist)

```

## License

This project is licensed under the [MIT Licence](LICENCE).
