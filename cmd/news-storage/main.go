package main

import (
	"github.com/habuvo/news/cmd/internal"
	"github.com/habuvo/news/storage"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	"github.com/nats-io/go-nats"
	"log"
	"os"
	"os/signal"
	"syscall"
)

func main() {

	cfg, err := internal.GetConfig()

	if err != nil {
		log.Fatal(err)
	}

	cfg.DB.AutoMigrate(storage.NewsItem{})

	//listen and serve

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	req := make(chan *nats.Msg, 64)
	_, err = cfg.NC.ChanSubscribe("news_req", req)
	if err != nil {
		log.Fatalf("can't subscribe to news_req", err)
	}
	for {
		select {
		case <-stop:
			err = cfg.NC.Drain()
			if err != nil {
				log.Printf("drain channel error : %v", err)
			}
			err = cfg.DB.Close()
			if err != nil {
				log.Printf("close db error : %v", err)
			}
			os.Exit(0)
		case msg := <-req:
			go func(m *nats.Msg) {
				//process message and pub response
			}(msg)
		}
	}
}
