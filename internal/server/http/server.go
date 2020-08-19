package http

import (
	"html/template"
	"net/http"
	"path"
	"taobaoke/internal/service"

	pb "taobaoke/api"
	"taobaoke/internal/model"

	"github.com/go-kratos/kratos/pkg/conf/paladin"
	"github.com/go-kratos/kratos/pkg/log"
	bm "github.com/go-kratos/kratos/pkg/net/http/blademaster"
)

var svr service.Server

// New new a bm server.
func New(s service.Server) (engine *bm.Engine, err error) {
	var (
		cfg struct {
			Server *bm.ServerConfig
			Client *bm.ClientConfig
		}
	)
	var (
		ct paladin.TOML
	)
	if err = paladin.Get("http.toml").Unmarshal(&ct); err != nil {
		return
	}
	if err = ct.Get("bm").UnmarshalTOML(&cfg); err != nil {
		return
	}
	svr = s
	engine = bm.DefaultServer(cfg.Server)
	pb.RegisterTBKBMServer(engine, s)
	initRouter(engine)
	err = engine.Start()
	return
}

func initRouter(e *bm.Engine) {
	e.Ping(ping)
	g := e.Group("/fileSystem")
	{
		g.GET("/start", howToStart)
	}
	e.GET(path.Join("/css", "/*filepath"), createStaticHandler("/css", http.Dir("res/css")))
	e.HEAD(path.Join("/css", "/*filepath"), createStaticHandler("/css", http.Dir("res/css")))

	e.GET(path.Join("/js", "/*filepath"), createStaticHandler("/js", http.Dir("res/js")))
	e.HEAD(path.Join("/js", "/*filepath"), createStaticHandler("/js", http.Dir("res/js")))

	e.GET("/item/:id", item)
}

func item(c *bm.Context) {
	id := c.Params.ByName("id")
	title, picURL, shopName, err := svr.QueryTitleByItemID(c, id)
	if err != nil {
		log.Error("get Title by id failed, err: %v", err)
		return
	}
	tkl, err := svr.GetTKL(c, title, picURL, id)
	if err != nil {
		log.Error("get tkl failed, err: %v", err)
		return
	}

	trendInfo, err := svr.PriceTrend(c, id)
	if err != nil {
		log.Error("get trendInfo failed, err: %v", err)
		return
	}

	t := template.Must(template.New("item.tmpl").Delims("{{", "}}").ParseFiles("./res/item.tmpl"))
	err = t.Execute(c.Writer, map[string]interface{}{
		"title":     title,
		"picURL":    picURL,
		"tkl":       tkl,
		"shopName":  shopName,
		"ngrok":     service.Ngrok,
		"trendInfo": trendInfo,
	})
	if err != nil {
		log.Error("render template failed, err: %v", err)
		return
	}
}

func createStaticHandler(absolutePath string, fs http.FileSystem) bm.HandlerFunc {
	fileServer := http.StripPrefix(absolutePath, http.FileServer(fs))
	return func(c *bm.Context) {
		file := c.Params.ByName("filepath")
		// Check if file exists and/or if we have permission to access it
		if _, err := fs.Open(file); err != nil {
			c.Writer.WriteHeader(http.StatusNotFound)
			// Reset index
			c.Abort()
			return
		}
		fileServer.ServeHTTP(c.Writer, c.Request)
	}
}
func ping(ctx *bm.Context) {
	if _, err := svr.Ping(ctx, nil); err != nil {
		log.Error("ping error(%v)", err)
		ctx.AbortWithStatus(http.StatusServiceUnavailable)
	}
}

// example for http request handler.
func howToStart(c *bm.Context) {
	k := &model.Kratos{
		Hello: "Golang 大法好 !!!",
	}
	c.JSON(k, nil)
}
