package http

import (
	"html/template"
	"net/http"
	"path"
	"strconv"
	"taobaoke/internal/service"
	"taobaoke/tools"

	"github.com/go-kratos/kratos/pkg/net/http/blademaster/binding"

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
	g := e.Group("/taobaoke")
	{
		g.GET("/start", howToStart)
	}
	e.GET(path.Join("/css", "/*filepath"), createStaticHandler("/css", http.Dir("res/css")))
	e.HEAD(path.Join("/css", "/*filepath"), createStaticHandler("/css", http.Dir("res/css")))

	e.GET(path.Join("/js", "/*filepath"), createStaticHandler("/js", http.Dir("res/js")))
	e.HEAD(path.Join("/js", "/*filepath"), createStaticHandler("/js", http.Dir("res/js")))

	e.GET("/item", item)
}

func item(c *bm.Context) {
	v := new(struct {
		ID       string `form:"id" binding:"required"`
		ItemID   string `form:"itemID" binding:"required"`
		AdZoneID string `form:"adZoneID" binding:"required"`
	})

	if err := c.BindWith(v, binding.Query); err != nil {
		http.NotFound(c.Writer, c.Request)
		return
	}

	itemID, err := strconv.ParseInt(v.ItemID, 10, 64)
	if err != nil {
		http.NotFound(c.Writer, c.Request)
		return
	}
	adZoneID, err := strconv.ParseInt(v.AdZoneID, 10, 64)
	if err != nil {
		http.NotFound(c.Writer, c.Request)
		return
	}
	order, err := svr.UnmatchGet(c, itemID, adZoneID)
	if err != nil {
		http.NotFound(c.Writer, c.Request)
		return
	}
	if order.ID != v.ID {
		http.NotFound(c.Writer, c.Request)
		return
	}
	if !order.TrendInfo.EffectiveDate.SameDay(tools.Now()) {
		log.Info("重新设置淘口令,商品ID: %s,原口令:%s", order.ID, order.TrendInfo.TKL)
		if trendInfo, err := svr.PriceTrend(c, itemID); err != nil {
			log.Error("item get trendInfo error: %+v", err)
		} else {
			order.TrendInfo = trendInfo
		}
		if tkl, _, _, err := svr.GetTklByItemID(c, order.ItemID, order.AdzoneID, order.Title); err != nil {
			log.Error("item get tkl failed, err: %v,title: %s, picURL: %s, adZoneID: %d", err, order.Title, order.PicURL, order.AdzoneID)
		} else {
			order.TrendInfo.TKL = tkl
			log.Info("重新设置淘口令,商品ID: %s,新口令:%s", order.ID, order.TrendInfo.TKL)
		}
		_, _ = svr.UpdateToUnmatch(c, order.ItemID, order.AdzoneID, order)
	}
	t := template.Must(template.New("item.tmpl").Delims("{{", "}}").ParseFiles("./res/item.tmpl"))
	err = t.Execute(c.Writer, map[string]interface{}{
		"title":      order.Title,
		"picURL":     order.PicURL,
		"tkl":        order.TrendInfo.TKL,
		"shopName":   order.ShopName,
		"serverAddr": template.URL(svr.GetServerAddr()),
		"trendInfo":  order.TrendInfo,
	})
	if err != nil {
		log.Error("render template failed, err: %v", err)
		http.NotFound(c.Writer, c.Request)
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
