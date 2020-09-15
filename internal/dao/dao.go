package dao

import (
	"context"
	"reflect"
	"strconv"
	"sync"
	pb "taobaoke/api"
	"taobaoke/tools"
	"time"

	"github.com/go-kratos/kratos/pkg/net/rpc/warden"
	"github.com/golang/protobuf/ptypes/empty"
	"google.golang.org/grpc"

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

var statues = []model.OrderStatus{model.OrderIllegal, model.OrderFailed, model.OrderCreate, model.OrderPaid, model.OrderFinish, model.OrderBalance}

// dao dao.
type dao struct {
	db           *sql.DB
	redis        *redis.ClusterClient
	mc           *memcache.Memcache
	mongo        *mongo.Client
	orderClient  *OrderClient
	cache        *fanout.Fanout
	client       pb.WechatClient
	once         sync.Once
	onceFun      func()
	logger       *log.Logger
	pool         sync.Pool
	statusesMap  tools.OrderedMap
	demoExpire   int32
	orderCacheCh chan map[string]interface{} // key -> tradeParentID:order
}

func (d *dao) MatchedTemplateMsgSend(ctx context.Context, in *pb.MatchedTemplateMsgSendReq, opts ...grpc.CallOption) (*empty.Empty, error) {
	d.once.Do(d.onceFun)
	defer func() {
		err := recover()
		if err != nil {
			d.logger.Error("MatchedTemplateMsgSend recover", zap.Error(err.(error)))
		}
	}()
	return d.client.MatchedTemplateMsgSend(ctx, in, opts...)
}

func (d *dao) BalanceTemplateMsgSend(ctx context.Context, in *pb.BalanceTemplateMsgSendReq, opts ...grpc.CallOption) (*empty.Empty, error) {
	d.once.Do(d.onceFun)
	defer func() {
		err := recover()
		if err != nil {
			d.logger.Error("BalanceTemplateMsgSend recover", zap.Error(err.(error)))
		}
	}()
	return d.client.BalanceTemplateMsgSend(ctx, in, opts...)
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
func (d *dao) UpdateOrderFinishStatus(ctx context.Context, id string, payPrice string) (err error) {
	return d.orderClient.updateOrderFinishStatus(ctx, id, payPrice)

}
func (d *dao) UpdateOrderBalanceStatus(ctx context.Context, id string, tradeParentID string, earningTime tools.Time, totalCommissionFee string, salaryScale int64) (err error) {
	afterOrder, err := d.orderClient.findOneAndUpdateOrderBalanceStatus(ctx, id, earningTime, totalCommissionFee, salaryScale)
	if err != nil {
		return
	}
	if _, err = d.DelFromMatchCache(ctx, []string{tradeParentID}); err != nil {
		d.logger.Error("UpdateOrderBalanceStatus", zap.Error(err), zap.String("tradeParentID", tradeParentID))
		err = nil
	}
	//todo 暂时先这样 没做用户钱包
	notWithDraw, err := d.QueryNotWithDrawOrderByUserID(ctx, afterOrder.UserID)
	if err != nil {
		d.logger.Error("UpdateOrderBalanceStatus", zap.Error(err), zap.String("ID", id))
		return
	}

	var balance int64
	for _, order := range notWithDraw {
		balance += order.Salary
	}

	if _, err = d.BalanceTemplateMsgSend(ctx, &pb.BalanceTemplateMsgSendReq{
		UserID:      afterOrder.UserID,
		OrderID:     afterOrder.TradeParentID,
		Title:       afterOrder.Title,
		EarningTime: afterOrder.EarningTime.String(),
		Salary:      strconv.FormatFloat(float64(afterOrder.Salary)/100, 'f', -1, 64),
		Balance:     strconv.FormatFloat(float64(balance)/100, 'f', -1, 64),
	}); err != nil {
		d.logger.Error("UpdateOrderBalanceStatus", zap.Error(err), zap.String("ID", id))
		err = nil
	}
	return
}

func (d *dao) Insert(ctx context.Context, o *model.Order) (err error) {
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
func New(logger *log.Logger, r *redis.ClusterClient, mc *memcache.Memcache, db *sql.DB, mongo *mongo.Client, orderClient *OrderClient) (d Dao, cf func(), err error) {
	return newDao(logger, r, mc, db, mongo, orderClient)
}

func newDao(logger *log.Logger, r *redis.ClusterClient, mc *memcache.Memcache, db *sql.DB, mongo *mongo.Client, orderClient *OrderClient) (d *dao, cf func(), err error) {
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
	d.onceFun = func() {
		//  现在是相互依赖的 只能懒加载 不然服务都跑不起来 应该建个代理转发请求解耦
		// 10秒重试一次 超过3分钟连不上就这样吧.. 因为在once里不会再有第二次赋值了  下面只要执行一次就会panic一次空指针
		err := tools.Retry(func() (err error, mayRetry bool) {
			d.client, err = pb.NewClient(&warden.ClientConfig{Timeout: xtime.Duration(time.Second)})
			if err == context.DeadlineExceeded {
				return nil, true
			}
			return
		})
		d.logger.Error("TemplateMsgSend once.Do", zap.Error(err))
	}

	d.pool.New = func() interface{} {
		return d.allocateContext()
	}
	{
		var statusesMap = tools.NewOrderedMap(tools.NewKeys(func(i interface{}, j interface{}) int8 {
			if i.(model.OrderStatus) == j.(model.OrderStatus) {
				return 0
			}
			var Ifinded, Jfinded int8
			for _, status := range statues {
				// 先找到的肯定比后找到的小
				switch status {
				case i.(model.OrderStatus):
					Ifinded += 1
				case j.(model.OrderStatus):
					Jfinded += 1
				default:
					continue
				}
				return Jfinded - Ifinded
			}
			return 0
		}, reflect.TypeOf(model.OrderCreate)), reflect.TypeOf(HandlerFunc(nil)))
		// OrderIllegal 对应的方法永远取不到  因为是左开右闭区间
		statusesMap.Put(model.OrderIllegal, HandlerFunc(func(c *Context) {
			c.logger.Error("Context", zap.String("原因", "进入了OrderIllegal更新方法"), zap.Any("更新字段", c.updateArg), zap.Any("本地订单", c.localOrder))
			return
		}))
		statusesMap.Put(model.OrderFailed, HandlerFunc(func(c *Context) {
			err := c.engine.UpdateOrderFailedStatus(c.ctx, c.localOrder.ID, c.localOrder.TradeParentID)
			if err != nil {
				c.logger.Error("Context", zap.Error(err), zap.Any("更新字段", c.updateArg), zap.Any("本地订单", c.localOrder))
				c.Abort()
			}
			c.logger.Info("Context", zap.String("原因", "发现已失效的单"), zap.Any("更新字段", c.updateArg), zap.Any("本地订单", c.localOrder))

			return
		}))
		statusesMap.Put(model.OrderPaid, HandlerFunc(func(c *Context) {
			err := c.engine.UpdateOrderPaidStatus(c.ctx, c.localOrder.ID, c.updateArg.PaidTime, c.updateArg.AlipayTotalPrice, c.updateArg.IncomeRate, c.updateArg.PubSharePreFee, c.updateArg.ItemNum)
			if err != nil {
				c.logger.Error("Context", zap.Error(err), zap.Any("更新字段", c.updateArg), zap.Any("本地订单", c.localOrder))
				c.Abort()
			}
			c.logger.Info("Context", zap.String("原因", "发现已付款的单"), zap.Any("更新字段", c.updateArg), zap.Any("本地订单", c.localOrder))
			return
		}))
		statusesMap.Put(model.OrderFinish, HandlerFunc(func(c *Context) {
			err := c.engine.UpdateOrderFinishStatus(c.ctx, c.localOrder.ID, c.updateArg.PayPrice)
			// 暂时没看到finish的单
			if err != nil {
				c.logger.Error("Context", zap.Error(err), zap.Any("更新字段", c.updateArg), zap.Any("本地订单", c.localOrder))
				c.Abort()
			}
			c.logger.Info("Context", zap.String("原因", "发现已完成的单"), zap.Any("更新字段", c.updateArg), zap.Any("本地订单", c.localOrder))
			return
		}))
		statusesMap.Put(model.OrderBalance, HandlerFunc(func(c *Context) {

			err := c.engine.UpdateOrderBalanceStatus(c.ctx, c.localOrder.ID, c.localOrder.TradeParentID, c.updateArg.EarningTime, c.updateArg.TotalCommissionFee, c.SalaryScale)
			if err != nil {
				c.logger.Error("Context", zap.Error(err), zap.Any("更新字段", c.updateArg), zap.Any("本地订单", c.localOrder))
				c.Abort()
			}
			c.logger.Info("Context", zap.String("原因", "发现已结算的单"), zap.Any("更新字段", c.updateArg), zap.Any("本地订单", c.localOrder))

			return
		}))
		statusesMap.Put(model.OrderCreate, HandlerFunc(func(c *Context) {
			c.logger.Error("Context", zap.String("原因", "进入了OrderCreate更新方法"), zap.Any("更新字段", c.updateArg), zap.Any("本地订单", c.localOrder))
			return
		}))
		d.statusesMap = statusesMap
	}
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
func (d *dao) allocateContext() *Context {
	return &Context{
		engine: d,
		logger: d.logger.With(zap.String("组件", "Context")),
	}
}

func (d *dao) UpdateStatus(ctx context.Context, localOrder *model.Order, fill model.Fill, salaryScale int64) {
	c := d.pool.Get().(*Context)
	c.reset()
	c.updateArg = fill.FillContext()
	c.localOrder = localOrder
	if c.updateArg.TradeParentID != c.localOrder.TradeParentID {
		return
	}
	c.ctx = ctx
	c.SalaryScale = salaryScale
	subMap := d.statusesMap.SubMap(localOrder.Status, c.updateArg.Status)
	// statusesMap不是并发安全的 但是只查不改没有问题
	for _, fun := range subMap.Elems() {
		c.handlers = append(c.handlers, fun.(HandlerFunc))
	}
	c.Next()
	d.pool.Put(c)
}
