// MIT License
// Copyright (c) 2020 ysicing <i@ysicing.me>

package ws

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	mid "github.com/ysicing/ginmid"
	"github.com/ysicing/go-utils/exid"
	"github.com/ysicing/go-utils/extime"
	"k8s.io/klog"
	"log"
	"net/http"
	"sync"
	"time"
)

// Manager 所有 websocket 信息
type Manager struct {
	Group map[string]map[string]*Client
	groupCount, clientCount uint
	Lock sync.Mutex
	Register, UnRegister chan *Client
	Message	chan *MessageData
	GroupMessage chan *GroupMessageData
	BroadCastMessage chan *BroadCastMessageData
}

// Client 单个 websocket 信息
type Client struct {
	Id, Group string
	Socket *websocket.Conn
	Message   chan []byte
}

// messageData 单个发送数据信息
type MessageData struct {
	Id, Group string
	Message    []byte
}

// groupMessageData 组广播数据信息
type GroupMessageData struct {
	Group string
	Message    []byte
}

// 广播发送数据信息
type BroadCastMessageData struct {
	Message    []byte
}

// 读信息，从 websocket 连接直接读取数据
func (c *Client) Read() {
	defer func() {
		WebsocketManager.UnRegister <- c
		klog.Infof("client [%s] disconnect", c.Id)
		if err := c.Socket.Close(); err != nil {
			klog.Errorf("client [%s] disconnect err: %s", c.Id, err)
		}
	}()

	for {
		messageType, message, err := c.Socket.ReadMessage()
		if err != nil || messageType == websocket.CloseMessage {
			break
		}
		klog.Infof("client [%s] receive message: %s", c.Id, string(message))
		c.Message <- message
	}
}

