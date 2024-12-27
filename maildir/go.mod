module github.com/rorycl/mailfinder/maildir

go 1.22.5

replace github.com/rorycl/mailfinder/mail => ../mail

require (
	github.com/google/go-cmp v0.6.0
	github.com/rorycl/mailfinder/mail v0.0.0-00010101000000-000000000000
)
