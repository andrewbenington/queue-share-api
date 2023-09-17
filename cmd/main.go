package main

import (
	"log"
	"os"

	"github.com/andrewbenington/go-spotify/app"
	"github.com/andrewbenington/go-spotify/db"
	"github.com/zmb3/spotify/v2"
)

var (
	client  *spotify.Client
	session string
)

func main() {
	a := app.App{}
	a.Initialize()
	err := db.Service().Initialize()
	if err != nil {
		log.Fatal(err)
	}

	addr := ":8080"
	if len(os.Args) > 1 {
		addr = os.Args[1]
	}
	a.Run(addr)
}
