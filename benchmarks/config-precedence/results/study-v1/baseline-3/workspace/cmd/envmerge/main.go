package main

import (
	"fmt"
	"os"

	"example.com/envmerge/internal/config"
)

func main() {
	loaded, err := config.Load("config.json", os.LookupEnv)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	fmt.Println(loaded.Region)
}
