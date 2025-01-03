package main

import (
	"log"

	"github.com/cugu/fomo/app"
)

func main() {
	if err := app.Run(); err != nil {
		log.Fatalf("fomo failed: %v", err)
	}
}
