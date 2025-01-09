package main

import (
	"errors"
	"fmt"
	"os"

	"github.com/rorycl/mailfinder/cmd"
)

func main() {
	options, err := cmd.ParseOptions()
	if err != nil {
		var e cmd.ParserError
		if !errors.As(err, &e) {
			fmt.Println(err)
		}
		os.Exit(1)
	}
	err = cmd.Process(options)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

}
