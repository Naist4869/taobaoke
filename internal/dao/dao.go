package dao

import (
	"context"
	"taobaoke/tools"
	"time"

	"go.uber.org/zap"

	"github.com/go-redis/redis/v8"

	"github.com/Naist4869/log"

	"go.mongodb.org/mongo-driver/mongo"

	"taobaoke/internal/model"

	"github.com/go-kratos/kratos/pkg/cache/memcache"
	"github.com/go-kratos/kratos/pkg/conf/paladin"
	"github.com/go-kratos/kratos/pkg/database/sql"
	"github.com/go-kratos/kratos/pkg/sync/pipeline/fanout"
	xtime "github.com/go-kratos/kratos/pkg/time"

	"github.com/google/wire"
)

var Provider = wire.NewSet(New, NewDB, NewRedis, NewMC, NewMongo, NewOrderClient, wire.Bind(new(OrderMatchService), new(Dao)))

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
	db           *sql.DB
	redis        *redis.Client
	mc           *memcache.Memcache
	mongo        *mongo.Client
	orderClient  *OrderClient
	cache        *fanout.Fanout
	logger       *log.Logger
	demoExpire   int32
	orderCacheCh chan map[string]interface{} // key -> tradeParentID:order
}

func (d *dao) UpdateOrderFailedStatus(ctx context.Context, id string, tradeParentID string) (err error) {
	err = d.orderClient.updateOrderFailedStatus(ctx, id)
	if err != nil {
		return err
	}
	if _, err = d.DelFromMatchCache(ctx, []string{tradeParentID}); err != nil {
		d.logger.Error("UpdateOrderFailedStatus", zap.Error(err), zap.String("tradeParentID", id))
		err = nil
	}
	return
}

func (d *dao) UpdateOrderPaidStatus(ctx context.Context, id string, paidTime tools.Time, AlipayTotalPrice string, IncomeRate string, pubSharePreFee string, ItemNum int) (err error) {
	return d.orderClient.updateOrderPaidStatus(ctx, id, paidTime, AlipayTotalPrice, IncomeRate, pubSharePreFee, ItemNum)
}

func (d *dao) UpdateManyWithDrawStatus(ctx context.Context, ids []string) (err error) {
	return d.orderClient.updateManyWithDrawStatus(ctx, ids)
}

func (d *dao) FindOneAndUpdateOrderBalanceStatus(ctx context.Context, id string, tradeParentID string, earningTime tools.Time, totalCommissionFee string, PayPrice string, salaryScale int64) (order *model.Order, err error) {
	order, err = d.orderClient.findOneAndUpdateOrderBalanceStatus(ctx, id, earningTime, totalCommissionFee, PayPrice, salaryScale)
	if err != nil {
		return
	}
	if _, err = d.DelFromMatchCache(ctx, []string{tradeParentID}); err != nil {
		d.logger.Error("UpdateOrderBalanceStatus", zap.Error(err), zap.String("tradeParentID", id))
		err = nil
	}
	return
}

func (d *dao) Insert(ctx context.Context, o *model.Order) (err error) {
	o.UpdateTime = tools.Now()
	return d.orderClient.Insert(ctx, o)
}

func (d *dao) FindOrderByID(ctx context.Context, id string) (order *model.Order, err error) {
	orders, err := d.orderClient.FindOrderByIDs(ctx, []string{id})
	if err != nil {
		return
	}
	return orders[0], nil
}

func (d *dao) FindOrderByIDs(ctx context.Context, ids []string) (order []*model.Order, err error) {
	return d.orderClient.FindOrderByIDs(ctx, ids)
}

func (d *dao) FindOrderByTradeParentID(ctx context.Context, ids []string) (order []*model.Order, err error) {
	return d.orderClient.FindOrderByTradeParentID(ctx, ids)
}

// New new a dao and return.
func New(logger *log.Logger, r *redis.Client, mc *memcache.Memcache, db *sql.DB, mongo *mongo.Client, orderClient *OrderClient) (d Dao, cf func(), err error) {
	return newDao(logger, r, mc, db, mongo, orderClient)
}

func newDao(logger *log.Logger, r *redis.Client, mc *memcache.Memcache, db *sql.DB, mongo *mongo.Client, orderClient *OrderClient) (d *dao, cf func(), err error) {
	var cfg struct {
		DemoExpire xtime.Duration
	}
	if err = paladin.Get("application.toml").UnmarshalTOML(&cfg); err != nil {
		return
	}
	d = &dao{
		db:           db,
		redis:        r,
		mc:           mc,
		mongo:        mongo,
		orderClient:  orderClient,
		cache:        fanout.New("cache", fanout.Worker(10), fanout.Buffer(10240)),
		logger:       logger,
		demoExpire:   int32(time.Duration(cfg.DemoExpire) / time.Second),
		orderCacheCh: make(chan map[string]interface{}, 10240),
	}
	cf = d.Close
	go d.setOrderCache()
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

func (d *dao) setOrderCache() {
	for {
		missCache, open := <-d.orderCacheCh
		if !open {
			d.logger.Info("设置缓存", zap.String("错误", "通道已关闭"))
			break
		}
		for key, m := range missCache {
			if n, err := d.redis.HSet(context.Background(), key, m).Result(); err != nil {
				d.logger.Warn("setOrderCache", zap.Error(err), zap.Int64("n", n))
			}
		}
		// 防止redis阻塞
		time.Sleep(time.Millisecond * 1)
	}
}
