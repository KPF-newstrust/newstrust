package app

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/boj/redistore"
	"github.com/gorilla/sessions"
	"github.com/labstack/echo"
	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/middleware"

	"github.com/garyburd/redigo/redis"
	"github.com/patrickmn/go-cache"
	"github.com/streadway/amqp"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

const (
	SVK_AuthUserId = "SAUT_UID"
	SVK_AuthLevel  = "SAUT_Level"
	SVK_AuthName   = "SAUT_Name"

	MVK_AuthUserId = "_AuthUID"
	MVK_AuthLevel  = "_AuthLevel"
	MVK_AuthName   = "_AuthName"
	MVK_AuthPhoto  = "_AuthPhoto"

	_FLK_Success = "flSuccess"
	_FLK_Error   = "flError"

	COLLNAME_YEAR1NEWS = "news"

	COLLNAME_USER              = "user"
	COLLNAME_NEWS              = "news"
	COLLNAME_NEWS_SRC          = "news_src"
	COLLNAME_NEWS_ENTITY       = "news_entity"     //형태소 분석
	COLLNAME_NEWS_AGGR         = "news_aggr"       //종합통계
	COLLNAME_CRAWLING          = "crawling"        //크롤링기사
	COLLNAME_CRAWLING_ENTITY   = "crawling_entity" //크롤링기사 형태소
	COLLNAME_ASSTATS           = "asStats"
	COLLNAME_ANNOTATE          = "annotate"
	COLLNAME_BUGREPORT         = "bugs"
	COLLNAME_APILAB            = "apilab"
	COLLNAME_DICTIONARY        = "dictionary"
	COLLNAME_DICTIONARY_CUSTOM = "dictionary_custom"
	_MaxRedisConnections       = 10

	USER_LEVEL_MEDIA = 5
	USER_LEVEL_ADMIN = 9
)

type AppConfig struct {
	IsDebug     bool
	UseMinified bool
	UseSSL      bool
	ListenPort  int
	GatewayPort int
	Hostname    string
	DsnMongo    string
	DsnMongoC   string
	DsnRedis    string
	DsnAmqp     string
}

type AuthContext struct {
	echo.Context
	model      map[string]interface{}
	sess       *sessions.Session
	authUserId string
}

func (ctx *AuthContext) renderTemplate(filename string) error {
	return ctx.Render(http.StatusOK, filename, ctx.model)
}

func (ctx *AuthContext) renderError(err error) error {
	ctx.model["errorMsg"] = err.Error()
	return ctx.Render(http.StatusInternalServerError, "error.html", ctx.model)
}

func (ctx *AuthContext) flashSuccess(format string, args ...interface{}) {
	ctx.sess.AddFlash(fmt.Sprintf(format, args...), _FLK_Success)
}

func (ctx *AuthContext) flashError(format string, args ...interface{}) {
	ctx.sess.AddFlash(fmt.Sprintf(format, args...), _FLK_Error)
}

var (
	BaseHostname string

	mgoSession  *mgo.Session
	mgoSessionC *mgo.Session
	mgoDB       *mgo.Database
	mgoDBC      *mgo.Database
	mgoOldDB    *mgo.Database
	//	mgoFS *mgo.GridFS

	rdsPool *redis.Pool

	amqpConn       *amqp.Connection
	amqpChn        *amqp.Channel
	amQueue_Task   amqp.Queue
	amqpCloseError chan *amqp.Error
	amQueue_Notice amqp.Queue

	localCache *cache.Cache
	tzLocation *time.Location
)

func makeBaseHostname(config *AppConfig) {
	isHttps := ""
	if config.UseSSL {
		isHttps += "s"
	}

	addPort := ""
	if config.GatewayPort != 80 {
		addPort = fmt.Sprintf(":%d", config.GatewayPort)
	}

	BaseHostname = fmt.Sprintf("http%s://%s%s", isHttps, config.Hostname, addPort)
}

