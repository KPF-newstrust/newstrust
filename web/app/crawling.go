package app

import (
	"net/http"
	"regexp"
	"strconv"
	"time"

	"github.com/labstack/echo"
	"gopkg.in/mgo.v2/bson"
)

type NewsCrawlingRaw struct {
	Id        bson.ObjectId `bson:"_id,omitempty" json:"DT_RowId"`
	NewsId    string        `bson:"newsId" json:"newsId"`
	Title     string        `bson:"title" json:"title"`
	CpId      int           `bson:"cpId" json:"cpId"`
	CpKorName string        `bson:"cpKorName" json:"cpKorName"`
	CpEngName string        `bson:"cpEngName" json:"cpEngName"`
	RegDt     string        `bson:"regDt" json:"regDt"`
	ModiDt    string        `bson:"modiDt" json:"modiDt"`
	Thumbnail string        `bson:"contents" json:"contents"`
	Content   string        `bson:"mobileNews" json:"mobileNews"`
	ImageUrl  string        `bson:"imageUrl" json:"imageUrl`
	VideoUrl  string        `bson:"videoUrl" json:"videoUrl`
}

type NewsCrawling struct {
	Id           bson.ObjectId  `bson:"_id,omitempty" json:"DT_RowId"`
	InsertDt     time.Time      `bson:"insertDt" json:"insertDt"`
	NewsId       string         `bson:"newsId" json:"newsId"`
	MediaId      string         `bson:"mediaId" json:"mediaId"`
	MediaName    string         `bson:"mediaName" json:"mediaName"`
	Title        string         `bson:"title" json:"title"`
	Content      string         `bson:"content" json:"content"`
	PubDate      string         `bson:"pubDate" json:"pubDate"`
	Byline       string         `bson:"byline" json:"byline"` // DELETE soon
	BylineWriter writerByline   `bson:"byline_writer" json:"bylineWriter"`
	Bylines      []writerByline `bson:"bylines" json:"bylines"`
	Url          string         `bson:"url" json:"url"`

	Category      string `bson:"category" json:"category"`
	CategoryXls   string `bson:"categoryXls" json:"categoryXls"`
	CategoryMan   string `bson:"categoryMan" json:"categoryMan"`
	CategoryCalc  string `bson:"categoryCalc" json:"categoryCalc"`
	CategoryFinal string `bson:"categoryFinal" json:"categoryFinal"`

	ImageCount                   int      `bson:"image_count"`
	ContentLength                int      `bson:"content_length" json:"content_length"`
	ContentNumberCount           int      `bson:"content_numNumber"`
	ContentAvgSentenceLength     float64  `bson:"content_avgSentenceLength"`
	ContentAvgAdverbsPerSentence float64  `bson:"content_avgAdverbsPerSentence"`
	ContentQuotePercent          float64  `bson:"content_quotePercent"`
	ContentAnonPredicates        []string `bson:"content_anonPredicates"`
	ContentForeignWords          []string `bson:"content_foreignWords"` // TODO
	//InformantReal int	`bson:"informantReal"`
	//QuoteRatioRealAnon int	`bson:"quoteRatioRealAnon"`

	TitleLength  int      `bson:"title_length" json:"title_length"`
	TitleAdverbs []string `bson:"title_adverbs"`

	TitleNumExclamation int                  `bson:"title_numExclamation" json:"title_numExclamation"`
	TitleNumQuestion    int                  `bson:"title_numQuestion" json:"title_numQuestion"`
	TitleNumSingleQuote int                  `bson:"title_numSingleQuote" json:"title_numSingleQuote"`
	TitleNumDoubleQuote int                  `bson:"title_numDoubleQuote" json:"title_numDoubleQuote"`
	TitleHasShock       int                  `bson:"title_hasShock" json:"title_hasShock"`
	TitleHasExclusive   int                  `bson:"title_hasExclusive" json:"title_hasExclusive"`
	TitleHasBreaking    int                  `bson:"title_hasBreaking" json:"title_hasBreaking"`
	TitleHasPlan        int                  `bson:"title_hasPlan" json:"title_hasPlan"`
	QuotedSentences     []quotedSentenceItem `bson:"quotes" json:"quotes"`
	Score               scoreInfo            `bson:"score" json:"score"`

	NerPersons       []string `bson:"ner_PS" json:"ner_PS"`
	NerOrganizations []string `bson:"ner_OG" json:"ner_OG"`
	NerLocations     []string `bson:"ner_LC" json:"ner_LC"`

	InformantReal []string `bson:"informant_real" json:"informant_real"`
	InformantAnno []string `bson:"informant_anno" json:"informant_anno"`

	ScoreSum   float64 `bson:"score_totalSum"`
	JournalSum float64 `bson:"journal_totalSum"`
	Journal    struct {
		Readability    float64 `bson:"readability"`
		Transparency   float64 `bson:"transparency"`
		Factuality     float64 `bson:"factuality"`
		Utility        float64 `bson:"utility"`
		Fairness       float64 `bson:"fairness"`
		Diversity      float64 `bson:"diversity"`
		Originality    float64 `bson:"originality"`
		Importance     float64 `bson:"importance"`
		Depth          float64 `bson:"depth"`
		Sensationalism float64 `bson:"sensationalism"`
	} `bson:"journal"`

	VanillaSum float64 `bson:"vanilla_totalSum"`
	Vanilla    struct {
		Readability    float64 `bson:"readability"`
		Transparency   float64 `bson:"transparency"`
		Factuality     float64 `bson:"factuality"`
		Utility        float64 `bson:"utility"`
		Fairness       float64 `bson:"fairness"`
		Diversity      float64 `bson:"diversity"`
		Originality    float64 `bson:"originality"`
		Importance     float64 `bson:"importance"`
		Depth          float64 `bson:"depth"`
		Sensationalism float64 `bson:"sensationalism"`
	} `bson:"vanilla"`

	Evaluation struct {
		Readability    float64 `bson:"readability"`  //독이성
		Transparency   float64 `bson:"clariry"`      //투명성
		Factuality     float64 `bson:"reality"`      //사실성
		Utility        float64 `bson:"usefulness"`   //유용성
		Fairness       float64 `bson:"balance"`      //균형성
		Diversity      float64 `bson:"variety"`      //다양성
		Originality    float64 `bson:"uniqueness"`   //독창성
		Importance     float64 `bson:"importance"`   //중요성
		Depth          float64 `bson:"deep"`         //심층성
		Sensationalism float64 `bson:"inflammation"` //선정성
		Average        float64 `bson:"average"`
	} `bson:"evaluation"`

	ClusterDelegate   bool    `bson:"clusterDelegate"`
	ClusterNewsId     string  `bson:"clusterNewsId"`
	ClusterId         string  `bson:"clusterId"`
	ClusterSimilarity float64 `bson:"clusterSimilarity"`

	WeightedJournalSum float64 `bson:"weight"`
}

