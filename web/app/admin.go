package app

import (
	//"fmt"
	//"time"
	"net/http"
	"net/url"
	"strconv"

	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"

	//"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

func (ctx *AuthContext) showAdminError(err error) error {
	return ctx.showAdminMsg(err.Error())
}

func (ctx *AuthContext) showAdminMsg(msg string) error {
	ctx.model["_error"] = msg
	return ctx.renderTemplate("error.html")
}

func showSysLog(_ctx echo.Context) error {
	ctx := _ctx.(*AuthContext)
	if ctx.isLevelUnder(USER_LEVEL_ADMIN) {
		return ctx.Redirect(http.StatusFound, "/")
	}
	ctx.model["lmenu"] = "t9b"
	return ctx.renderTemplate("syslog.html")
}

type evtLogList struct {
	Rows        []appEvent `json:"rows"`
	CurrentPage int        `json:"page"`
	TotalPages  int        `json:"total"`
}

func jsonEventLog(_ctx echo.Context) error {
	ctx := _ctx.(*AuthContext)
	if ctx.isLevelUnder(USER_LEVEL_ADMIN) {
		return ctx.Redirect(http.StatusFound, "/")
	}

	pgNum, _ := strconv.Atoi(ctx.FormValue("pg"))
	nrows, _ := strconv.Atoi(ctx.FormValue("rows"))
	lvMin, _ := strconv.Atoi(ctx.FormValue("lvMin"))
	lvMax, _ := strconv.Atoi(ctx.FormValue("lvMax"))
	tag := ctx.FormValue("tag")

	if nrows < 10 {
		nrows = 10
	} else if nrows > 100 {
		nrows = 100
	}

	var where = make(bson.M)
	where["lv"] = bson.M{"$gte": lvMin, "$lte": lvMax}
	if tag != "" {
		where["tag"] = bson.RegEx{tag, ""}
	}

	sessionCopy := mgoSession.Copy()
	defer sessionCopy.Close()
	coll := mgoDB.With(sessionCopy).C(COLLNAME_EVENTLOG)

	qry := coll.Find(where)
	totalRows, err := qry.Count()
	if err != nil {
		panic(err)
	}

	var json evtLogList
	json.TotalPages = (totalRows + nrows - 1) / nrows
	if json.TotalPages <= pgNum {
		pgNum = json.TotalPages - 1
		if pgNum < 0 {
			pgNum = 0
		}
	}

	json.CurrentPage = pgNum
	skipRows := pgNum * nrows

	//fmt.Printf("lv=%d~%d,pg=%d,skip=%d,rows=%d,tag=%s, totalRows=%d,totalPages=%d\n", lvMin, lvMax, pgNum, skipRows, nrows, tag, totalRows, json.TotalPages)
	err = qry.Sort("-ts").Skip(skipRows).Limit(nrows).All(&json.Rows)
	if err != nil {
		panic(err)
	}

	if json.Rows == nil {
		json.Rows = make([]appEvent, 0)
	}

	return ctx.JSON(http.StatusOK, json)
}

func setupAdminRoutes(ec *echo.Echo) {
	grp := ec.Group("/admin",
		middleware.CSRFWithConfig(middleware.CSRFConfig{
			TokenLookup: "form:_csrf",
			CookiePath:  "/admin",
		}),
		func(h echo.HandlerFunc) echo.HandlerFunc {
			return func(_ctx echo.Context) error {
				ctx := _ctx.(*AuthContext)

				if !ctx.isLoggedIn() {
					curUri := ctx.Path()
					qs := ctx.QueryString()
					if qs != "" {
						curUri += "?" + qs
					}
					return ctx.Redirect(http.StatusFound, "/login?url="+url.QueryEscape(curUri))
				}

				if ctx.isLevelUnder(USER_LEVEL_MEDIA) {
					return ctx.Redirect(http.StatusFound, "/")
				}

				_csrf := ctx.Get("csrf")
				if _csrf == nil {
					_csrf = "nil"
				}
				ctx.model["_csrf"] = _csrf

				return h(ctx)
			}
		},
	)

	grp.GET("/syslog", showSysLog)        // >9
	grp.GET("/syslog.json", jsonEventLog) // >9

	setupStaffRoutes(grp)

	setupNewsRoutes(grp)
	setupCrawlingRoutes(grp)
	setupYear1Routes(grp)
	setupAnnotateRoutes(grp)
	setupDicRoutes(grp)
	setupLabRoutes(grp)
}
