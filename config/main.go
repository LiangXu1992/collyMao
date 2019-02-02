package config

import (
	"github.com/jinzhu/configor"
	"time"
)

var Config = &config{}

type config struct {
	AppName                          string
	GO_CORE_URL                      string
	New_order_sleep_milli_second     time.Duration
	History_order_sleep_milli_second time.Duration
	Bind_phone_url                   string

	Mysql struct {
		Host      string
		User      string
		Password  string
		Port      string
		Database  string
		MaxActive int
		MaxIdle   int
		Logdebug  bool
	}

	Redis struct {
		Host        string
		Password    string
		Port        string
		MaxActive   int
		MaxIdle     int
		Idletimeout time.Duration
	}
}

func (c *config) Start(fileName string) {
	err := configor.Load(c, fileName)
	if err != nil {
		panic(err)
	}
}