func _ensureSingleColumnUniqueIndex(collname string, fieldName string) {
	coll := mgoDB.With(mgoSession).C(collname)
	err := coll.EnsureIndex(mgo.Index{
		Key:        []string{fieldName},
		Unique:     true,
		DropDups:   true,
		Background: true,
		Sparse:     true,
	})
	if err != nil {
		panic(err)
	}
}

func ensureMongoIndices() {

	coll := mgoDB.With(mgoSession).C(COLLNAME_USER)
	err := coll.EnsureIndex(mgo.Index{
		Key:        []string{"loginId"},
		Unique:     true,
		DropDups:   true,
		Background: true,
		Sparse:     true,
	})
	if err != nil {
		panic(err)
	}
	//admin-test 계정 생성
	var where = make(bson.M)
	where["loginId"] = "admin"
	count, err := coll.Find(where).Count()
	if count == 0 {
		var user AdminUser
		user.CreatedAt = NewJsonNow()
		user.LoginId = "admin"
		user.Passwd.Data = encodePasswd("0000test")
		user.Name = "admin-test"
		user.Email = ""
		user.Phone = ""
		user.Level = USER_LEVEL_ADMIN
		user.Roles = []string{}
		err = coll.Insert(user)
		if err != nil {
			panic(err)
		}
	}
	where["loginId"] = "media"
	count, err = coll.Find(where).Count()
	if count == 0 {
		var user AdminUser
		user.CreatedAt = NewJsonNow()
		user.LoginId = "media"
		user.Passwd.Data = encodePasswd("0000test")
		user.Name = "media-test"
		user.Email = ""
		user.Phone = ""
		user.Level = USER_LEVEL_MEDIA
		user.Roles = []string{}
		err = coll.Insert(user)
		if err != nil {
			panic(err)
		}
	}

	coll = mgoDB.With(mgoSession).C(COLLNAME_NEWS)
	err = coll.EnsureIndex(mgo.Index{
		Key:        []string{"newsId", "insertDt", "clusterNewsId"},
		Unique:     true,
		DropDups:   false,
		Background: true,
		Sparse:     true,
	})
	if err != nil {
		panic(err)
	}

	coll = mgoDB.With(mgoSession).C(COLLNAME_NEWS_SRC)
	err = coll.EnsureIndex(mgo.Index{
		Key:        []string{"insert_dt", "newsitem_id"},
		Unique:     false,
		DropDups:   false,
		Background: true,
		Sparse:     true,
	})
	if err != nil {
		panic(err)
	}

	coll = mgoDB.With(mgoSession).C(COLLNAME_ANNOTATE)
	err = coll.EnsureIndex(mgo.Index{
		Key:        []string{"newsId"},
		Unique:     true,
		DropDups:   false,
		Background: true,
		Sparse:     true,
	})
	if err != nil {
		panic(err)
	}

	coll = mgoDB.With(mgoSession).C(COLLNAME_BUGREPORT)
	err = coll.EnsureIndex(mgo.Index{
		Key:        []string{"newsId"},
		Unique:     false,
		DropDups:   false,
		Background: true,
		Sparse:     true,
	})
	if err != nil {
		panic(err)
	}

	coll = mgoDB.With(mgoSession).C(COLLNAME_DICTIONARY)
	err = coll.EnsureIndex(mgo.Index{
		Key:        []string{"word", "tag", "source"},
		Unique:     true,
		DropDups:   false,
		Background: true,
		Sparse:     true,
	})
	if err != nil {
		panic(err)
	}

	coll = mgoDB.With(mgoSession).C(COLLNAME_NEWS_AGGR)
	err = coll.EnsureIndex(mgo.Index{
		Key:        []string{"category", "mediaType"},
		Unique:     true,
		DropDups:   false,
		Background: true,
		Sparse:     true,
	})
	if err != nil {
		panic(err)
	}
	count, err = coll.Count()
	if count == 0 {
		fmt.Println("Initializing Collection : news_aggr")
		jsonFile, err := os.Open("data/news_aggr.json")
		if err != nil {
			fmt.Println("Open..", err)
		}

		byteValue, _ := ioutil.ReadAll(jsonFile)
		var v []interface{}
		if err := json.Unmarshal(byteValue, &v); err != nil {
			fmt.Println("Unmarshal..", err)
		}
		if err := coll.Insert(v...); err != nil {
			fmt.Println("Insert..", err)
		}
		defer jsonFile.Close()
	}
}

