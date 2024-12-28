# mailfinder 

version 0.0.0 : 26 December 2024

A programme to search for emails in mbox or maildir format by (golang)
regular expressions, saving the output to a unix mbox.

```
./maildirfinder
	-m 1.mbox -m 1.mbox
    -d 2.maildir/ -d 3.maildir/
    -s "(?i)(regular|expression)"
    -o regular_expression.mbox
```

## Process

	for each mbox or maildir:
		initialise

	go consumers
		initialise mail readers (num)

	go producers
		for each mbox/maildir iterator:
			process mail



	







## License

This project is licensed under the [MIT Licence](LICENCE).
