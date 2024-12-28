module github.com/rorycl/mailfinder

go 1.22.5

replace github.com/rorycl/mailfinder/mail => ./mail

replace github.com/rorycl/mailfinder/mbox => ./mbox

replace github.com/rorycl/mailfinder/maildir => ./maildir

replace github.com/rorycl/mailfinder/finder => ./finder

require (
	github.com/ProtonMail/go-mbox v1.1.0 // indirect
	github.com/k3a/html2text v1.2.1 // indirect
	github.com/mnako/letters v0.2.3 // indirect
	github.com/rorycl/mailfinder/finder v0.0.0-00010101000000-000000000000 // indirect
	github.com/rorycl/mailfinder/mail v0.0.0-00010101000000-000000000000 // indirect
	github.com/rorycl/mailfinder/maildir v0.0.0-00010101000000-000000000000 // indirect
	github.com/rorycl/mailfinder/mbox v0.0.0-00010101000000-000000000000 // indirect
	golang.org/x/net v0.30.0 // indirect
	golang.org/x/text v0.19.0 // indirect
)
