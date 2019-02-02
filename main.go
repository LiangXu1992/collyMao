package main

import (
	"collyMao/app/Schedules"
	"collyMao/config"
	"collyMao/orm"
	"log"
	"net/http"
	"os"
)

func init() {
	log.SetFlags(log.Lshortfile | log.Ldate + log.Ltime)
	var logFile, err = os.OpenFile("a.log", os.O_CREATE|os.O_RDWR, 0777)
	if err != nil {
		panic(err.Error())
	}
	log.SetOutput(logFile)
}

func main() {
	//读取配置文件
	config.Config.Start("config/config.develop.yml")
	//开启orm
	orm.Start()
	//定时任务
	Schedules.Start()

	http.ListenAndServe(":8888", nil)
}
