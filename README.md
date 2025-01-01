# mailfinder 

version 0.0.2 : 01 January 2025

A programme to search concurrently for emails in mbox or maildir format
by (golang) regular expressions, saving the output to a unix mbox.

```
./mailfinder -h

Usage:
  mailfinder [options] OutputMbox

Find email in mbox and maildirs using golang regular expressions. At
least one mbox or maildir must be specified, together with at least one
regular expression. Searches can optionally be extended to some header
fields specified individually or by using the Headers option.

All regular expressions must match.

version 0.0.2

e.g. mailfinder --headers -d maildir -b mbox1 -b mbox2 -r "fire.*safety"  OutputMbox

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
