package main

import (
	"context"
	"fmt"
	"os"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "usage: winch <status|stream> [--url URL]")
		os.Exit(2)
	}
	var err error
	switch os.Args[1] {
	case "status":
		err = runStatus(os.Args[2:], os.Stdout)
	case "stream":
		err = runStream(context.Background(), os.Args[2:], os.Stdout)
	default:
		fmt.Fprintln(os.Stderr, "usage: winch <status|stream> [--url URL]")
		os.Exit(2)
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
