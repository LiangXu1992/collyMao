package Schedules

import (
	"collyMao/app/Models"
	"collyMao/orm"
	"github.com/PuerkitoBio/goquery"
	"github.com/gocolly/colly"
	"log"
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
			orm.Gorm.Create(&maoGame)
		}
		//记录商品数
		var tmpDate = time.Now().Format("2006-01-02")
		var maoGamesGoodsDetail = Models.TableMaoGamesGoodsCount{}
		orm.Gorm.Where("game_id = ? and create_date = ?", maoGame.Id, tmpDate).First(&maoGamesGoodsDetail)
		if maoGamesGoodsDetail.Id == 0 {
			//now row
			maoGamesGoodsDetail = Models.TableMaoGamesGoodsCount{
				GameId:     maoGame.Id,
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
	var maoGamesSlice = []Models.TableMaoGames{}
	orm.Gorm.Find(&maoGamesSlice)
	for _, maoGame := range maoGamesSlice {
		if maoGame.Id != 687 {
			continue
		}
		var c = colly.NewCollector()
		c.OnHTML(`ul[class="list-con specialList"] > li`, func(h *colly.HTMLElement) {
			var price, count, categoryId, title, goodsId string
			h.DOM.Children().Each(func(a int, s *goquery.Selection) {
				if attrV, _ := s.Attr("class"); attrV == "price" {
					price = s.Text()
					log.Println("price:" + price)
				}
				if attrV, _ := s.Attr("class"); attrV == "count" {
					count = s.Text()
					log.Println("count:" + count)
				}
				if attrV, _ := s.Attr("name"); attrV == "goodsbg" {
					categoryId, _ = s.Attr("category-id")
					log.Println("category-id:" + categoryId)

				}
				if attrV, _ := s.Attr("class"); attrV == "name" {
					s.ChildrenFiltered("a").Each(func(a int, s *goquery.Selection) {
						log.Println("title:" + s.Text())
						title = s.Text()
						//goodsId
						var tmpHref, _ = s.Attr("href")
						var valid = regexp.MustCompile(`[\d]{3,}`)
						goodsId = valid.FindStringSubmatch(tmpHref)[0]
						log.Println("goods_id:" + goodsId)
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
					maoGameGoods.MaoGamesId = maoGame.Id
					orm.Gorm.Create(&maoGameGoods)
				}
				var floatPrice, _ = strconv.ParseFloat(price, 64)
				var IntCount, _ = strconv.ParseInt(count, 10, 64)
				var maoGameGoodsDetail = Models.TableMaoGamesGoodsDetail{
					MaoGamesGoodsId: maoGameGoods.Id,
					CreateDatetime:  time.Now().Format("2006-01-02 15:04:05"),
					Price:           floatPrice,
					GoodsCount:      IntCount,
					Title:           title,
				}
				orm.Gorm.Create(&maoGameGoodsDetail)
			}
		})
		c.Visit(maoGame.Url)
	}
}

func Start() {
	//go InsertGoodsCount()
	InsertGoodsDetail()

}
