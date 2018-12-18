package app

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/labstack/echo"
	"gopkg.in/mgo.v2/bson"
)

type DicCustom struct {
	Id       bson.ObjectId `bson:"_id,omitempty" json:"_id"`
	Word     string        `bson:"word" json:"word"`
	Meta     []string      `bson:"meta" json:"meta"`
	Date     time.Time     `bson:"date" json:"date"`
	Applied  bool          `bson:"applied" json:"applied"`
	UserId   string        `bson:"user_id" json:"user_id"`
	UserName string        `bson:"user_name" json:"user_name"`
}
type Dic struct {
	Id        bson.ObjectId `bson:"_id,omitempty" json:"DT_RowId"`
	Source    string        `bson:"source" json:"source"`
	Word      string        `bson:"word" json:"word"`
	Tag       string        `bson:"tag" json:"tag"`
	Meaning   string        `bson:"meaning" json:"meaning"`
	Tail      string        `bson:"tail" json:"tail"`
	Sound     string        `bson:"sound" json:"sound"`
	Type      string        `bson:"type" json:"type"`
	TagFirst  string        `bson:"tag_first" json:"tag_first"`
	TagLast   string        `bson:"tag_last" json:"tag_last"`
	Structure string        `bson:"structure" json:"structure"`
	Version   string        `bson:"version" json:"version"`
	Meta      []string      `bson:"meta,omitempty" json:"meta"`
}

func jsonMecabList(_ctx echo.Context) error {
	ctx := _ctx.(*AuthContext)

	skipRows, _ := strconv.Atoi(ctx.FormValue("start"))
	fetchRows, _ := strconv.Atoi(ctx.FormValue("length"))
	if fetchRows <= 0 {
		fetchRows = 10
	}

	filterWord := ctx.FormValue("search[value]")
	filterLike := ctx.FormValue("like")

	var jsonData JqDataTable
	jsonData.Draw, _ = strconv.Atoi(ctx.FormValue("draw"))

	var findParam = make(bson.M)

	if filterWord != "" {
		fmt.Print(filterLike)
		if filterLike == "true" {
			findParam["word"] = bson.RegEx{filterWord, ""}
		} else {
			findParam["word"] = filterWord
		}
	}

	sessionCopy := mgoSession.Copy()
	defer sessionCopy.Close()
	coll := mgoDB.With(sessionCopy).C(COLLNAME_DICTIONARY)

	var err error
	jsonData.Total, err = coll.Count()
	if err != nil {
		panic(err)
	}

	qry := coll.Find(&findParam)

	jsonData.Filtered, err = qry.Count()
	if err != nil {
		panic(err)
	}

	var results []Dic
	err = qry.Sort("-_id").Skip(skipRows).Limit(fetchRows).Select(bson.M{}).All(&results)
	if err != nil {
		panic(err)
	}

	jsonData.Rows = make([]interface{}, len(results))
	for i := range results {
		jsonData.Rows[i] = &results[i]
	}

	return ctx.JSON(http.StatusOK, jsonData)
}

func showMecabList(_ctx echo.Context) error {
	ctx := _ctx.(*AuthContext)
	ctx.model["lmenu"] = "t0a"
	return ctx.renderTemplate("dic_mecab_list.html")
}

func jsonCustomList(_ctx echo.Context) error {
	ctx := _ctx.(*AuthContext)

	orderDir := ctx.FormValue("order[0][dir]")
	orderColumn, _ := strconv.Atoi(ctx.FormValue("order[0][column]"))
	skipRows, _ := strconv.Atoi(ctx.FormValue("start"))
	fetchRows, _ := strconv.Atoi(ctx.FormValue("length"))
	if fetchRows <= 0 {
		fetchRows = 10
	}

	filterWord := ctx.FormValue("search[value]")
	filterLike := ctx.FormValue("like")

	var jsonData JqDataTable
	jsonData.Draw, _ = strconv.Atoi(ctx.FormValue("draw"))

	var findParam = make(bson.M)

	if filterWord != "" {
		fmt.Print(filterLike)
		if filterLike == "true" {
			findParam["word"] = bson.RegEx{filterWord, ""}
		} else {
			findParam["word"] = filterWord
		}
	}

	sessionCopy := mgoSession.Copy()
	defer sessionCopy.Close()
	coll := mgoDB.With(sessionCopy).C(COLLNAME_DICTIONARY_CUSTOM)

	var err error
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
	var orderFieldName = "-"
	if orderDir == "asc" {
		orderFieldName = ""
	}
	switch orderColumn {
	case 6:
		orderFieldName += "applied"
	default:
		orderFieldName += "_id"
	}

	var results []DicCustom
	err = qry.Sort(orderFieldName).Skip(skipRows).Limit(fetchRows).Select(bson.M{}).All(&results)
	if err != nil {
		panic(err)
	}

	jsonData.Rows = make([]interface{}, len(results))
	for i := range results {
		jsonData.Rows[i] = &results[i]
	}

	return ctx.JSON(http.StatusOK, jsonData)
}

func showCustomList(_ctx echo.Context) error {
	ctx := _ctx.(*AuthContext)
	ctx.model["lmenu"] = "t0b"
	return ctx.renderTemplate("dic_custom_list.html")
}

