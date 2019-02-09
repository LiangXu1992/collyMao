package main

import (
	"collyMao/app/Schedules"
	"collyMao/config"
	"collyMao/orm"
	"log"
	"time"
)

func init() {
	log.SetFlags(log.Lshortfile | log.Ldate + log.Ltime)

}

func main() {
	//读取配置文件
	config.Config.Start("config/config.develop.yml")
	//开启orm
	orm.Start()
	//定时任务
	Schedules.Start()

	for {
		time.Sleep(time.Hour * 24)
	}
}
