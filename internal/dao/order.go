package dao

import (
	"context"
	"math"
	"strconv"
	"strings"
	pb "taobaoke/api"
	"taobaoke/internal/dao/gdbc"
	"taobaoke/internal/model"
	"taobaoke/tools"

	"github.com/go-redis/redis/v8"

	"github.com/golang/protobuf/ptypes/empty"
	"google.golang.org/grpc"

	"github.com/go-kratos/kratos/pkg/conf/paladin"

	"github.com/pkg/errors"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"

	"go.uber.org/zap"

	"github.com/Naist4869/log"
	"go.mongodb.org/mongo-driver/mongo"
)

var (
	defaultSort = []string{"_id"}
)

type OrderDataService interface {
	OrderMatchService
	OrderMonitor
	QueryOrderByStatus(ctx context.Context, start, end tools.Time, status ...model.OrderStatus) ([]*model.Order, error)
	QueryNotWithDrawOrderByUserID(ctx context.Context, id string) (result []*model.Order, err error)
	UpdateSingleOrderGeneric(ctx context.Context, id string, additionalFilter, operation bson.M) (err error)
}
type OrderMonitor interface {
	UpdateStatus(ctx context.Context, localOrder *model.Order, fill model.Fill, salaryScale int64)
	DelFromMatchCache(ctx context.Context, tradeParentIDs ...string) (n int64, err error)
	REMFromMatchSet(ctx context.Context, tradeParentID ...string) (n int64, err error)
	FindOrderByID(ctx context.Context, id string) (order *model.Order, err error)
	UpdateOrderFailedStatus(ctx context.Context, id string, tradeParentID string) (err error)
	UpdateOrderPaidStatus(ctx context.Context, id string, paidTime tools.Time, AlipayTotalPrice string, IncomeRate string, pubSharePreFee string, ItemNum int) (err error)
	UpdateManyWithDrawStatus(ctx context.Context, ids []string) (err error)
	UpdateOrderBalanceStatus(ctx context.Context, id string, tradeParentID string, earningTime tools.Time, totalCommissionFee string, salaryScale int64) (err error)
	UpdateOrderFinishStatus(ctx context.Context, id string, payPrice string) (err error)
}

type CacheObjService interface {
	FindOrderByTradeParentID(ctx context.Context, ids []string) (order []*model.Order, err error)
	FindOrderByIDs(ctx context.Context, ids []string) (order []*model.Order, err error)
}

type OrderMatchService interface {
	MatchedTemplateMsgSend(ctx context.Context, in *pb.MatchedTemplateMsgSendReq, opts ...grpc.CallOption) (*empty.Empty, error)
	Insert(ctx context.Context, o *model.Order) (err error)
	QueryOrderByTradeParentID(ctx context.Context, ids []string, onlyUnfinished bool) (results []*model.Order, err error)
	SetNXToUnmatch(ctx context.Context, itemID, adZoneID int64, orderNo string) (ok bool, err error)
	SetToUnmatch(ctx context.Context, itemID, adZoneID int64, order *model.Order, nonce string) (ok bool, err error)
	UpdateFromUnmatch(ctx context.Context, itemID, adZoneID int64, order *model.Order) (ok bool, err error)
	GetUnmatch(ctx context.Context, itemID, adZoneID int64) (*model.Order, error)
	GetAllUnmatch(ctx context.Context) (map[string]*model.Order, error)
	MatchGetAll(ctx context.Context) ([]*model.Order, error)
	ExistInUnmatch(ctx context.Context, itemID, adZoneID int64) (exist bool, err error)
	DelFromUnmatchAndSetToMatch(ctx context.Context, order *model.Order) (ok bool, err error)
	PSubscribeKeyspace() <-chan *redis.Message // 到时候做转发 避免依赖
	HSetNXToMatch(ctx context.Context, order *model.Order) (ok bool, err error)
}

type OrderClient struct {
	collections map[string]*mongo.Collection // 数据表map
	Logger      *log.Logger
}

