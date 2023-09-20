package main

import (
	"log"
	"os"

	"github.com/andrewbenington/queue-share-api/app"
	"github.com/andrewbenington/queue-share-api/db"
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
		log.Fatal(err)
	}

	addr := ":8080"
	if len(os.Args) > 1 {
		addr = os.Args[1]
	}
	a.Run(addr)
}
