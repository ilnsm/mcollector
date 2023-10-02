package main

import "github.com/ilnsm/mcollector/internal/server"

func main() {
	//TODO: parse config
	//TODO: connect to DB
	if err := server.Run(); err != nil {
		panic(err)
	}
}
