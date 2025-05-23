package main

import (
	"log"
	"os"
	"time"

	"github.com/andrewbenington/queue-share-api/app"
	"github.com/andrewbenington/queue-share-api/db"
	"github.com/andrewbenington/queue-share-api/engine"
	"github.com/andrewbenington/queue-share-api/version"
	"gopkg.in/yaml.v3"
)

func main() {
	v := version.Get()
	bytes, err := yaml.Marshal(v)
	if err != nil {
		log.Panicf("marshal version data: %s", err)
	}
	log.Println("version:\n" + string(bytes))

	a := app.App{}
	a.Initialize()
	err = db.Service().Initialize()
	if err != nil {
		log.Println(err)
		// log.Fatal(err)
	}

	addr := "0.0.0.0:8080"
	if len(os.Args) > 1 {
		addr = os.Args[1]
	}
	now := time.Now()
	engine.LastFetch = &now
	go engine.Run()

	a.Run(addr)
}
