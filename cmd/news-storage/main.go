package main

import (
	"log"
	"github.com/nats-io/go-nats"
)

func main() {

	//connect to NATS
	nc, err := nats.Connect(nats.DefaultURL)
	if err != nil {
		log.Fatal(err)
	}
	defer nc.Close()
}