func openMongoDB(config *AppConfig) bool {
	var err error

	log.Printf("MongoDB DSN=%s\n", config.DsnMongo)
	mgoSession, err = mgo.Dial(config.DsnMongo)
	mgoSessionC, err = mgo.Dial(config.DsnMongoC)
	if err != nil {
		log.Printf("MongoDB connect failed: %s\n", err.Error())
		return false
	}

	mgoSession.SetMode(mgo.Monotonic, true)
	mgoDB = mgoSession.DB("")
	mgoOldDB = mgoSession.DB("ntrust1")
	mgoDBC = mgoSessionC.DB("")
	//mgoFS = mgoDB.GridFS("fs")

	ensureMongoIndices()
	return true
}

////////////////////////////////////////////////////////////////////////////////

func connectToRabbitMQ(uri string) *amqp.Connection {
	for {
		conn, err := amqp.Dial(uri)
		if err == nil {
			return conn
		}

		log.Println(err)
		log.Printf("Trying to reconnect to RabbitMQ at %s\n", uri)
		time.Sleep(10000 * time.Millisecond)
	}
}

func rabbitConnector(uri string) {
	var rabbitErr *amqp.Error
	var err error

	for {
		rabbitErr = <-amqpCloseError
		if rabbitErr != nil {
			amqpConn = connectToRabbitMQ(uri)
			amqpCloseError = make(chan *amqp.Error)
			amqpConn.NotifyClose(amqpCloseError)
			log.Printf("RabbitMQ DSN=%s\n", uri)

			amqpChn, err = amqpConn.Channel()
			if err != nil {
				log.Printf("AMQP Channel open failed: %v\n", err)
				//return false
			}

			amQueue_Task, err = amqpChn.QueueDeclare(
				"task", // name
				true,   // durable
				false,  // delete when unused
				false,  // exclusive
				false,  // no-wait
				nil,    // arguments
			)
			if err != nil {
				log.Printf("QueueDeclare(task) failed: %v", err)
				//return false
			}
		}
	}
}

func mqSend_Task(cmd string, ver int, data bson.M) error {
	data["cmd"] = cmd
	data["ver"] = ver
	jsonBody, err := json.Marshal(data)
	if err == nil {
		err = amqpChn.Publish(
			"",                // exchange
			amQueue_Task.Name, // routing key
			false,             // mandatory
			false,             // immediate
			amqp.Publishing{
				ContentType: "application/json",
				Body:        jsonBody,
			})
	}

	return err
}

func mqSend_Notice() error {
	data := bson.M{}
	data["cmd"] = "update"
	data["ver"] = 1
	jsonBody, err := json.Marshal(data)
	if err != nil {
		panic(err)
	}
	err = amqpChn.ExchangeDeclare(
		"notice",
		"fanout",
		false,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		panic(err)
	}
	err = amqpChn.Publish(
		"notice",
		"",
		false,
		false,
		amqp.Publishing{
			ContentType: "text/plain",
			Body:        jsonBody,
		})
	if err != nil {
		panic(err)
	}

	return err
}

////////////////////////////////////////////////////////////////////////////////

func Close() {
	if amqpChn != nil {
		amqpChn.Close()
		amqpChn = nil
	}

	if amqpConn != nil {
		amqpConn.Close()
		amqpConn = nil
	}

	if rdsPool != nil {
		rdsPool.Close()
		rdsPool = nil
	}

	if mgoSession != nil {
		mgoSession.Close()
		mgoSession = nil
	}
}

