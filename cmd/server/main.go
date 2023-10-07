package main

import (
	"github.com/ilnsm/mcollector/internal/server"
	memoryStorage "github.com/ilnsm/mcollector/internal/storage/memory"
	"log"
)

func main() {
	//TODO: parse config
	s, err := memoryStorage.New()
	if err != nil {
		log.Fatal("could not inizialize storage")
	}
	if err := server.Run(s); err != nil {
		panic(err)
	}
}
