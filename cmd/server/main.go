package main

import (
	"github.com/ilnsm/mcollector/internal/server"
	"github.com/ilnsm/mcollector/internal/storage/memory"
	"log"
)

func main() {
	s, err := memorystorage.New()
	if err != nil {
		log.Fatal("could not inizialize storage")
	}
	if err := server.Run(s); err != nil {
		panic(err)
	}
}
