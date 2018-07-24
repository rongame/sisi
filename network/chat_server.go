package network

import (
	"encoding/json"
	"github.com/gorilla/websocket"
	"github.com/rongame/sisi/conf"
	"github.com/rongame/sisi/db/maria"
	"github.com/rongame/sisi/db/redis"
	"github.com/rongame/sisi/util"
	"github.com/satori/go.uuid"
	"log"
	"net/http"
	"strconv"
)

type CHATServer struct {
	Addr           string //聊天服务器地址
	SaveMessage    bool   //是否保存聊天记录
	SaveMessageDay int    //保存聊天记录天数

	Redis *redis.Handler //redis数据库句柄
	Maria *maria.Handler //maria数据库句柄

	Parent interface{}
}

var chatServer *CHATServer = new(CHATServer)

func GetChatServer() *CHATServer {
	return chatServer
}

func (c *CHATServer) OnConfig() {
	serverEnv := conf.LoadEnv()
	if c == nil {
		c = chatServer
	}
	c.Addr = serverEnv.ChatAddr

	//初始化Redis数据库
	max_pool_size, _ := strconv.Atoi(serverEnv.RedisMaxPoolSize)
	chatServer.Redis = &redis.Handler{Addr: serverEnv.RedisAddr, Pass: serverEnv.RedisPass, Port: serverEnv.RedisPort,
		DataBase: serverEnv.RedisDatabase, Max_pool_size: max_pool_size}
	chatServer.Redis.Init()
}

func (c *CHATServer) OnStart() {
	if c == nil {
		c = chatServer
	}
	log.Println("初始化聊天服务器")
	log.Println("Starting application...", c)
	go manager.start()
	http.HandleFunc("/chat", chatPage)
	http.ListenAndServe(c.Addr, nil)
}

func (c *CHATServer) OnInit() {
	log.Println("服务器Init方法")
}

func (c *CHATServer) OnDestroy() {
	log.Println("销毁聊天服务器")
}

func (c *CHATServer) UpdateMessage(message []byte) []byte {
	return []byte("heiheiheihei")
}

const ( //聊天数据库表名
	RD_ROOM_CHAT     = "chat:room"      //聊天室id SADD
	RD_ROOM_CHAT_UID = "chat:room:uid:" //+聊天室id 聊天室玩家uid SADD
	RD_ROOM_CHAT_MSG = "chat:room:uid:" //+room_id + : + uid 玩家在聊天室中发的聊天记录 (聊天内容+聊天时间) LPUSH

	RD_PRIVATE_CHAT             = "chat:private"              //聊天用户id SADD
	RD_PRIVATE_CHAT_UID         = "chat:private:uid:"         //+uid 聊天玩家的fuid SADD
	RD_PRIVATE_CHAT_MSG_SEND    = "chat:private:send:uid:"    //+uid + : + fuid uid发送给fuid的私聊记录 (聊天内容+聊天时间) LPUSH
	RD_PRIVATE_CHAT_MSG_RECEIVE = "chat:private:receive:uid:" //+uid + : + fuid uid收到fuid发送的私聊记录 (聊天内容+聊天时间) LPUSH
)

/**
 * 发出一条聊天室中的聊天，保存到redis中
 * @param  {[type]} c *Client)      SendRoomChat(msg string [description]
 * @return {[type]}   [description]
 */
func (c *Client) sendRoomChatToRD(msg string) {
	if c.room_id != "" { //如果玩家当前不在聊天室中，则不发送
		chatServer.Redis.SADD(RD_ROOM_CHAT_UID+c.room_id, c.id)
		chatServer.Redis.LPUSH(RD_ROOM_CHAT_MSG+c.room_id+":"+c.id, msg+":"+util.GetTimestampString())
	}
}

/**
 * 发出一条私聊信息，保存到redis中
 * @param  {[type]} c *Client)      sendPrivateChatToRD(fuid, msg string [description]
 * @return {[type]}   [description]
 */
