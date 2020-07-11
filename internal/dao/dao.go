package dao

import (
	"context"
	"time"

	"github.com/Naist4869/log"

	"go.mongodb.org/mongo-driver/mongo"

	"taobaoke/internal/model"

	"github.com/go-kratos/kratos/pkg/cache/memcache"
	"github.com/go-kratos/kratos/pkg/cache/redis"
	"github.com/go-kratos/kratos/pkg/conf/paladin"
	"github.com/go-kratos/kratos/pkg/database/sql"
	"github.com/go-kratos/kratos/pkg/sync/pipeline/fanout"
	xtime "github.com/go-kratos/kratos/pkg/time"

	"github.com/google/wire"
)

var Provider = wire.NewSet(New, NewDB, NewRedis, NewMC, NewMongo, NewOrderClient)

//go:generate kratos tool genbts
// Dao dao interface
type Dao interface {
	Close()
	Ping(ctx context.Context) (err error)
	// bts: -nullcache=&model.Article{ID:-1} -check_null_code=$!=nil&&$.ID==-1
	Article(c context.Context, id int64) (*model.Article, error)
	OrderDataService
}

// dao dao.
type dao struct {
	db          *sql.DB
	redis       *redis.Redis
	mc          *memcache.Memcache
	mongo       *mongo.Client
	orderClient *OrderClient
	cache       *fanout.Fanout
	logger      *log.Logger
	demoExpire  int32
}

func (d *dao) Insert(ctx context.Context, o *model.Order) (err error) {
	return d.orderClient.Insert(ctx, o)
}

func (d *dao) FindOrderByID(ctx context.Context, id string) (order *model.Order, err error) {
	return d.orderClient.FindOrderByID(ctx, id)
}

// New new a dao and return.
func New(logger *log.Logger, r *redis.Redis, mc *memcache.Memcache, db *sql.DB, mongo *mongo.Client, orderClient *OrderClient) (d Dao, cf func(), err error) {
	return newDao(logger, r, mc, db, mongo, orderClient)
}

func newDao(logger *log.Logger, r *redis.Redis, mc *memcache.Memcache, db *sql.DB, mongo *mongo.Client, orderClient *OrderClient) (d *dao, cf func(), err error) {
	var cfg struct {
		DemoExpire xtime.Duration
	}
	if err = paladin.Get("application.toml").UnmarshalTOML(&cfg); err != nil {
		return
	}
	d = &dao{
		db:          db,
		redis:       r,
		mc:          mc,
		mongo:       mongo,
		orderClient: orderClient,
		cache:       fanout.New("cache"),
		logger:      logger,
		demoExpire:  int32(time.Duration(cfg.DemoExpire) / time.Second),
	}
	cf = d.Close
	return
}

// Close close the resource.
func (d *dao) Close() {
	d.cache.Close()
}

// Ping ping the resource.
func (d *dao) Ping(ctx context.Context) (err error) {
	return d.mongo.Ping(ctx, nil)
}