func postCustom(_ctx echo.Context) error {
	ctx := _ctx.(*AuthContext)

	var dicCustom DicCustom
	dicCustom.Word = ctx.FormValue("word")
	dicCustom.Meta = strings.Split(ctx.FormValue("meta"), ",")
	dicCustom.Date = time.Now()
	dicCustom.UserId = ctx.authUserId
	dicCustom.UserName = fmt.Sprintf("%s", ctx.model[MVK_AuthName])

	sessionCopy := mgoSession.Copy()
	defer sessionCopy.Close()
	coll := mgoDB.With(sessionCopy).C(COLLNAME_DICTIONARY_CUSTOM)

	var err error
	err = coll.Insert(dicCustom)
	if err != nil {
		return ctx.showAdminError(err)
	}

	ctx.flashSuccess("단어가 등록되었습니다. 관리자의 승인 후 형태소사전에 최종 등록됩니다. (\"%s\")", dicCustom.Word)

	return ctx.Redirect(http.StatusFound, "/dic/mecab")
}

func putCustom(_ctx echo.Context) error {
	ctx := _ctx.(*AuthContext)

	words := strings.Split(ctx.FormValue("words"), ",")

	fmt.Println("words", words)
	bwords := Map(words, bson.ObjectIdHex)
	fmt.Println("words", bwords)

	result := 0

	if words != nil {
		sessionCopy := mgoSession.Copy()
		defer sessionCopy.Close()
		coll := mgoDB.With(sessionCopy).C(COLLNAME_DICTIONARY_CUSTOM)

		info, err := coll.UpdateAll(bson.M{"_id": bson.M{"$in": bwords}}, bson.M{"$set": bson.M{"applied": true}})
		fmt.Println("info", info)
		fmt.Println("err", err)
		result = info.Updated
	}

	return ctx.JSON(http.StatusOK, result)
}

func Map(vs []string, f func(string) bson.ObjectId) []bson.ObjectId {
	vsm := make([]bson.ObjectId, len(vs))
	for i, v := range vs {
		vsm[i] = f(v)
	}
	return vsm
}

func publishCustom(_ctx echo.Context) error {
	ctx := _ctx.(*AuthContext)

	sessionCopy := mgoSession.Copy()
	defer sessionCopy.Close()
	coll := mgoDB.With(sessionCopy).C(COLLNAME_DICTIONARY_CUSTOM)

	var dicCustom []DicCustom
	err := coll.Find(bson.M{"applied": true}).All(&dicCustom)

	csv := ""
	if len(dicCustom) > 0 {
		for _, d := range dicCustom {
			word := d.Word
			r := []rune(word)
			tail := "T"
			if (int(r[len(r)-1 : len(r)][0])-44032)%28 == 0 {
				tail = "F"
			}
			fmt.Printf("%s / %s\n", word, tail)
			csv += word + ",,,,NNP,*," + tail + "," + word + ",*,*,*,*,*\n"
		}
	}
	// userdic-n
	d_new, err := os.Create("static/userdic-n.csv")
	if err != nil {
		panic(err)
	}
	defer d_new.Close()
	n, err := d_new.WriteString(csv)
	if err != nil || n < 0 {
		panic(err)
	}
	/////////////////////////////////////////////////////////////////
	renew := false
	dat_old, err := ioutil.ReadFile("static/userdic.csv")
	if err != nil {
		renew = true
		println("NO FILE - userdic.csv")
		d_old, err := os.Create("static/userdic.csv")
		if err != nil {
			panic(err)
		}
		defer d_old.Close()
		n, err := d_old.WriteString(csv)
		if err != nil || n < 0 {
			panic(err)
		}

		dat_old, err = ioutil.ReadFile("static/userdic.csv")
	}
	dat_new, err := ioutil.ReadFile("static/userdic-n.csv")
	if err != nil {
		panic(err)
	}

	// print("new\n" + string(dat_new))
	// print("old\n" + string(dat_old))
	if renew == false && strings.Compare(string(dat_new), string(dat_old)) != 0 {
		renew = true
		d_old, err := os.Create("static/userdic.csv")
		if err != nil {
			panic(err)
		}
		defer d_old.Close()
		n, err := d_old.WriteString(csv)
		if err != nil || n < 0 {
			panic(err)
		}
	}

	renew = true //////////////////////////////TEST		*/
	if renew == true {
		println("RENEW - 메시지 발송")
		err = mqSend_Notice()
		if err != nil {
			panic(err)
		}
		return ctx.apiOk("OK", bson.M{"message": "메시지를 발송했습니다."})
	} else {
		return ctx.apiOk("OK", bson.M{"message": "변경사항이 없습니다."})
	}
}

func setupDicRoutes(grp *echo.Group) {
	grp.GET("/dic/custom", showCustomList)
	grp.POST("/dic/custom.json", jsonCustomList)
	grp.POST("/dic/custom.add", putCustom)
	grp.POST("/dic/custom.publish", publishCustom)
}

func setupDicOpenRoutes(ec *echo.Echo) {
	ec.GET("/dic/mecab", showMecabList)
	ec.POST("/dic/mecab.json", jsonMecabList)
	ec.POST("/dic/custom", postCustom)
}
