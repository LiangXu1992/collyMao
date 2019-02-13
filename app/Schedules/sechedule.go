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
		for {
			time.Sleep(time.Second * 10)
			if time.Now().Format("04") == "00" {
				break
			}
		}
		c := colly.NewCollector()
		c.OnHTML("a[href]", func(h *colly.HTMLElement) {
			var tmpHref = h.Attr("href")
			var tmpTitle = h.Attr("title")
			if strings.Contains(tmpHref, "-c12") == true && strings.Contains(tmpTitle, "苹果") == true {
				gameDetail(&tmpHref, &tmpTitle)
				time.Sleep(time.Second * 1)
			}
		})

		c.Visit("https://www.jiaoyimao.com/youxi/")
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

	time.Sleep(time.Second * time.Duration(rand.Int63n(3)))
	c.Visit(url)
}

//统计游戏销量比例
func InsertGameSaleDetail() {
	for {
		for {
			time.Sleep(time.Second * 10)
			if time.Now().Format("04") == "55" {
				break
			}
		}
		//查出数据
		rows, err := orm.DbClient.Query(`select
       count(*) as sc,mao_games_goods_count.goods_count as tc, count(*)/mao_games_goods_count.goods_count  as stc, mao_games.title,mao_games.url,mao_games.game_id
from mao_games
       inner join mao_games_goods on mao_games_goods.game_id = mao_games.game_id
       inner join mao_games_goods_detail on mao_games_goods_detail.goods_id = mao_games_goods.goods_id
       inner join mao_games_goods_count on mao_games_goods_count.game_id = mao_games.game_id and mao_games_goods_count.create_date = CURRENT_DATE()
where mao_games_goods_detail.goods_id in (select goods_id from mao_games_goods_detail as c where c.goods_count < 100 and c.price > 5.00 AND c.create_datetime >= CURRENT_DATE() and c.create_datetime < DATE_SUB(curdate(),INTERVAL -1 DAY) group by c.goods_id having count(*) >= 1)
and mao_games_goods_detail.goods_count < 100 and mao_games_goods_detail.price > 5.00 AND mao_games_goods_detail.create_datetime >= CURRENT_DATE() and mao_games_goods_detail.create_datetime < DATE_SUB(curdate(),INTERVAL -1 DAY)
group by mao_games.game_id
order by stc desc
`)
		defer rows.Close()
		if err != nil {
			log.Println("select stc err")
			break
		}
		for rows.Next() {
			var maoGamesStc Models.TableMaoGamesStc
			err = rows.Scan(&maoGamesStc.SaleCount, &maoGamesStc.GoodsTotalCount, &maoGamesStc.Stc, &maoGamesStc.Title, &maoGamesStc.Url, &maoGamesStc.GameId)
			if err != nil {
				log.Println("stc row scan err")
				break
			}
			maoGamesStc.CreateDatetime = time.Now().Add(time.Hour * 1).Format("2006-01-02 15:00:00")
			orm.Gorm.Create(&maoGamesStc)
		}
		time.Sleep(time.Minute * 2)
	}
}

func Start() {
	go InsertGameSaleDetail()
	go InsertGoodsCount()
	go InsertGoodsDetail()

}
