package main

import (
	"log"

	"insurance/db"
	"insurance/router"
)

func main() {
	if err := db.Init("insurance.db"); err != nil {
		log.Fatal(err)
	}

	e := router.New()
	e.Logger.Fatal(e.Start(":8080"))
}
