package main

import (
	"errors"
	"fmt"
	"os"

	mbo "github.com/rorycl/mailboxoperator"
)

var exit func(code int) = os.Exit

func main() {
	finder, err := ParseOptions()
	if err != nil {
		var e ParserError
		if !errors.As(err, &e) {
			fmt.Println(err)
		}
		exit(1)
		return
	}

	// initialise mailbox operator
	mo, err := mbo.NewMailboxOperator(finder.mboxes, finder.maildirs, finder)
	if err != nil {
		fmt.Println(err)
		exit(1)
		return
	}

	// perform the operation on all the emails
	err = mo.Operate()
	if err != nil {
		fmt.Println(err)
		exit(1)
		return
	}

	// print summary
	fmt.Println(finder.Summary())

}
