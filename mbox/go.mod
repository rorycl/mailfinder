module github.com/rorycl/mailfinder/mbox

go 1.22.5

replace github.com/rorycl/mailfinder/mail => ../mail

require (
	github.com/ProtonMail/go-mbox v1.1.0
	github.com/rorycl/mailfinder/mail v0.0.0-00010101000000-000000000000
)
