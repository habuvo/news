package internal

import (
	"fmt"
	"github.com/habuvo/news/storage"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	"github.com/nats-io/go-nats"
	"log"
	"os"
	"time"
)

type Config struct {
	DB          *gorm.DB
	NC          *nats.Conn
	NewsStorage *storage.NewsDataStore
}

const tries = 20

var Global *Config

func GetConfig() (con *Config, err error) {

	var c Config
	//dirty hack for db and nats init waiting
	tick := time.NewTicker(time.Second * 1)
	defer tick.Stop()
	tr := 1
	for range tick.C {
		c.NC, err = nats.Connect(nats.DefaultURL)
		if err != nil {
			if tr < tries {
				tr++
				continue
			}
			return nil, err
		}
		break
	}

	tr = 1
	println(fmt.Sprintf("user=%s dbname=%s password=%s",
		os.Getenv("POSTGRES_USER"),
		os.Getenv("POSTGRES_DB"),
		os.Getenv("POSTGRES_PASSWORD")))
	for range tick.C {
		c.DB, err = gorm.Open("postgres",
			fmt.Sprintf("user=%s dbname=%s password=%s",
				os.Getenv("POSTGRES_USER"),
				os.Getenv("POSTGRES_DB"),
				os.Getenv("POSTGRES_PASSWORD")))
		if err != nil {
			if tr < tries {
				tr++
				continue
			}
			return nil, err
		}
		break
	}

	c.NewsStorage = &storage.NewsDataStore{Conn: c.DB}
	return &c, nil
}

func CloseConfig(c *Config) {
	if c == nil || c.NC == nil || c.DB == nil {
		return
	}
	err := c.NC.Drain()
	if err != nil {
		log.Printf("error drain nsq %v", err)
	}
	err = c.DB.Close()
	if err != nil {
		log.Printf("error close db %v", err)
	}
}
