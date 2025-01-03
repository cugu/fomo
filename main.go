package main

import (
	"fmt"
	"os"

	"github.com/cugu/fomo/app"
)

func main() {
	if err := app.Run(); err != nil {
		fmt.Println(err) //nolint:forbidigo
		os.Exit(1)
	}
}
