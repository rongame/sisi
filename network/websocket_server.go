package network

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/rongame/sisi/conf"
	"github.com/rongame/sisi/db/maria"
	"github.com/rongame/sisi/db/redis"
	"strconv"
)

type WebSocketServer struct {
	Addr string //服务器地址

	Router *gin.Engine

	Redis *redis.Handler //redis数据库句柄
	Maria *maria.Handler //maria数据库句柄
}

var webSocketServer *WebSocketServer = new(WebSocketServer)

func (w *WebSocketServer) OnConfig() {
	fmt.Println("onConfig")
	serverEnv := conf.LoadEnv()
	if w == nil {
		// w = webSocketServer
	}
	w.Addr = serverEnv.ChatAddr

	//初始化Redis数据库
	max_pool_size, _ := strconv.Atoi(serverEnv.RedisMaxPoolSize)
	webSocketServer.Redis = &redis.Handler{Addr: serverEnv.RedisAddr, Pass: serverEnv.RedisPass, Port: serverEnv.RedisPort,
		DataBase: serverEnv.RedisDatabase, Max_pool_size: max_pool_size}
	webSocketServer.Redis.Init()
	webSocketServer.Router = gin.Default()
}

func (w *WebSocketServer) OnStart() {
	if w == nil {
		// w = webSocketServer
	}
	webSocketServer.Router.Use(Cors())

	w.InitWebsocketRouter()

	webSocketServer.Router.Run(webSocketServer.Addr)
}

func (w *WebSocketServer) OnInit() {
	fmt.Println("onInit")
}

func (w *WebSocketServer) OnDestroy() {
	fmt.Println("onDestroy")
}

func (w *WebSocketServer) OnLogic(it interface{}) {
	fmt.Println("onLogic")
}

/**
 * 解决http跨域
 */
func Cors() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Origin, X-Requested-With, Content-Type, Accept, Access-Token, Access-Sign")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT,DELETE,OPTIONS")
		c.Next()
	}
}
