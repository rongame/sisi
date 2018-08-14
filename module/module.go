package module

import (
	"log"
)

type Module interface {
	OnConfig()              //系统启动方法
	OnStart()               //系统启动方法
	OnInit()                //可重写启动方法
	OnLogic(it interface{}) //可重写逻辑方法
	OnDestroy()
}

type module struct {
	mi       Module
	closeSig chan bool
}

var mods []*module

func Register(mi Module) {
	log.Println("模块注册")
	m := new(module)
	m.mi = mi

	mods = append(mods, m)
}

func Init() {
	log.Println("模块初始化")

	for _, mod := range mods {
		mod.mi.OnConfig()
		mod.mi.OnInit()
		mod.mi.OnStart()
	}
}

func Destory() {
	for _, mod := range mods {
		mod.mi.OnDestroy()
	}
}