// 写信息，从 channel 变量 Send 中读取数据写入 websocket 连接
func (c *Client) Write() {
	defer func() {
		klog.Infof("client [%s] disconnect", c.Id)
		if err := c.Socket.Close(); err != nil {
			klog.Errorf("client [%s] disconnect err: %s", c.Id, err)
		}
	}()

	for {
		select {
		case message, ok := <-c.Message:
			if !ok {
				_ = c.Socket.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			klog.Infof("client [%s] write message: %s", c.Id, string(message))
			err := c.Socket.WriteMessage(websocket.BinaryMessage, message)
			if err != nil {
				klog.Errorf("client [%s] writemessage err: %s", c.Id, err)
			}
		}
	}
}

// 启动 websocket 管理器
func (manager *Manager) Start() {
	log.Printf("websocket manage start")
	for {
		select {
		// 注册
		case client := <-manager.Register:
			klog.Infof("client [%s] connect", client.Id)
			klog.Infof("register client [%s] to group [%s]", client.Id, client.Group)

			manager.Lock.Lock()
			if manager.Group[client.Group] == nil {
				manager.Group[client.Group] = make(map[string]*Client)
				manager.groupCount += 1
			}
			manager.Group[client.Group][client.Id] = client
			manager.clientCount += 1
			manager.Lock.Unlock()

		// 注销
		case client := <-manager.UnRegister:
			klog.Infof("unregister client [%s] from group [%s]", client.Id, client.Group)
			manager.Lock.Lock()
			if _, ok := manager.Group[client.Group]; ok {
				if _, ok := manager.Group[client.Group][client.Id]; ok {
					close(client.Message)
					delete(manager.Group[client.Group], client.Id)
					manager.clientCount -= 1
					if len(manager.Group[client.Group]) == 0 {
						//log.Printf("delete empty group [%s]", client.Group)
						delete(manager.Group, client.Group)
						manager.groupCount -= 1
					}
				}
			}
			manager.Lock.Unlock()

			// 发送广播数据到某个组的 channel 变量 Send 中
			//case data := <-manager.boardCast:
			//	if groupMap, ok := manager.wsGroup[data.GroupId]; ok {
			//		for _, conn := range groupMap {
			//			conn.Send <- data.Data
			//		}
			//	}
		}
	}
}

// 处理单个 client 发送数据
func (manager *Manager) SendService() {
	for {
		select {
		case data := <-manager.Message:
			if groupMap, ok := manager.Group[data.Group]; ok {
				if conn, ok := groupMap[data.Id]; ok {
					conn.Message <- data.Message
				}
			}
		}
	}
}

// 处理 group 广播数据
func (manager *Manager) SendGroupService() {
	for {
		select {
		// 发送广播数据到某个组的 channel 变量 Send 中
		case data := <-manager.GroupMessage:
			if groupMap, ok := manager.Group[data.Group]; ok {
				for _, conn := range groupMap {
					conn.Message <- data.Message
				}
			}
		}
	}
}

// 处理广播数据
func (manager *Manager) SendAllService() {
	for {
		select {
		case data := <-manager.BroadCastMessage:
			for _, v := range manager.Group {
				for _, conn := range v {
					conn.Message <- data.Message
				}
			}
		}
	}
}

// 向指定的 client 发送数据
func (manager *Manager) Send(id string, group string, message []byte) {
	data := &MessageData{
		Id: id,
		Group: group,
		Message:    message,
	}
	manager.Message <- data
}

// 向指定的 Group 广播
func (manager *Manager) SendGroup(group string, message []byte) {
	data := &GroupMessageData{
		Group: group,
		Message:    message,
	}
	manager.GroupMessage <- data
}

// 广播
func (manager *Manager) SendAll(message []byte) {
	data := &BroadCastMessageData{
		Message:    message,
	}
	manager.BroadCastMessage <- data
}

// 注册
func (manager *Manager) RegisterClient(client *Client) {
	manager.Register <- client
}

// 注销
func (manager *Manager) UnRegisterClient(client *Client) {
	manager.UnRegister <- client
}

// 当前组个数
func (manager *Manager) LenGroup() uint {
	return manager.groupCount
}

// 当前连接个数
func (manager *Manager) LenClient() uint {
	return manager.clientCount
}

// 获取 wsManager 管理器信息
func (manager *Manager) Info() map[string]interface{} {
	managerInfo := make(map[string]interface{})
	managerInfo["groupLen"] = manager.LenGroup()
	managerInfo["clientLen"] = manager.LenClient()
	managerInfo["chanRegisterLen"] = len(manager.Register)
	managerInfo["chanUnregisterLen"] = len(manager.UnRegister)
	managerInfo["chanMessageLen"] = len(manager.Message)
	managerInfo["chanGroupMessageLen"] = len(manager.GroupMessage)
	managerInfo["chanBroadCastMessageLen"] = len(manager.BroadCastMessage)
	return managerInfo
}

// 初始化 wsManager 管理器
var WebsocketManager = Manager{
	Group: make(map[string]map[string]*Client),
	Register:    make(chan *Client, 128),
	UnRegister:  make(chan *Client, 128),
	GroupMessage:   make(chan *GroupMessageData, 128),
	Message:   make(chan *MessageData, 128),
	BroadCastMessage: make(chan *BroadCastMessageData, 128),
	groupCount: 0,
	clientCount: 0,
}

// gin 处理 websocket handler
func (manager *Manager) WsClient(ctx *gin.Context) {
	upGrader := websocket.Upgrader{
		// cross origin domain
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
		// 处理 Sec-WebSocket-Protocol Header
		//Subprotocols: []string{ctx.GetHeader("Sec-WebSocket-Protocol")},
	}

	conn, err := upGrader.Upgrade(ctx.Writer, ctx.Request, nil)
	if err != nil {
		klog.Infof("websocket connect error: %s", ctx.Param("channel"))
		return
	}

	client := &Client{
		Id:     exid.GenUUID(),
		Group:  ctx.Param("channel"),
		Socket: conn,
		Message:   make(chan []byte, 1024),
	}

	manager.RegisterClient(client)
	go client.Read()
	go client.Write()
	time.Sleep(time.Second * 15)
	// 测试单个 client 发送数据
	manager.Send(client.Id, client.Group, []byte("Send message ----" + extime.GetToday() + "--- requestid " + mid.GetRequestID(ctx)))
}

// 测试组广播
func TestSendGroup() {
	for {
		time.Sleep(time.Second * 20)
		WebsocketManager.SendGroup("leffss", []byte("SendGroup message ----" + extime.GetToday()))
	}
}

// 测试广播
func TestSendAll() {
	for {
		time.Sleep(time.Second * 25)
		WebsocketManager.SendAll([]byte("SendAll message ----" + extime.GetToday()))
		fmt.Println(WebsocketManager.Info())
	}
}