func Setup(ec *echo.Echo, config *AppConfig) bool {
	var err error

	tzLocation, err = time.LoadLocation("Asia/Seoul")
	if err != nil {
		log.Printf("TimeZone loading failed: %v\n", err)
		return false
	}

	makeBaseHostname(config)

	// Middleware

	if !config.IsDebug {
		//ec.Use(middleware.Gzip())
		ec.Use(middleware.Logger())
		ec.Use(middleware.Recover())
	}

	if !openMongoDB(config) {
		return false
	}

	amqpCloseError = make(chan *amqp.Error)
	go rabbitConnector(config.DsnAmqp)
	amqpCloseError <- amqp.ErrClosed

	// Redis
	// check https://github.com/soveran/redisurl/blob/master/redisurl.go
	log.Printf("Redis DSN=%s\n", config.DsnRedis)
	redisUrl, err := url.Parse(config.DsnRedis)
	rdsPool = redis.NewPool(func() (redis.Conn, error) {
		rdconn, err := redis.Dial("tcp", redisUrl.Host)
		if err != nil {
			return nil, err
		}

		return rdconn, err
	}, _MaxRedisConnections)
	// try connect redis
	{
		_rconn := rdsPool.Get()
		defer _rconn.Close()
		if _rconn.Err() != nil {
			log.Printf("Redis connect failed: %s\n", err.Error())
			return false
		}
	}

	store, err := redistore.NewRediStore(32, "tcp", redisUrl.Host, "", []byte("ntrust2nd"))
	if err != nil {
		log.Printf("NewRedisStore failed: %v\n", err)
		return false
	}

	ec.Use(session.Middleware(store))

	/*ec.Use(middleware.CSRFWithConfig(middleware.CSRFConfig{
		TokenLookup: "form:_csrf",
	}))*/

	ec.Use(func(h echo.HandlerFunc) echo.HandlerFunc {
		return func(ctx echo.Context) error {
			sess, err := session.Get("session", ctx)
			if err != nil {
				return err
			}

			cc := &AuthContext{ctx,
				make(map[string]interface{}),
				sess,
				"",
			}
			cc.sess.Options = &sessions.Options{
				Path:     "/",
				MaxAge:   86400, // 24 hours
				HttpOnly: true,
			}

			authUserId := cc.sess.Values[SVK_AuthUserId]
			cc.model[MVK_AuthLevel] = 0
			if authUserId != nil {
				cc.authUserId = authUserId.(string)
				cc.model[MVK_AuthUserId] = cc.authUserId
				cc.model[MVK_AuthLevel] = cc.sess.Values[SVK_AuthLevel]
				cc.model[MVK_AuthName] = cc.sess.Values[SVK_AuthName]
				cc.model[MVK_AuthPhoto] = "/public/img/noavatar.png"
			}

			_csrf := ctx.Get("csrf")
			if _csrf == nil {
				_csrf = "nil"
			}
			cc.model["_csrf"] = _csrf

			cc.model["_flSuccess"] = cc.sess.Flashes(_FLK_Success)
			cc.model["_flError"] = cc.sess.Flashes(_FLK_Error)

			defer cc.sess.Save(ctx.Request(), ctx.Response())
			return h(cc)
		}
	})

	// Create a cache with a default expiration time of 5 minutes,
	// and which purges expired items every 30 seconds
	localCache = cache.New(5*time.Minute, 30*time.Second)

	setupAuthRoutes(ec)
	setupAdminRoutes(ec)
	setupApiRoutes(ec)
	setupClusterOpenRoutes(ec)
	setupDicOpenRoutes(ec)

	ec.GET("/", func(_ctx echo.Context) error {
		ctx := _ctx.(*AuthContext)
		ctx.model["var1"] = 123
		ctx.model["var2"] = "ABC"
		return ctx.Redirect(http.StatusFound, "/api/intro")
	})

	return true
}