func jsonCrawlingRawList(_ctx echo.Context) error {
	ctx := _ctx.(*AuthContext)

	skipRows, _ := strconv.Atoi(ctx.FormValue("start"))
	fetchRows, _ := strconv.Atoi(ctx.FormValue("length"))
	if fetchRows <= 0 {
		fetchRows = 10
	}

	var jsonData JqDataTable
	jsonData.Draw, _ = strconv.Atoi(ctx.FormValue("draw"))
	// orderDir := ctx.FormValue("order[0][dir]")
	// orderColumn, _ := strconv.Atoi(ctx.FormValue("order[0][column]"))

	// filterMediaId := ctx.FormValue("columns[1][search][value]")
	// filterMediaName := ctx.FormValue("columns[2][search][value]")
	// filterCategory1 := ctx.FormValue("columns[3][search][value]")
	// filterCategory2 := ctx.FormValue("columns[4][search][value]")
	// filterCategory3 := ctx.FormValue("columns[5][search][value]")
	// filterByline := ctx.FormValue("columns[6][search][value]")
	// filterTitle := ctx.FormValue("columns[7][search][value]")
	// filterNewsId := ctx.FormValue("search[value]")

	// 검색
	var err error
	var findParam = make(bson.M)

	// if filterNewsId != "" {
	// 	findParam["newsId"] = bson.RegEx{filterNewsId, ""}
	// } else {
	// 	ymd := ctx.FormValue("ymd")
	// 	dtBegin, err := time.ParseInLocation("2006-01-02", ymd, tzLocation)
	// 	if err == nil {
	// 		dtEnd := dtBegin.AddDate(0, 0, 1)
	// 		findParam["insertDt"] = bson.M{"$gte": dtBegin, "$lt": dtEnd}
	// 	}

	// 	if filterMediaId != "" {
	// 		findParam["mediaId"] = bson.RegEx{filterMediaId, ""}
	// 	}
	// 	if filterMediaName != "" {
	// 		findParam["mediaName"] = bson.RegEx{filterMediaName, ""}
	// 	}
	// 	if filterCategory1 != "" {
	// 		findParam["categoryXls"] = bson.RegEx{filterCategory1, ""}
	// 	}
	// 	if filterCategory2 != "" {
	// 		findParam["categoryMan"] = bson.RegEx{filterCategory2, ""}
	// 	}
	// 	if filterCategory3 != "" {
	// 		findParam["categoryCalc"] = bson.RegEx{filterCategory3, ""}
	// 	}
	// 	if filterTitle != "" {
	// 		findParam["title"] = bson.RegEx{filterTitle, ""}
	// 	}

	// 	if strings.HasPrefix(filterByline, "~!") {
	// 		notIncl := filterByline[2:]
	// 		if notIncl == "" {
	// 			findParam["bylines"] = nil
	// 		} else {
	// 			findParam["bylines"] = bson.M{"$elemMatch": bson.M{"$and": []bson.M{
	// 				bson.M{"name": bson.M{"$exists": true}},
	// 				bson.M{"name": bson.M{"$not": bson.RegEx{notIncl, ""}}},
	// 			}}}
	// 		}
	// 	} else if filterByline != "" {
	// 		findParam["bylines"] = bson.M{"$elemMatch": bson.M{"$or": []bson.M{
	// 			bson.M{"name": bson.RegEx{filterByline, ""}},
	// 			bson.M{"email": bson.RegEx{filterByline, ""}},
	// 		}}}
	// 	}
	// }

	sessionCopy := mgoSessionC.Copy()
	defer sessionCopy.Close()
	coll := mgoDBC.With(sessionCopy).C("news-c")

	jsonData.Total, err = coll.Count()
	if err != nil {
		panic(err)
	}
	qry := coll.Find(&findParam)

	jsonData.Filtered, err = qry.Count()
	if err != nil {
		panic(err)
	}

	// 정렬
	var orderFieldName string
	orderFieldName += "_id"
	/*	if orderDir == "desc" {
			orderFieldName = "-"
		} else {
			orderFieldName = ""
		}

		switch orderColumn {
		case 0:
			orderFieldName += "insertDt"
		case 1:
			orderFieldName += "mediaId"
		case 2:
			orderFieldName += "mediaName"
		case 3:
			orderFieldName += "categoryXls"
		case 4:
			orderFieldName += "categoryMan"
		case 5:
			orderFieldName += "categoryCalc"
		case 6:
			orderFieldName += "byline"
		case 7:
		default:
			orderFieldName += "title"
		} */

	selFields := bson.M{
		"newsId":    1,
		"cpId":      1,
		"cpKorName": 1,
		"title":     1,
		"contents":  1,
		"regDt":     1,
	}
	var results []NewsCrawlingRaw
	err = qry.Sort(orderFieldName).Skip(skipRows).Limit(fetchRows).Select(selFields).All(&results)
	if err != nil {
		panic(err)
	}

	jsonData.Rows = make([]interface{}, len(results))
	for i := range results {
		jsonData.Rows[i] = &results[i]
	}

	return ctx.JSON(http.StatusOK, jsonData)
}

