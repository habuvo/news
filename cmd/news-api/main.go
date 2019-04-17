package main

import (
	"github.com/gin-gonic/gin"
	"github.com/habuvo/news/cmd/internal"
	"log"
	"net/http"
	"os"
	"time"
)

func main() {

	defer internal.CloseConfig(internal.Global)

	var err error
	internal.Global, err = internal.GetConfig()

	if err != nil {
		log.Fatal(err)
	}

	router := gin.New()

	r := router.Group("/news")
	r.GET("/get/:id", internal.GetNews)

	s := &http.Server{
		Addr:         os.Getenv("API_BINDING"),
		Handler:      router,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}
	log.Fatal(s.ListenAndServe())
}
