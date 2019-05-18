package orm

import (
	"collyMao/config"
	"database/sql"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"github.com/gomodule/redigo/redis"
	"github.com/jinzhu/gorm"
	"log"
)

var (
	DbClient  *sql.DB
	Gorm      *gorm.DB
	RedisPool *redis.Pool
	err       error
)

var c = config.Config

func Start() {
	//mysql连接池
	a := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s", c.Mysql.User, c.Mysql.Password, c.Mysql.Host, c.Mysql.Port, c.Mysql.Database)
	Gorm, err = gorm.Open("mysql", a)
	if err != nil {
		log.Println("connect mysql err:" + err.Error())
		return
	}
	DbClient = Gorm.DB()
	//Gorm.LogMode(c.Mysql.Logdebug)
	DbClient.SetMaxOpenConns(c.Mysql.MaxActive) //用于设置最大打开的连接数，默认值为0表示不限制。
	DbClient.SetMaxIdleConns(c.Mysql.MaxIdle)   //最大空闲数
	DbClient.Ping()

	/*
		//redis连接池
		RedisPool = &redis.Pool{
			MaxIdle:     c.Redis.MaxIdle,
			MaxActive:   c.Redis.MaxActive,
			IdleTimeout: c.Redis.Idletimeout,
			Dial: func() (redis.Conn, error) {
				redisClient, err := redis.Dial("tcp", fmt.Sprintf("%s:%s", c.Redis.Host, c.Redis.Port))
				if err != nil {
					log.Println("connect redis err:" + err.Error())
					return nil, err
				}
				if c.Redis.Password != "" {
					if _, err := redisClient.Do("AUTH", c.Redis.Password); err != nil {
						log.Println("connect redis password err:" + err.Error())
						redisClient.Close()
						return nil, err
					}
				}
				return redisClient, nil
			},
		}
	*/
}