func NewOrderClient(client *mongo.Client, logger *log.Logger) (*OrderClient, error) {
	var (
		cfg mongoConfig
		ct  paladin.TOML
	)
	if err := paladin.Get("db.toml").Unmarshal(&ct); err != nil {
		return nil, err
	}
	if err := ct.Get("Mongo").UnmarshalTOML(&cfg); err != nil {
		return nil, err
	}
	oc := &OrderClient{
		collections: make(map[string]*mongo.Collection, 3),
		Logger:      logger.With(zap.String("组件", "order")),
	}

	//if !tools.IsDebug() {
	if err := oc.Init(client, cfg.DB); err != nil {
		return nil, err
	}
	//}

	return oc, nil
}
func (oc *OrderClient) Keys() map[string]*gdbc.Spec {
	specs := make(map[string]*gdbc.Spec, 1)
	orderSpec, err := gdbc.NewSpec(model.DBOrderVersion, func() interface{} {
		return &model.Order{}
	}, func(data interface{}) error {
		return nil
	})
	if err != nil {
		oc.Logger.Fatal("构建updater失败:", zap.Error(err))
	}
	specs[model.DBOrderKey] = orderSpec
	return specs
}

func (oc *OrderClient) Init(client *mongo.Client, db string) error {

	keys := oc.Keys()
	for key, spec := range keys {
		collection := client.Database(db).Collection(key)
		oc.collections[key] = collection
		if spec != nil {
			spec.SetCollection(collection)
		}
		if err := gdbc.EnsureIndex(collection, []gdbc.Index{
			{
				Name: "订单编号",
				Data: mongo.IndexModel{
					Keys:    bson.M{model.TradeParentIDField: 1},
					Options: options.Index().SetUnique(true).SetPartialFilterExpression(bson.M{model.DeletedField: false}),
				},
				Version: 1,
			},
		}, oc.Logger); err != nil {
			return errors.Wrap(err, "检查并创建索引失败")
		}
	}
	updater, err := gdbc.NewUpdater(keys, oc.Logger)
	if err != nil {
		return errors.Wrap(err, "构建collection升级器失败")
	}
	if err := updater.Update(); err != nil {
		return errors.Wrap(err, "检查并升级collection数据失败")
	}
	return nil
}

func (oc *OrderClient) Insert(ctx context.Context, o *model.Order) (err error) {
	if _, err = oc.collections[model.DBOrderKey].InsertOne(ctx, *o); err != nil {
		if IsInsertDuplicateError(err) {
			errMsg := err.Error()
			strings.Contains(errMsg, "index: 订单编号")
			err = errors.New("订单编号已存在")
			return
		}
		err = errors.New("保存订单信息失败")
		return
	}
	return
}

func (oc *OrderClient) FindOrderByIDs(ctx context.Context, ids []string) (orders []*model.Order, err error) {
	if len(ids) == 0 {
		err = NewErrIDNotFound(ids)
		return
	}

	query := bson.M{
		model.IDField: bson.M{IN: ids},
	}
	orders, _, err = oc.queryOrder(ctx, query, nil, 0, 0, nil, nil)
	if err != nil {
		oc.Logger.Error("FindOrderByIDs", zap.Error(err), ZapBsonM("query", query))
		err = errors.New("根据id获取订单失败")
		return
	}
	if len(orders) == 0 {
		err = NewErrIDNotFound(ids)
		return
	}
	return
}

