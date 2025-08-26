package main

import (
	"log"

	"github.com/vera/vera-drive-service/internal/app"
)

func main() {
	a, err := app.InitApp()
	if err != nil {
		log.Fatal("failed to initialize app", err)
	}
	defer a.Close()

	a.Run()
}
