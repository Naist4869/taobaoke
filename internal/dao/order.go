package dao

import (
	"context"
	"fmt"
	"strings"
	"taobaoke/internal/dao/gdbc"
	"taobaoke/internal/model"
	"taobaoke/tools"

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
	FindOrderByID(ctx context.Context, id int64) (order *model.Order, err error)
	FindOrderByIDs(ctx context.Context, ids []int64) (order []*model.Order, err error)
	FindOrderByTradeParentID(ctx context.Context, ids []string) (order []*model.Order, err error)
	FindOneAndUpdateCommission(ctx context.Context, id string, order *model.Order) (err error)
	OrderMatchService
}

type OrderMatchService interface {
	Insert(ctx context.Context, o *model.Order) (err error)
	SetNXToUnmatch(ctx context.Context, itemID, adZoneID int64, orderNo string) (ok bool, err error)
	SetToUnmatch(ctx context.Context, itemID, adZoneID int64, order *model.Order) (ok bool, err error)
	UnmatchGet(ctx context.Context, itemID, adZoneID int64) (*model.Order, error)
	UnmatchGetAll(ctx context.Context) ([]string, error)
	ExistInUnmatch(ctx context.Context, itemID, adZoneID int64) (exist bool, err error)
	DelFromUnmatchMap(ctx context.Context, itemID, adZoneID int64) (int64, error)
	QueryOrderByTradeParentID(ctx context.Context, ids []string, onlyUnfinished bool) (results []*model.Order, err error)
	QueryOrderByStatus(ctx context.Context, start, end tools.Time, status ...model.OrderStatus) ([]*model.Order, error)
	QueryNotGiveSalaryOrderByUserID(ctx context.Context, id string) (result []*model.Order, err error)
	UpdateSingleOrderGeneric(ctx context.Context, id string, additionalFilter, operation bson.M) (err error)
	DelFromOrderCache(ctx context.Context, tradeParentIDs []string) (n int64, err error)
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
				Name: "用户ID",
				Data: mongo.IndexModel{
					Keys:    bson.M{model.UserIDField: 1},
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
			strings.Contains(errMsg, "index: 用户ID")
			err = errors.New("用户ID已存在")
			return
		}
		err = errors.New("保存订单信息失败")
		return
	}
	return
}

func (oc *OrderClient) FindOrderByIDs(ctx context.Context, ids []int64) (orders []*model.Order, err error) {
	query := bson.M{
		model.IDField: bson.M{IN: ids},
	}
	orders, _, err = oc.queryOrder(ctx, query, nil, 0, 0, nil, nil)
	if err != nil {
		oc.Logger.Error("FindOrderByID", zap.Error(err), ZapBsonM("query", query))
		err = errors.New("根据id获取订单失败")
		return
	}
	if len(orders) == 0 {
		err = fmt.Errorf("未找到ID为%d的订单", ids)
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
func (d *dao) FindOneAndUpdateCommission(ctx context.Context, id string, order *model.Order) (err error) {
	filter := bson.M{
		model.IDField:      id,
		model.DeletedField: false,
		model.SalaryField:  bson.M{LTE: 0},
	}
	update := bson.M{
		SET: bson.M{
			model.CommissionField:  order.Commission,
			model.SalaryField:      order.Salary,
			model.EarningTimeField: order.EarningTime,
			model.StatusField:      order.Status,
			model.UpdateTimeField:  tools.Now(),
		},
	}
	option := options.FindOneAndUpdate().SetReturnDocument(options.Before)
	singleResult := d.orderClient.collections[model.DBOrderKey].FindOneAndUpdate(ctx, filter, update, option)
	err = singleResult.Err()
	if err != nil && err != mongo.ErrNoDocuments {
		d.logger.Error("结算佣金更新错误", zap.Error(err))
		return err
	}
	v := &model.Order{}
	err = singleResult.Decode(v)
	if err != nil {
		return err
	}
	if v.Salary != 0 {
		return fmt.Errorf("FindOneAndUpdateCommission 有人在此次更新之前更新了一遍,id: %s", id)
	}
	return nil
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