func (oc *OrderClient) FindOrderByTradeParentID(ctx context.Context, tradeParentID []string) (orders []*model.Order, err error) {
	query := bson.M{
		model.TradeParentIDField: bson.M{IN: tradeParentID},
	}

	orders, _, err = oc.queryOrder(ctx, query, nil, 0, 0, nil, nil)
	if err != nil {
		oc.Logger.Error("FindOrderByTradeParentID", zap.Error(err), ZapBsonM("query", query))
		err = errors.New("根据淘宝订单号获取订单失败")
		return
	}
	if len(orders) == 0 {
		err = NewErrTradeParentIDNotFound(tradeParentID)
		return
	}
	return
}
func (oc *OrderClient) updateOrderFinishStatus(ctx context.Context, id string, PayPrice string) (err error) {
	payPrice, err := strconv.ParseFloat(PayPrice, 64)
	if err != nil {
		oc.Logger.Error("updateOrderFinishStatus", zap.Error(err))
		return
	}
	filter := bson.M{
		model.IDField:      id,
		model.DeletedField: false,
	}
	operation := bson.M{
		SET: bson.M{
			model.StatusField:   model.OrderFinish,
			model.PayPriceField: int64(payPrice * 100),
		},
		ADDTOSET: bson.M{model.TimelinesField: model.NewTimeLine(-1, model.OrderFinish.String())},
	}
	updateResult, err := oc.collections[model.DBOrderKey].UpdateOne(ctx, filter, operation)
	if err != nil {
		oc.Logger.Error("updateOrderFinishStatus", zap.Error(err), ZapBsonM("filter", filter))
		return
	}
	if updateResult.ModifiedCount == 0 {
		return NewErrIDNotFound([]string{id})
	}
	oc.Logger.Info("updateOrderFinishStatus", zap.String("id", id), zap.Int64("匹配数量", updateResult.MatchedCount), zap.Int64("更新数量", updateResult.ModifiedCount))
	return
}
func (oc *OrderClient) findOneAndUpdateOrderBalanceStatus(ctx context.Context, id string, earningTime tools.Time, totalCommissionFee string, salaryScale int64) (order *model.Order, err error) {
	commission, err := strconv.ParseFloat(totalCommissionFee, 64)
	if err != nil {
		oc.Logger.Error("findOneAndUpdateOrderBalanceStatus", zap.Error(err))
		return
	}

	filter := bson.M{
		model.IDField:      id,
		model.DeletedField: false,
	}
	operation := bson.M{
		SET: bson.M{
			model.StatusField:      model.OrderBalance,
			model.CommissionField:  int64(commission * 100),
			model.SalaryScaleField: salaryScale,
			model.SalaryField:      int64(commission*100) * salaryScale / 100,
			model.EarningTimeField: earningTime,
		},
		ADDTOSET: bson.M{model.TimelinesField: model.NewTimeLine(-1, model.OrderBalance.String())},
	}
	option := options.FindOneAndUpdate().SetReturnDocument(options.After)
	updateResult := oc.collections[model.DBOrderKey].FindOneAndUpdate(ctx, filter, operation, option)
	err = updateResult.Err()
	if err != nil {
		oc.Logger.Error("findOneAndUpdateOrderBalanceStatus", zap.Error(err), ZapBsonM("filter", filter))
		return
	}
	v := &model.Order{}
	err = updateResult.Decode(v)
	if err != nil {
		oc.Logger.Error("findOneAndUpdateOrderBalanceStatus", zap.Error(err), ZapBsonM("filter", filter))
		return
	}
	oc.Logger.Info("findOneAndUpdateOrderBalanceStatus", zap.String("id", id), zap.Int64("匹配数量", 1), zap.Int64("更新数量", 1))
	return v, nil

}

func (oc *OrderClient) updateManyWithDrawStatus(ctx context.Context, ids []string) (err error) {
	if len(ids) == 0 {
		return
	}
	filter := bson.M{
		model.IDField:      bson.M{IN: ids},
		model.DeletedField: false,
	}
	operation := bson.M{
		SET: bson.M{
			model.WithDrawStatusField: true,
		},
		ADDTOSET: bson.M{model.TimelinesField: model.NewTimeLine(-1, "已提现")},
	}

	updateResult, err := oc.collections[model.DBOrderKey].UpdateMany(ctx, filter, operation)
	if err != nil {
		oc.Logger.Error("UpdateWithDrawStatus", zap.Error(err), ZapBsonM("filter", filter))
		return
	}
	oc.Logger.Info("UpdateWithDrawStatus", zap.Strings("ids", ids), zap.Int64("匹配数量", updateResult.MatchedCount), zap.Int64("更新数量", updateResult.ModifiedCount))
	return
}