func (c *Client) sendPrivateChatToRD(fuid, msg string) {
	chatServer.Redis.SADD(RD_PRIVATE_CHAT, c.id)
	chatServer.Redis.SADD(RD_PRIVATE_CHAT_UID+c.id, fuid)
	chatServer.Redis.LPUSH(RD_PRIVATE_CHAT_MSG_SEND+c.id+":"+fuid, msg+":"+util.GetTimestampString())
	chatServer.Redis.LPUSH(RD_PRIVATE_CHAT_MSG_RECEIVE+fuid+":"+c.id, msg+":"+util.GetTimestampString())
}

func chatPage(res http.ResponseWriter, req *http.Request) {
	conn, error := (&websocket.Upgrader{CheckOrigin: func(r *http.Request) bool { return true }}).Upgrade(res, req, nil)
	if error != nil {
		http.NotFound(res, req)
		return
	}
	client_uid, _ := uuid.NewV4()
	client := &Client{id: client_uid.String(), socket: conn, send: make(chan []byte)}

	manager.register <- client

	go client.read()
	go client.write()
}

type ClientManager struct {
	clients        map[*Client]bool
	broadcast      chan []byte
	register       chan *Client
	unregister     chan *Client
	room_broadcast chan map[*Client][]byte
	sendmsg        chan map[*Client][]byte
	privatemsg     chan map[string][]byte
}

type Client struct {
	id      string
	socket  *websocket.Conn
	send    chan []byte
	room_id string
	uid     string
}

type Message struct {
	Sender    string `json:"sender,omitempty"`
	Recipient string `json:"recipient,omitempty"`
	Content   string `json:"content,omitempty"`
}

var manager = ClientManager{
	broadcast:      make(chan []byte),
	register:       make(chan *Client),
	unregister:     make(chan *Client),
	clients:        make(map[*Client]bool),
	room_broadcast: make(chan map[*Client][]byte), //聊天室广播
	sendmsg:        make(chan map[*Client][]byte), //发送给指定用户信息
	privatemsg:     make(chan map[string][]byte),  //私聊发送消息
}

func (manager *ClientManager) start() {
	for {
		select {
		case conn := <-manager.register:
			log.Println("新用户接入")
			manager.clients[conn] = true
			jsonMessage, _ := json.Marshal(&Message{Content: "/A new socket has connected."})
			manager.send(jsonMessage, conn)
		case conn := <-manager.unregister:
			log.Println("用户退出")
			if _, ok := manager.clients[conn]; ok {
				close(conn.send)
				delete(manager.clients, conn)
				jsonMessage, _ := json.Marshal(&Message{Content: "/A socket has disconnected."})
				manager.send(jsonMessage, conn)
			}
		case message := <-manager.broadcast:
			log.Println("开始广播")
			for conn := range manager.clients {
				select {
				case conn.send <- message:
				default:
					close(conn.send)
					delete(manager.clients, conn)
				}
			}
		case room_broadcast := <-manager.room_broadcast:
			log.Println("开始聊天室中广播")
			for conn, message := range room_broadcast {
				log.Println("开始记录数据库")
				if room := GetChatRoom(conn.room_id); room != nil {
					for _, conn := range room.clients {
						select {
						case conn.send <- message:
						default:
							close(conn.send)
							delete(manager.clients, conn)
						}
					}
				}
			}
		case sendmsg := <-manager.sendmsg:
			for conn, message := range sendmsg {
				conn.send <- message
			}
		case privatemsg := <-manager.privatemsg:
			for uid, message := range privatemsg {
				for conn := range manager.clients {
					if conn.uid == uid {
						select {
						case conn.send <- message:
							log.Println("开始记录数据库")
						default:
							close(conn.send)
							delete(manager.clients, conn)
						}
					}
				}
			}
		}
	}
}

func (manager *ClientManager) send(message []byte, ignore *Client) {
	for conn := range manager.clients {
		if conn != ignore {
			conn.send <- message
		}
	}
}

type WebsocketData struct {
	Code    int    `json:"code"`    //事件类型 1.加入聊天室 2.退出聊天室 3.广播 4.私聊
	Id      string `json:"id"`      //事件发起的id
	Content string `json:"content"` //事件内容
}

