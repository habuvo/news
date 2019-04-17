package internal

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/golang/protobuf/proto"
	"github.com/habuvo/news/storage"
	"github.com/nats-io/go-nats"
	"log"
	"math/rand"
	"net/http"
	"strconv"
	"time"
)

func GetNews(c *gin.Context) {

	idc, ok := c.Params.Get("id")
	if !ok {
		c.JSON(http.StatusBadRequest, "id required")
		return
	}

	id, err := strconv.Atoi(idc)
	if err != nil {
		c.JSON(http.StatusBadRequest, "wrong id")
		return
	}

	jobId := int64(rand.Int())

	data, err := proto.Marshal(&NewsReq{
		JobId:  jobId,
		Action: NewsReq_GET,
		Item: &NewsItem{
			Id: int64(id),
		},
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, fmt.Sprintf("proto marshalling : %v", err))
		return
	}

	err = Global.NC.Publish("news-req", data)
	if err != nil {
		c.JSON(http.StatusInternalServerError, fmt.Sprintf("publish message : %v", err))
		return
	}
	res := make(chan *nats.Msg, 64)
	_, err = Global.NC.ChanSubscribe("news_res", res)
	timer := time.After(time.Second * 10)
	for {
		select {
		case <-timer:
			c.JSON(http.StatusInternalServerError, fmt.Sprintf("break by timeout"))
			return
		case m := <-res:
			message := &NewsRes{}
			err := proto.Unmarshal(m.Data, message)
			if err != nil {
				c.JSON(http.StatusInternalServerError, fmt.Sprintf("error with message marshalling %v", err))
				return
			}
			if message.JobId != jobId {
				continue
			} else {
				if !message.Success {
					c.JSON(http.StatusInternalServerError, fmt.Sprintf("error processing request %v", message.Error))
					return
				} else {
					c.JSON(http.StatusOK, message.Item)
					return
				}
			}
		default:
		}
	}
}

func Worker(cfg *Config, m *nats.Msg) {
	//process message and pub response
	message := &NewsReq{}
	err := proto.Unmarshal(m.Data, message)
	if err != nil {
		log.Printf("unmarshall request message error %v", err)
		return
	}

	item := message.GetItem()
	//maybe check if time is set
	timestamp, err := time.Parse(time.Stamp, item.TimeStamp)
	if err != nil {
		log.Printf("parse timestamp error %v", err)
		return
	}

	switch {
	case message.Action == NewsReq_GET:
		n := storage.NewsItem{
			TimeStamp: timestamp,
			Header:    item.Header,
		}
		n.ID = uint(item.Id)

		err = cfg.NewsStorage.GetByObject(&n)
		resp := NewsRes{
			Success: err == nil,
			Error:   err.Error(),
			JobId:   message.JobId,
		}
		if err != nil {
			resp.Item = append(resp.Item, &NewsItem{
				TimeStamp: n.TimeStamp.Format(time.Stamp),
				Header:    n.Header,
				Id:        int64(n.ID),
			})
		}
		data, err := proto.Marshal(&resp)
		if err != nil {
			log.Printf("marshall response error %v", err)
			return
		}
		err = cfg.NC.Publish("news-res", data)
		if err != nil {
			log.Printf("marshall response error %v", err)
		}
		return
	case message.Action == NewsReq_POST:
		//same like GET
	case message.Action == NewsReq_PUT:
		//same like GET
	default:
		log.Printf("wrong action type %v", message.Action)
		return
	}
}