func (oc *OrderClient) updateOrderFailedStatus(ctx context.Context, id string) (err error) {
	filter := bson.M{
		model.IDField:      id,
		model.DeletedField: false,
	}
	operation := bson.M{
		SET: bson.M{
			model.StatusField: model.OrderFailed,
		},
		ADDTOSET: bson.M{model.TimelinesField: model.NewTimeLine(-1, model.OrderFailed.String())},
	}
	updateResult, err := oc.collections[model.DBOrderKey].UpdateOne(ctx, filter, operation)
	if err != nil {
		oc.Logger.Error("UpdateOrderFailedStatus", zap.Error(err), ZapBsonM("filter", filter))
		return
	}
	if updateResult.ModifiedCount == 0 {
		return NewErrIDNotFound([]string{id})
	}
	oc.Logger.Info("UpdateOrderFailedStatus", zap.String("id", id), zap.Int64("匹配数量", updateResult.MatchedCount), zap.Int64("更新数量", updateResult.ModifiedCount))
	return
}

func (oc *OrderClient) updateOrderPaidStatus(ctx context.Context, id string, paidTime tools.Time, AlipayTotalPrice string, IncomeRate string, pubSharePreFee string, ItemNum int) (err error) {

	alipayTotalPrice, err := strconv.ParseFloat(AlipayTotalPrice, 64)
	if err != nil {
		oc.Logger.Error("UpdateOrderPaidStatus", zap.Error(err))
		return err
	}
	incomeRate, err := strconv.ParseFloat(IncomeRate, 64)
	if err != nil {
		oc.Logger.Error("UpdateOrderPaidStatus", zap.Error(err))
		return err
	}
	rebate, err := strconv.ParseFloat(pubSharePreFee, 64)
	if err != nil {
		oc.Logger.Error("UpdateOrderPaidStatus", zap.Error(err))
		return err
	}
	calculateCommission := alipayTotalPrice * incomeRate / 100
	// 保留两位小数四舍五入
	roundCommission := math.Round(calculateCommission*100) / 100
	if roundCommission != rebate {
		err = model.NewCalculateCommissionInconsistentError(id, rebate, roundCommission, calculateCommission)
		oc.Logger.Error("UpdateOrderPaidStatus", zap.Error(err))
		err = nil
	}
	filter := bson.M{
		model.IDField:      id,
		model.DeletedField: false,
	}
	operation := bson.M{
		SET: bson.M{
			model.AlipayTotalPriceField: int64(alipayTotalPrice * 100),
			model.PaidTimeField:         paidTime,
			model.StatusField:           model.OrderPaid,
			model.CountField:            ItemNum,
		},
		ADDTOSET: bson.M{model.TimelinesField: model.NewTimeLine(-1, model.OrderPaid.String())},
	}
	updateResult, err := oc.collections[model.DBOrderKey].UpdateOne(ctx, filter, operation)
	if err != nil {
		oc.Logger.Error("UpdateOrderPaidStatus", zap.Error(err), ZapBsonM("filter", filter))
		return
	}
	if updateResult.ModifiedCount == 0 {
		return NewErrIDNotFound([]string{id})
	}
	oc.Logger.Info("UpdateOrderPaidStatus", zap.String("id", id), zap.Int64("匹配数量", updateResult.MatchedCount), zap.Int64("更新数量", updateResult.ModifiedCount))
	return
}

func (oc *OrderClient) queryOrder(ctx context.Context, query bson.M, sort []string, start, limit int64, include, exclude []string, collations ...*options.Collation) (records []*model.Order, count int64, err error) {
	var (
		data    = &model.Order{}
		results []interface{}
	)

	if len(sort) == 0 {
		sort = defaultSort
	}
	if len(query) == 0 {
		query = make(bson.M, 1)
	}
	query[model.DeletedField] = false // 查询未删除的数据
	results, count, err = baseQuery(oc.collections[model.DBOrderKey], ctx, query, sort, start, limit, include, exclude, data, collations...)
	if err != nil {
		return
	}
	records = make([]*model.Order, 0, len(results))
	for _, data := range results {
		records = append(records, data.(*model.Order))
	}
	return
}
