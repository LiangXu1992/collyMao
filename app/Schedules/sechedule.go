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
				time.Sleep(time.Second * 5)
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
			return
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
			log.Printf("CurrentMaoGame:%+v", maoGame)
			currentVisitPage = 1
			collyGoodsDetail(maoGame.Url, maoGame.GameId, true)
			time.Sleep(time.Millisecond * 300)
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
			//获取游戏的店主信息
			var sellerName, sellerType = getSellerName(goodsUrl)
			if maoGameGoods.Id == 0 {
				//create
				maoGameGoods.GoodsId, _ = strconv.ParseInt(goodsId, 10, 64)
				maoGameGoods.GameId = gameId
				maoGameGoods.Url = goodsUrl
				maoGameGoods.SellerName = sellerName
				maoGameGoods.SellerType = sellerType
				orm.Gorm.Create(&maoGameGoods)
			}
			//计算商品数量有无变化
			//找出当前商品最近一条的记录
			var maoGameGoodsDetail = Models.TableMaoGamesGoodsDetail{}
			orm.Gorm.Where("goods_id = ?", goodsId).Last(&maoGameGoodsDetail)
			//当前商品的数量
			var IntCount, _ = strconv.ParseInt(count, 10, 64)
			//当前商品的价格
			var floatPrice, _ = strconv.ParseFloat(price, 64)
			//商品不存在于数据库
			if maoGameGoodsDetail.Id == 0 {
				maoGameGoodsDetail = Models.TableMaoGamesGoodsDetail{
					CreateDatetime: time.Now().Format("2006-01-02 15:04:05"),
					Price:          floatPrice,
					GoodsCount:     IntCount,
					Title:          title,
					GoodsId:        maoGameGoods.GoodsId,
				}
				orm.Gorm.Create(&maoGameGoodsDetail)
				return
			}
			//1.//无变化,不处理直接返回
			if maoGameGoodsDetail.GoodsCount == IntCount {
				return
			}
			//数量发生了变化
			//1.变化的数量只有一个
			if maoGameGoodsDetail.GoodsCount-IntCount == 1 {
				maoGameGoodsDetail = Models.TableMaoGamesGoodsDetail{
					CreateDatetime: time.Now().Format("2006-01-02 15:04:05"),
					Price:          floatPrice,
					GoodsCount:     IntCount,
					Title:          title,
					GoodsId:        maoGameGoods.GoodsId,
				}
				orm.Gorm.Create(&maoGameGoodsDetail)
				return
			}
			//2.变化的数量不只一个
			if maoGameGoodsDetail.GoodsCount-IntCount > 1 {
				var saleCount = maoGameGoodsDetail.GoodsCount - IntCount
				var oldGoodsCount = maoGameGoodsDetail.GoodsCount
				var i int64 = 0
				for i = 1; i <= saleCount; i++ {
					var floatPrice, _ = strconv.ParseFloat(price, 64)
					maoGameGoodsDetail = Models.TableMaoGamesGoodsDetail{
						CreateDatetime: time.Now().Add(time.Second * time.Duration(i)).Format("2006-01-02 15:04:05"),
						Price:          floatPrice,
						GoodsCount:     oldGoodsCount - i,
						Title:          title,
						GoodsId:        maoGameGoods.GoodsId,
					}
					orm.Gorm.Create(&maoGameGoodsDetail)
				}
				return
			}

			//3.新的商品数量大于最新的一条记录的数量，所以是商家补货，所以更新最后一条记录的数据就ok
			if maoGameGoodsDetail.GoodsCount-IntCount < 1 {
				maoGameGoodsDetail.GoodsCount = IntCount
				maoGameGoodsDetail.Title = title
				maoGameGoodsDetail.Price = floatPrice
				maoGameGoodsDetail.CreateDatetime = time.Now().Format("2006-01-02 15:04:05")
				orm.Gorm.Save(&maoGameGoodsDetail)
				return
			}
		}
	})

	if deepVists == true {
		c.OnHTML(`span[class=page-count] > a`, func(h *colly.HTMLElement) {
			if intPage, _ := strconv.ParseInt(h.Text, 10, 64); intPage <= deepVisitsPage {
				if intPage == 1 {
					return
				}
				if intPage <= currentVisitPage {
					return
				}
				currentVisitPage = intPage
				if intPage == 5 {
					collyGoodsDetail(h.Attr("href"), gameId, true)
				} else {
					collyGoodsDetail(h.Attr("href"), gameId, false)
				}
			}
		})
	}

	time.Sleep(time.Millisecond * time.Duration(rand.Int63n(300)))
	log.Println(url)
	var err = c.Visit(url)
	if err != nil {
		log.Println("collyGoodsDetail visit url err:" + err.Error())
		return
	}
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
where mao_games_goods_detail.goods_id in (select goods_id from mao_games_goods_detail as c where c.goods_count < 100 and c.price >= 3.00 AND c.create_datetime >= CURRENT_DATE() and c.create_datetime < DATE_SUB(curdate(),INTERVAL -1 DAY) group by c.goods_id having count(*) >= 1)
and mao_games_goods_detail.goods_count < 100 and mao_games_goods_detail.price >= 3.00 AND mao_games_goods_detail.create_datetime >= CURRENT_DATE() and mao_games_goods_detail.create_datetime < DATE_SUB(curdate(),INTERVAL -1 DAY)
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