//NEW
func jsonCrawlingList(_ctx echo.Context) error {
	ctx := _ctx.(*AuthContext)

	skipRows, _ := strconv.Atoi(ctx.FormValue("start"))
	fetchRows, _ := strconv.Atoi(ctx.FormValue("length"))
	if fetchRows <= 0 {
		fetchRows = 10
	}

	var jsonData JqDataTable
	jsonData.Draw, _ = strconv.Atoi(ctx.FormValue("draw"))

	evaluated := ctx.FormValue("evaluated")

	// 검색
	var err error
	var findParam = make(bson.M)
	if evaluated == "true" {
		findParam["evaluation"] = bson.M{"$exists": true}
	}

	// 	if strings.HasPrefix(filterByline, "~!") {
	// 		notIncl := filterByline[2:]
	// 		if notIncl == "" {
	// 			findParam["bylines"] = nil
	// 		} else {
	// 			findParam["bylines"] = bson.M{"$elemMatch": bson.M{"$and": []bson.M{
	// 				bson.M{"name": bson.M{"$exists": true}},
	// 				bson.M{"name": bson.M{"$not": bson.RegEx{notIncl, ""}}},
	// 			}}}
	// 		}
	// 	} else if filterByline != "" {
	// 		findParam["bylines"] = bson.M{"$elemMatch": bson.M{"$or": []bson.M{
	// 			bson.M{"name": bson.RegEx{filterByline, ""}},
	// 			bson.M{"email": bson.RegEx{filterByline, ""}},
	// 		}}}
	// 	}
	// }

	sessionCopy := mgoSession.Copy()
	defer sessionCopy.Close()
	coll := mgoDB.With(sessionCopy).C("crawling")

	jsonData.Total, err = coll.Count()
	if err != nil {
		panic(err)
	}
	qry := coll.Find(&findParam)

	jsonData.Filtered, err = qry.Count()
	if err != nil {
		panic(err)
	}

	// 정렬
	var orderFieldName string
	orderFieldName += "_id"

	selFields := bson.M{
		"newsId":       1,
		"mediaName":    1,
		"categoryCalc": 1,
		"title":        1,
		// "contents":  1,
		// "regDt":     1,
	}
	var results []NewsCrawling
	err = qry.Sort(orderFieldName).Skip(skipRows).Limit(fetchRows).Select(selFields).All(&results)
	if err != nil {
		panic(err)
	}

	jsonData.Rows = make([]interface{}, len(results))
	for i := range results {
		jsonData.Rows[i] = &results[i]
	}

	return ctx.JSON(http.StatusOK, jsonData)
}

