# mailfinder 

version 0.0.3 : 02 January 2025

A programme to search concurrently for emails in mbox or maildir format
by (golang) regular expressions, saving the output to a unix mbox.

```
./mailfinder -h

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

version 0.0.3

e.g. mailfinder --headers -d maildir1 -b mbox2.xz -b mbox3 -r "fire.*safety"  OutputMbox

Application Options:
  -d, --maildir=    path to one or more maildirs
  -b, --mbox=       path to one or more mboxes
  -r, --regexes=    one or more golang regular expressions (required)
  -f, --from        also search email From header
  -t, --to          also search email To header
  -c, --cc          also search email Cc header
  -s, --subject     also search email Subject header
  -h, --headers     search email From, To, Cc and Subject headers

Help Options:
  -h, --help        Show this help message

Arguments:
  OutputMbox:       output mbox path (must be unique)

```

## License

This project is licensed under the [MIT Licence](LICENCE).
