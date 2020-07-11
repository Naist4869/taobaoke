package dao

import (
	"context"
	"strings"
	"taobaoke/internal/dao/gdbc"
	"taobaoke/internal/model"

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
	Insert(ctx context.Context, o *model.Order) (err error)
	FindOrderByID(ctx context.Context, id string) (order *model.Order, err error)
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
	//if err := oc.Init(client, cfg.DB); err != nil {
	//	return nil, err
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

func (oc *OrderClient) FindOrderByID(ctx context.Context, id string) (order *model.Order, err error) {
	query := bson.M{
		model.IDField: id,
	}
	if orders, _, err := oc.queryOrder(ctx, query, nil, 0, 1, nil, nil); err != nil {
		return nil, errors.New("根据id获取用户错误")
	} else if len(orders) == 0 {

		return nil, NewErrIDNotFound(id)
	} else {
		return orders[0], nil
	}
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
