package main

import (
	"log"
	"os"
	"strings"
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

	shouldRunEngine := true

	for _, arg := range os.Args {
		if arg == "--no-engine" {
			shouldRunEngine = false
			continue
		}

		if specifiedAddr, ok := strings.CutPrefix(arg, "--addr="); ok {
			addr = specifiedAddr
		}
	}

	if shouldRunEngine {
		now := time.Now()
		engine.LastFetch = &now

		go engine.Run()

	}

	a.Run(addr)
}