func showCrawlingRawList(_ctx echo.Context) error {
	ctx := _ctx.(*AuthContext)
	ctx.model["lmenu"] = "t0c"
	return ctx.renderTemplate("crawling_raw_list.html")
}

func showCrawlingList(_ctx echo.Context) error {
	ctx := _ctx.(*AuthContext)
	ctx.model["lmenu"] = "t0d"
	return ctx.renderTemplate("crawling_list.html")
}

func readCrawlingRawContent(_ctx echo.Context) error {
	ctx := _ctx.(*AuthContext)
	newsId := ctx.Param("id")

	ctx.model["item"] = getNewsCrawlingRaw(newsId)
	ctx.model["lmenu"] = "t0c"
	return ctx.renderTemplate("crawling_view.html")
}

func showCrawlingRawContent(_ctx echo.Context) error {
	ctx := _ctx.(*AuthContext)
	newsId := ctx.Param("id")

	ctx.model["item"] = getNewsCrawlingRaw(newsId)
	ctx.model["lmenu"] = "t0c"
	return ctx.renderTemplate("crawling_raw_read.html")
}

func getNewsCrawlingRaw(id string) NewsCrawlingRaw {
	sessionCopy := mgoSessionC.Copy()
	defer sessionCopy.Close()
	coll := mgoDBC.With(sessionCopy).C("news-c")

	var data NewsCrawlingRaw
	err := coll.Find(bson.D{{"newsId", id}}).One(&data)
	if err != nil {
		// panic(err)
	}

	// \<section([\W\w]+)section\>
	r, _ := regexp.Compile("\\<section([\\W\\w]+)section\\>")
	content := r.FindString(data.Content)
	data.Content = content

	return data
}

func readCrawlingContent(_ctx echo.Context) error {
	ctx := _ctx.(*AuthContext)
	newsId := ctx.Param("id")
	where := bson.D{{"newsId", newsId}}

	sessionCopy := mgoSession.Copy()
	defer sessionCopy.Close()
	coll := mgoDB.With(sessionCopy).C(COLLNAME_CRAWLING)

	var data NewsCrawling
	err := coll.Find(where).One(&data)
	if err != nil {
		return ctx.showAdminMsg("Crawling News not found")
	}
	ctx.model["item"] = &data
	ctx.model["raw"] = getNewsCrawlingRaw(newsId)

	ctx.model["lmenu"] = "t0d"
	return ctx.renderTemplate("crawling_read.html")
}

func readCrawlingEntity(_ctx echo.Context) error {
	ctx := _ctx.(*AuthContext)
	newsId := ctx.Param("id")

	sessionCopy := mgoSession.Copy()
	defer sessionCopy.Close()

	// coll_news := mgoDB.With(sessionCopy).C(COLLNAME_NEWS)
	// var dataNews MdbNews
	// err := coll_news.Find(bson.M{"newsId":newsId}).Select(bson.M{"ner_PS":1, "ner_OG":1, "ner_LC":1}).One(&dataNews)
	// if err != nil {
	// 	return ctx.showAdminMsg("Object not found")
	// } else {
	// 	ctx.model["item"] = &dataNews
	// }

	coll := mgoDB.With(sessionCopy).C(COLLNAME_CRAWLING_ENTITY)
	var dataEntity MdbNewsEntity
	err := coll.Find(bson.M{"newsId": newsId}).One(&dataEntity)
	if err == nil {
		ctx.model["entity"] = &dataEntity
	}

	return ctx.renderTemplate("crawling_read_entity.html")
}

func setupCrawlingRoutes(grp *echo.Group) {
	grp.GET("/crawling/raw", showCrawlingRawList)
	grp.GET("/crawling", showCrawlingList)
	grp.POST("/crawling/raw.json", jsonCrawlingRawList)
	grp.POST("/crawling.json", jsonCrawlingList)
	grp.GET("/crawling/raw/:id", readCrawlingRawContent)
	grp.GET("/crawling/:id", readCrawlingContent)
	grp.GET("/crawling/:id/entity", readCrawlingEntity)
}
