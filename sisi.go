package sisi

import (
	"github.com/rongame/sisi/module"
	"log"
	"os"
	"os/signal"
)

func Run(mods ...module.Module) {
	log.Println("服务器启动")

	// module
	for i := 0; i < len(mods); i++ {
		module.Register(mods[i])
	}
	module.Init()

	// close
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, os.Kill)
	sig := <-c
	log.Println("Sisi 服务器已经关闭：", sig)
	//下面是关闭服务器时要处理的代码
	module.Destory()
}
