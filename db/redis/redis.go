package redis

import (
	"github.com/garyburd/redigo/redis"
	"log"
	"reflect"
	"strconv"
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

func (h *Handler) HMSET(key, field string, value interface{}) {
	pool := h.Pool.Get()
	defer pool.Close()

	if ok, err := pool.Do("HMSET", key, field, value); err == nil {
		log.Println("HMSET写入成功")
	} else {
		log.Println(err, ok, "HMSET")
	}
}

func (h *Handler) ZINCRBY(key, member string, value interface{}) {
	pool := h.Pool.Get()
	defer pool.Close()

	if ok, err := pool.Do("ZINCRBY", key, value, member); err == nil {
		log.Println("ZINCRBY写入成功")
	} else {
		log.Println(err, ok, "ZINCRBY")
	}
}

func (h *Handler) ZREM(key, member string) {
	pool := h.Pool.Get()
	defer pool.Close()

	if ok, err := pool.Do("ZREM", key, member); err == nil {
		log.Println("ZREM成功")
	} else {
		log.Println(err, ok, "ZREM")
	}
}

func (h *Handler) SAVE(tableName string, obj interface{}) error {
	mutable := reflect.ValueOf(obj).Elem()
	typeOfType := mutable.Type()
	for i := 0; i < mutable.NumField(); i++ {
		f := typeOfType.Field(i).Name

		h.HMSET(tableName, f, mutable.Field(i))
	}

	return nil
}

func (h *Handler) SMEMBERS(key string) ([]string, error) {
	pool := h.Pool.Get()
	defer pool.Close()

	if ok, err := redis.Strings(pool.Do("SMEMBERS", key)); err == nil {
		return ok, nil
	} else {
		return nil, err
	}
}

func (h *Handler) GET(key string) (string, error) {
	pool := h.Pool.Get()
	defer pool.Close()

	if ok, err := redis.String(pool.Do("GET", key)); err == nil {
		return ok, nil
	} else {
		return "", err
	}
}

func (h *Handler) HGET(key, field string) (string, error) {
	pool := h.Pool.Get()
	defer pool.Close()

	if ok, err := redis.String(pool.Do("HGET", key, field)); err == nil {
		return ok, nil
	} else {
		return "", err
	}
}

func (h *Handler) HGETInt(key, field string) (int, error) {
	pool := h.Pool.Get()
	defer pool.Close()

	if ok, err := redis.Int(pool.Do("HGET", key, field)); err == nil {
		return ok, nil
	} else {
		return 0, err
	}
}

func (h *Handler) HGETBool(key, field string) (bool, error) {
	pool := h.Pool.Get()
	defer pool.Close()

	if ok, err := redis.Bool(pool.Do("HGET", key, field)); err == nil {
		return ok, nil
	} else {
		return false, err
	}
}

func (h *Handler) HGETALL(k string, item interface{}) (map[string]string, error) {
	pool := h.Pool.Get()
	defer pool.Close()

	if ok, err := redis.StringMap(pool.Do("HGETALL", k)); err == nil {
		mutable := reflect.ValueOf(item).Elem()
		for key, value := range ok {
			switch mutable.FieldByName(key).Kind() {
			case reflect.Int:
				tmpValue, _ := strconv.ParseInt(value, 10, 64)
				mutable.FieldByName(key).SetInt(tmpValue)
			case reflect.String:
				mutable.FieldByName(key).SetString(value)
			case reflect.Bool:
				tmpValue, _ := strconv.ParseBool(value)
				mutable.FieldByName(key).SetBool(tmpValue)
			case reflect.Float64:
				tmpValue, _ := strconv.ParseFloat(value, 64)
				mutable.FieldByName(key).SetFloat(tmpValue)
			}
		}

		return ok, nil
	} else {
		log.Println(err, "------------", ok)
		return nil, err
	}
}

func (h *Handler) ZSCORE(key, field string) (int, error) {
	pool := h.Pool.Get()
	defer pool.Close()

	if ok, err := redis.Int(pool.Do("ZSCORE", key, field)); err == nil {
		return ok, nil
	} else {
		return 0, err
	}
}

func (h *Handler) ZCOUNT(key string) (int, error) {
	pool := h.Pool.Get()
	defer pool.Close()

	if ok, err := redis.Int(pool.Do("ZCOUNT", key, 0, -1)); err == nil {
		return ok, nil
	} else {
		return 0, err
	}
}

func (h *Handler) ZRANGE(key string) ([]string, error) {
	pool := h.Pool.Get()
	defer pool.Close()

	if ok, err := redis.Strings(pool.Do("ZRANGE", key, 0, -1)); err == nil {
		return ok, nil
	} else {
		return nil, err
	}
}

func (h *Handler) SCARD(key string) (int, error) {
	pool := h.Pool.Get()
	defer pool.Close()

	if ok, err := redis.Int(pool.Do("SCARD", key)); err == nil {
		return ok, nil
	} else {
		return 0, err
	}
}

func (h *Handler) DEL(key string) error {
	pool := h.Pool.Get()
	defer pool.Close()

	if _, err := redis.Bool(pool.Do("DEL", key)); err == nil {
		return nil
	} else {
		return err
	}
}
