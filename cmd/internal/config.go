package internal

import (
	"fmt"
	"github.com/habuvo/news/storage"
	"github.com/jinzhu/gorm"
	"github.com/nats-io/go-nats"
	"log"
	"os"
)

type Config struct {
	DB          *gorm.DB
	NC          *nats.Conn
	NewsStorage *storage.NewsDataStore
}

var Global *Config

func GetConfig() (c *Config, err error) {

	c.DB, err = gorm.Open("postgres",
		fmt.Sprintf("user=%s dbname=%s password=%s",
			os.Getenv("POSTGRES_USER"),
			os.Getenv("POSTGRES_DB"),
			os.Getenv("POSTGRES_PASSWORD")))
	if err != nil {
		return
	}

	c.NC, err = nats.Connect(nats.DefaultURL)
	if err != nil {
		return
	}

	c.NewsStorage = &storage.NewsDataStore{Conn: c.DB}
	return
}

func CloseConfig(c *Config) {
	err := c.NC.Drain()
	if err != nil {
		log.Printf("error drain nsq %v", err)
	}
	err = c.DB.Close()
	if err != nil {
		log.Printf("error close db %v", err)
	}
}
