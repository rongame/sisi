package network

import (
	// "encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/rongame/sisi/module"
	"github.com/satori/go.uuid"
	"net/http"
	"time"
)

type ClientManager struct {
	clients    map[*Client]bool
	broadcast  chan []byte
	register   chan *Client
	unregister chan *Client
	sendmsg    chan map[*Client][]byte
}

type Client struct {
	id        string
	uid       string
	socket    *websocket.Conn
	send      chan []byte
	stay_time int //在线等待时间
}

type Message struct {
	Sender    string `json:"sender,omitempty"`
	Recipient string `json:"recipient,omitempty"`
	Content   string `json:"content,omitempty"`
}

var manager = ClientManager{
	broadcast:  make(chan []byte),
	register:   make(chan *Client),
	unregister: make(chan *Client),
	clients:    make(map[*Client]bool),
	sendmsg:    make(chan map[*Client][]byte),
}

// var user_conn chan int = make(chan int)

func (manager *ClientManager) start() {
	fmt.Println("进入star")
	fmt.Println("进入star成功")

	for {
		select {
		case conn := <-manager.register:
			manager.clients[conn] = true
		case conn := <-manager.unregister:
			if _, ok := manager.clients[conn]; ok {
				fmt.Println("用户下线-----------------")

				close(conn.send)
				delete(manager.clients, conn)
			}
		case message := <-manager.broadcast:
			for conn := range manager.clients {
				select {
				case conn.send <- message:
				default:
					close(conn.send)
					delete(manager.clients, conn)
				}
			}
		case sendmsg := <-manager.sendmsg:
			for conn, message := range sendmsg {
				if conn != nil {
					conn.send <- message
				}
			}
		}
	}
}

/**
 * 服务器轮询任务 **后期需要改成携程池**
 * @param  {[type]} manager *ClientManager) runTask( [description]
 * @return {[type]}         [description]
 */
func (manager *ClientManager) runTask() {
	for {
		fmt.Println("runTask进行中")
		time.Sleep(time.Second * 1)
	}
}

func (manager *ClientManager) send(message []byte, ignore *Client) {
	for conn := range manager.clients {
		if conn != ignore {
			conn.send <- message
		}
	}
}

func (c *Client) read() {
	defer func() {
		manager.unregister <- c
		c.socket.Close()
	}()

	for {
		_, message, err := c.socket.ReadMessage()
		fmt.Println(message, err)

		// 广播发送
		// jsonMessage, _ := json.Marshal(&Message{Sender: c.id, Content: string(message)})
		// manager.broadcast <- jsonMessage
	}
}

func (c *Client) write() {
	defer func() {
		c.socket.Close()
	}()

	for {
		select {
		case message, ok := <-c.send:
			if !ok {
				c.socket.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			c.socket.WriteMessage(websocket.TextMessage, message)
		}
	}
}

func wshandler(res http.ResponseWriter, req *http.Request) {
	conn, error := (&websocket.Upgrader{CheckOrigin: func(r *http.Request) bool { return true }}).Upgrade(res, req, nil)
	if error != nil {
		http.NotFound(res, req)
		return
	}
	client_uid, _ := uuid.NewV4()
	client := &Client{id: client_uid.String(), uid: "", socket: conn, send: make(chan []byte)}

	fmt.Println("连接的客户端为：", client)

	manager.register <- client

	go client.read()
	go client.write()
}

func (w *WebSocketServer) InitWebsocketRouter() {
	go manager.start()
	go manager.runTask()
	fmt.Println("jinrusocket")
	var interface1 module.Module = w
	go interface1.OnLogic(manager)
	w.Router.GET("/ws", func(c *gin.Context) {
		wshandler(c.Writer, c.Request)
	})
}