type TableGoodsRank struct {
	Id         int64
	Title      string
	CreateTime int64
	GoodsId    int64
	Idx        int
	GameId     int64
}

type TableTest struct {
	St string
}

// 查看商品排位
func GoodsRank(gameId int64) {
	var c = colly.NewCollector()
	c.OnHTML(`ul[id="goodsList"] > li`, func(h *colly.HTMLElement) {
		if strings.Contains(h.Text, "代充") == true {
			//return
		}
		var valid = regexp.MustCompile(`[\d]+`)
		goodsId, _ := strconv.ParseInt(valid.FindString(h.Attr("id")), 10, 64)

		if h.Index == 0 {
			var d TableGoodsRank
			orm.Gorm.Table("goods_rank").Where("idx = ? and game_id = ?", h.Index, gameId).Last(&d)

			if d.Id == 0 {
				//创建记录
				var insertData = TableGoodsRank{
					Title:      h.ChildAttr("a", "title"),
					CreateTime: time.Now().Unix(),
					GoodsId:    goodsId,
					Idx:        h.Index,
					GameId:     gameId,
				}
				orm.Gorm.Table("goods_rank").Create(&insertData)
				return
			}
			if d.GoodsId == goodsId {
				//无变化
				log.Println(h.ChildAttr("a", "title"))
			} else {
				var updateData = map[string]interface{}{
					"create_time": time.Now().Unix(),
					"title":       h.ChildAttr("a", "title"),
					"goods_id":    goodsId,
				}
				orm.Gorm.Table("goods_rank").Where("idx = ? and game_id = ?", h.Index, gameId).Update(updateData)
			}
		}
	})
	_ = c.Visit("https://m.jiaoyimao.com/g" + strconv.FormatInt(gameId, 10) + "/")
	time.Sleep(time.Millisecond * 100)
}

func allGameGoodsRank() {
	var gameSlice = []int64{
		6587,
		6378,
	}
	for {
		for _, v := range gameSlice {
			GoodsRank(v)
		}
	}
}

//查找店家的id
func getSellerName(goodsUrl string) (sellerName string, sellerType int64) {
	var c = colly.NewCollector()
	c.OnHTML(`div[class="shop-info-card clearfix"] > a`, func(h *colly.HTMLElement) {
		//店铺卖家
		sellerName = "errorShop"
		sellerType = 1
		var reg = regexp.MustCompile(`(com\/)(.*)(\?)`)
		if len(reg.FindAllStringSubmatch(h.Attr("href"), -1)) >= 1 {
			if len(reg.FindAllStringSubmatch(h.Attr("href"), -1)[0]) >= 4 {
				sellerName = reg.FindAllStringSubmatch(h.Attr("href"), -1)[0][2]
				sellerType = 1
			}
		}
	})
	c.OnHTML(`div[class="shop-info-card shop-personal-card clearfix"] > a`, func(h *colly.HTMLElement) {
		//个人卖家
		sellerName = "errorPerson"
		sellerType = 2
		var reg = regexp.MustCompile(`[\d]+`)
		if len(reg.FindAllString(h.Attr("href"), -1)) >= 1 {
			sellerName = reg.FindAllString(h.Attr("href"), -1)[0]
			sellerType = 2
		}
	})
	c.Visit(strings.Replace(goodsUrl, "https://www.", "https://m.", 1))
	return
}

var deepVisitsPage int64 = 10  //搜集多少页的数据
var currentVisitPage int64 = 1 //当前在第几页访问

func Start() {
	go InsertGameSaleDetail()
	go InsertGoodsCount()
	go InsertGoodsDetail()
	go allGameGoodsRank()
}