/**
 * 获取客户端发送的消息
 * @param  {[type]} c *Client)      read( [description]
 * @return {[type]}   [description]
 */
func (c *Client) read() {
	defer func() {
		manager.unregister <- c
		c.socket.Close()
	}()

	for {
		_, message, err := c.socket.ReadMessage()
		if err != nil {
			manager.unregister <- c
			c.socket.Close()
			break
		}
		wsData := new(WebsocketData)
		json.Unmarshal(message, wsData)

		switch wsData.Code {
		case 0: //加入聊天室
			c.uid = wsData.Id
		case 1: //加入房间
			if c.room_id != "" { //如果之前加入过聊天室，则需要退出之前的聊天室
				c.quitChatRoom(c.room_id)
			}

			c.joinChatRoom(wsData.Id)

			jsonMessage, _ := json.Marshal(wsData)
			sendmsg := make(map[*Client][]byte)
			sendmsg[c] = jsonMessage

			manager.sendmsg <- sendmsg
		case 2: //退出房间
			c.quitChatRoom(wsData.Id)
		case 3: //房间中广播
			jsonMessage, _ := json.Marshal(&Message{Sender: c.id, Content: wsData.Content})
			room_broadcast := make(map[*Client][]byte)
			room_broadcast[c] = jsonMessage
			c.sendRoomChatToRD(wsData.Content)

			manager.room_broadcast <- room_broadcast
		case 4: //私聊
			jsonMessage, _ := json.Marshal(&Message{Sender: c.id, Content: wsData.Content})
			privatemsg := make(map[string][]byte)
			privatemsg[wsData.Id] = jsonMessage
			c.sendPrivateChatToRD(c.id, wsData.Content)

			manager.privatemsg <- privatemsg
		}

		// jsonMessage, _ := json.Marshal(&Message{Sender: c.id, Content: string(message)})
		// log.Println("客户端发送的消息：", string(message))
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

			// message = chatServer.UpdateMessage(message)
			c.socket.WriteMessage(websocket.TextMessage, message)
		}
	}
}

type chatRoom struct {
	Name    string    //房间名字
	Id      string    //房间id
	clients []*Client //房间中的用户
}

var rooms []*chatRoom

/**
 * 创建房屋
 * @param  {[type]} c *ClientManager)  CreateRoom(roomName string [description]
 * @return {[type]}   [description]
 */
func (c *ClientManager) CreateRoom(roomName string) {
	if !ExistRoom(roomName) {
		room := new(chatRoom)
		room.Name = roomName
		client_uid, _ := uuid.NewV4()
		room.Id = client_uid.String()

		rooms = append(rooms, room)

		chatServer.Redis.SADD(RD_ROOM_CHAT, room.Id)

		log.Println("创建房间：", room, "所有房间:", rooms)
	}
}

/**
 * 该房间名称是否存在
 */
func ExistRoom(roomName string) bool {
	for _, room := range rooms {
		if room.Name == roomName {
			return true
		}
	}

	return false
}

/**
 * 根据id获取聊天室
 */
func GetChatRoom(room_id string) *chatRoom {
	for _, room := range rooms {
		if room.Id == room_id {
			return room
		}
	}

	return nil
}

/**
 * 加入聊天室
 * @param  {[type]} c *Client)      joinChatRoom(room_id string [description]
 * @return {[type]}   [description]
 */
func (c *Client) joinChatRoom(room_id string) {
	log.Println(c, "加入聊天室", room_id)
	c.room_id = room_id
	for _, room := range rooms {
		if room.Id == room_id {
			room.clients = append(room.clients, c)
		}
	}
}

/**
 * 退出聊天室
 * @param  {[type]} c *Client)      quitChatRoom(room_id string [description]
 * @return {[type]}   [description]
 */
func (c *Client) quitChatRoom(room_id string) {
	c.room_id = ""
	for _, room := range rooms {
		if room.Id == room_id {
			for index, client := range room.clients {
				if client.id == c.id {
					room.clients = append(room.clients[:index], room.clients[index+1:]...)
				}
			}
		}
	}
}
