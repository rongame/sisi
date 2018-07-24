package conf

import (
	"bufio"
	"io"
	"os"
	"strings"
)

const middle = ":"

type Env struct {
	Mymap  map[string]string
	strcet string
}

/**
 * 配置文件结构体，服务器配置用
 */
type ServerEnv struct {
	ChatAddr         string //聊天服务器端口
	RedisAddr        string //Redis地址
	RedisPort        string //Redis端口
	RedisPass        string //Redis密码
	RedisDatabase    string //Redis连接数据库
	RedisMaxPoolSize string //数据库最大线程池
	DatabaseName     string //Maria数据库账号名
	DatabasePass     string //Maria数据库密码
	DatabaseAddr     string //Maria数据库地址
	DatabasePort     string //Maria数据库端口
	Database         string //Maria数据库名
	Debug            string //是否开启调试模式
}

func (c *Env) InitConfig(path string) {
	c.Mymap = make(map[string]string)

	f, err := os.Open(path)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	r := bufio.NewReader(f)
	for {
		b, _, err := r.ReadLine()
		if err != nil {
			if err == io.EOF {
				break
			}
			panic(err)
		}

		s := strings.TrimSpace(string(b))
		if strings.Index(s, "#") == 0 {
			continue
		}

		n1 := strings.Index(s, "[")
		n2 := strings.LastIndex(s, "]")
		if n1 > -1 && n2 > -1 && n2 > n1+1 {
			c.strcet = strings.TrimSpace(s[n1+1 : n2])
			continue
		}

		if len(c.strcet) == 0 {
			continue
		}
		index := strings.Index(s, "=")
		if index < 0 {
			continue
		}

		frist := strings.TrimSpace(s[:index])
		if len(frist) == 0 {
			continue
		}
		second := strings.TrimSpace(s[index+1:])

		pos := strings.Index(second, "\t#")
		if pos > -1 {
			second = second[0:pos]
		}

		pos = strings.Index(second, " #")
		if pos > -1 {
			second = second[0:pos]
		}

		pos = strings.Index(second, "\t//")
		if pos > -1 {
			second = second[0:pos]
		}

		pos = strings.Index(second, " //")
		if pos > -1 {
			second = second[0:pos]
		}

		if len(second) == 0 {
			continue
		}

		key := c.strcet + middle + frist
		c.Mymap[key] = strings.TrimSpace(second)
	}
}

func (c Env) Read(node, key string) string {
	key = node + middle + key
	v, found := c.Mymap[key]
	if !found {
		return ""
	}
	return v
}

func LoadEnv(filePath ...string) *ServerEnv {
	myEnv := new(Env)
	if len(filePath) > 0 {
		myEnv.InitConfig(filePath[0])
	} else {
		myEnv.InitConfig("./.env")
	}

	serverEnv := &ServerEnv{
		Database:         myEnv.Read("mysql", "database"),
		DatabaseAddr:     myEnv.Read("mysql", "addr"),
		DatabaseName:     myEnv.Read("mysql", "name"),
		DatabasePass:     myEnv.Read("mysql", "pass"),
		DatabasePort:     myEnv.Read("mysql", "port"),
		ChatAddr:         myEnv.Read("chat", "addr"),
		RedisAddr:        myEnv.Read("redis", "addr"),
		RedisDatabase:    myEnv.Read("redis", "database"),
		RedisPass:        myEnv.Read("redis", "pass"),
		RedisPort:        myEnv.Read("redis", "port"),
		Debug:            myEnv.Read("default", "debug"),
		RedisMaxPoolSize: myEnv.Read("redis", "max_pool_size"),
	}

	return serverEnv
}
