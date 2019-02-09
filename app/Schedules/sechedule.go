package Schedules

import (
	"collyMao/app/Models"
	"collyMao/orm"
	"github.com/PuerkitoBio/goquery"
	"github.com/gocolly/colly"
	"log"
	"math/rand"
	"regexp"
	"strconv"
	"strings"
	"time"
)

func InsertGoodsCount() {
	for {
		c := colly.NewCollector()
		c.OnHTML("a[href]", func(h *colly.HTMLElement) {
			var tmpHref = h.Attr("href")
			var tmpTitle = h.Attr("title")
			if strings.Contains(tmpHref, "-c12") == true && strings.Contains(tmpTitle, "苹果") == true {
				gameDetail(&tmpHref, &tmpTitle)
			}
		})

		c.Visit("https://www.jiaoyimao.com/youxi/")

		time.Sleep(time.Hour * 24)
	}
}

func gameDetail(url, title *string) {
	var c = colly.NewCollector()
	c.OnHTML("span[class=em]", func(h *colly.HTMLElement) {
		var goodsCountInt int64
		var err error
		if goodsCountInt, err = strconv.ParseInt(h.Text, 10, 64); err != nil {
			log.Println("xxxxxx")
			panic("aaa")
		}
		if goodsCountInt == 0 {
			return
		}
		var maoGame = Models.TableMaoGames{}
		orm.Gorm.Where("url = ?", *url).First(&maoGame)
		if maoGame.Id == 0 {
			//no row
			maoGame.Url = *url
			maoGame.Title = *title
			var valid = regexp.MustCompile(`[\d]{1,}`)
			var gameId, _ = strconv.ParseInt(valid.FindStringSubmatch(*url)[0], 10, 64)
			maoGame.GameId = gameId
			orm.Gorm.Create(&maoGame)
		}
		//记录商品数
		var tmpDate = time.Now().Format("2006-01-02")
		var maoGamesGoodsDetail = Models.TableMaoGamesGoodsCount{}
		orm.Gorm.Where("game_id = ? and create_date = ?", maoGame.GameId, tmpDate).First(&maoGamesGoodsDetail)
		if maoGamesGoodsDetail.Id == 0 {
			//now row
			maoGamesGoodsDetail = Models.TableMaoGamesGoodsCount{
				GameId:     maoGame.GameId,
				CreateDate: tmpDate,
				GoodsCount: goodsCountInt,
			}
			orm.Gorm.Create(&maoGamesGoodsDetail)
		}
	})
	c.Visit(*url)

}

//获取游戏下每个商品的销量情况
func InsertGoodsDetail() {
	for {
		var maoGamesSlice = []Models.TableMaoGames{}
		orm.Gorm.Find(&maoGamesSlice)
		for _, maoGame := range maoGamesSlice {
			collyGoodsDetail(maoGame.Url, maoGame.GameId, true)
			time.Sleep(time.Second * 2)
		}
		time.Sleep(time.Second * 10)
	}
}

func collyGoodsDetail(url string, gameId int64, deepVists bool) {
	var c = colly.NewCollector()
	c.OnHTML(`ul[class="list-con specialList"] > li`, func(h *colly.HTMLElement) {
		var price, count, categoryId, title, goodsId, goodsUrl string
		h.DOM.Children().Each(func(a int, s *goquery.Selection) {
			if attrV, _ := s.Attr("class"); attrV == "price" {
				price = s.Text()
			}
			if attrV, _ := s.Attr("class"); attrV == "count" {
				count = s.Text()
			}
			if attrV, _ := s.Attr("name"); attrV == "goodsbg" {
				categoryId, _ = s.Attr("category-id")
			}
			if attrV, _ := s.Attr("class"); attrV == "name" {
				s.ChildrenFiltered("a").Each(func(a int, s *goquery.Selection) {
					title = s.Text()
					//goodsId
					goodsUrl, _ = s.Attr("href")
					var valid = regexp.MustCompile(`[\d]{3,}`)
					goodsId = valid.FindStringSubmatch(goodsUrl)[0]
				})
			}
		})
		//商品存在
		if price != "" && count != "" && title != "" && categoryId != "" && goodsId != "" {
			var maoGameGoods = Models.TableMaoGamesGoods{}
			orm.Gorm.Where("goods_id = ?", goodsId).First(&maoGameGoods)
			if maoGameGoods.Id == 0 {
				//create
				maoGameGoods.GoodsId, _ = strconv.ParseInt(goodsId, 10, 64)
				maoGameGoods.GameId = gameId
				maoGameGoods.Url = goodsUrl
				orm.Gorm.Create(&maoGameGoods)
			}
			//计算商品数量有无变化
			var maoGameGoodsDetail = Models.TableMaoGamesGoodsDetail{}
			orm.Gorm.Where("goods_id = ? and goods_count = ?", goodsId, count).Last(&maoGameGoodsDetail)
			if maoGameGoodsDetail.Id != 0 {
				//无变化
				return
			}
			var floatPrice, _ = strconv.ParseFloat(price, 64)
			var IntCount, _ = strconv.ParseInt(count, 10, 64)
			maoGameGoodsDetail = Models.TableMaoGamesGoodsDetail{
				CreateDatetime: time.Now().Format("2006-01-02 15:04:05"),
				Price:          floatPrice,
				GoodsCount:     IntCount,
				Title:          title,
				GoodsId:        maoGameGoods.GoodsId,
			}
			orm.Gorm.Create(&maoGameGoodsDetail)
		}
	})

	if deepVists == true {
		c.OnHTML(`span[class=page-count] > a`, func(h *colly.HTMLElement) {
			if intPage, _ := strconv.ParseInt(h.Text, 10, 64); intPage <= 5 {
				if intPage == 1 {
					return
				}
				collyGoodsDetail(h.Attr("href"), gameId, false)
			}
		})
	}
	log.Println(url)
	time.Sleep(time.Second * time.Duration(rand.Int63n(3)))
	c.Visit(url)
}

func Start() {
	go InsertGoodsCount()
	go InsertGoodsDetail()
}
