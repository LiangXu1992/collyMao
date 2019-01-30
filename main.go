package main

import (
	"github.com/gocolly/colly"
	"log"
	"os"
	"strconv"
	"strings"
)

var gameCount int64

func init() {
	log.SetFlags(log.Lshortfile | log.Ldate + log.Ltime)
	var logFile, err = os.OpenFile("a.log", os.O_CREATE|os.O_RDWR, 0777)
	if err != nil {
		panic(err.Error())
	}
	log.SetOutput(logFile)
}

var logText string

func main() {
	c := colly.NewCollector()
	c.OnHTML("a[href]", func(h *colly.HTMLElement) {
		logText = ""
		var tmpHref = h.Attr("href")
		var tmpTitle = h.Attr("title")
		if strings.Contains(tmpHref, "-c12") == true && strings.Contains(tmpTitle, "苹果") == true {
			gameCount = gameCount + 1
			logText = "title:" + tmpTitle + " href:" + tmpHref
			gameDetail(tmpHref)
			if logText != "" {
				log.Println(logText)
			}
		}
	})

	c.Visit("https://www.jiaoyimao.com/youxi/")
	log.Println(gameCount)
}

func gameDetail(url string) {
	var c = colly.NewCollector()
	c.OnHTML("span[class=em]", func(h *colly.HTMLElement) {
		var goodsCountInt int64
		var err error
		if goodsCountInt, err = strconv.ParseInt(h.Text, 10, 64); err != nil {
			log.Println("xxxxxx")
		}
		if goodsCountInt == 0 {
			logText = ""
			return
		}
		logText = logText + " goodsCount:" + h.Text
	})
	c.Visit(url)
}