// MIT License
// Copyright (c) 2020 ysicing <i@ysicing.me>

package main

import (
	"github.com/gin-gonic/gin"
	mid "github.com/ysicing/ginmid"
	"github.com/ysicing/go-lab/ginws/ws"
)

func main()  {

	go ws.WebsocketManager.Start()
	go ws.WebsocketManager.SendService()
	go ws.WebsocketManager.SendService()
	go ws.WebsocketManager.SendGroupService()
	go ws.WebsocketManager.SendGroupService()
	go ws.WebsocketManager.SendAllService()
	go ws.WebsocketManager.SendAllService()
	go ws.TestSendGroup()
	go ws.TestSendAll()

	r := gin.Default()
	r.Use(mid.RequestID(), mid.PromMiddleware(nil))

	wsGroup := r.Group("/ws")
	{
		wsGroup.GET("/:channel", ws.WebsocketManager.WsClient)
	}
	r.Run("0.0.0.0:9001")
}