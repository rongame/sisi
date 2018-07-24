package redis

import (
	"github.com/garyburd/redigo/redis"
	"log"
	"time"
)

type Handler struct {
	Pool *redis.Pool //连接池

	Max_pool_size int //最大连接池数
	Timeout       int //超时时间

	Addr     string //Redis地址
	Pass     string //Redis密码
	Port     string //Redis端口
	DataBase string //Redis数据库
}

func (h *Handler) Init() {
	h.Pool = &redis.Pool{
		MaxIdle:     h.Max_pool_size,
		MaxActive:   h.Max_pool_size,
		IdleTimeout: 180 * time.Second, //h.Timeout * time.Second,
		Dial: func() (redis.Conn, error) {
			c, err := redis.Dial("tcp", h.Addr+":"+h.Port)
			if err != nil {
				return nil, err
			}
			if h.Pass != "" {
				if _, err := c.Do("AUTH", h.Pass); err != nil {
					panic(err)
				}
			}

			if _, err := c.Do("SELECT", h.DataBase); err != nil {
				panic(err)
			}
			return c, nil
		},
	}
}

func (h *Handler) SET(key string, value interface{}) {
	pool := h.Pool.Get()
	defer pool.Close()

	if ok, err := redis.String(pool.Do("SET", key, value)); ok == "OK" {
		log.Println("SET写入成功")
	} else {
		log.Println(err, ok)
	}
}

func (h *Handler) SADD(key string, value interface{}) {
	pool := h.Pool.Get()
	defer pool.Close()

	if ok, err := redis.Int(pool.Do("SADD", key, value)); ok > 0 {
		log.Println("SADD写入成功")
	} else {
		log.Println(err, ok, "SADD")
	}
}

func (h *Handler) ZADD(key string, member interface{}, score int) {
	pool := h.Pool.Get()
	defer pool.Close()

	if ok, err := redis.Int(pool.Do("ZADD", key, score, member)); ok > 0 {
		log.Println("ZADD写入成功")
	} else {
		log.Println(err, ok, "ZADD")
	}
}

func (h *Handler) LPUSH(key string, value interface{}) {
	pool := h.Pool.Get()
	defer pool.Close()

	if ok, err := redis.Int(pool.Do("LPUSH", key, value)); ok > 0 {
		log.Println("LPUSH写入成功")
	} else {
		log.Println(err, ok, "LPUSH")
	}
}
