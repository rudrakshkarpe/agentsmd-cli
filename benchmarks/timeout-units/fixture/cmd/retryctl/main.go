package main

import (
	"fmt"
	"os"

	"example.com/retryctl/internal/retry"
)

func main() {
	delay, err := retry.Delay()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	fmt.Println(delay)
}
