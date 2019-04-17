package internal

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/golang/protobuf/proto"
	"github.com/nats-io/go-nats"
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
