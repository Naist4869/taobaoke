//+build id

package dao

import (
	"context"
	"sync"
	"taobaoke/internal/dao/gdbc"

	"github.com/go-kratos/kratos/pkg/conf/paladin"

	"github.com/Naist4869/log"
	"github.com/pkg/errors"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.uber.org/zap"
)

//IDGenerator 对ID生成器的抽象，每次调用Next()生成一个未使用过得ID
type IDGenerator interface {
	Next() (int64, error)
}

const (
	dbIDVersion = 1
	dbIDKey     = "id"
)

type idGenerator struct {
	Current     int64                        `bson:"current"` //当前值，也就是下一次Next()的返回值
	Key         string                       `bson:"key"`     //key
	collections map[string]*mongo.Collection // 数据表map
	Logger      *log.Logger
	lock        *sync.Mutex
}

/*Next 生成一个未使用过得ID
参数:
返回值:
*	int64	int64
*	error	error
*/
func (i *idGenerator) Next() (int64, error) {
	i.lock.Lock()
	defer i.lock.Unlock()

	value := i.Current
	i.Current++
	_, err := i.collections[dbIDKey].UpdateOne(context.Background(), bson.M{"key": i.Key}, bson.M{SET: bson.M{"current": i.Current}})
	if err != nil {
		return 0, errors.Wrap(err, "更新数据库错误")
	}
	return value, nil
}
func NewIDGenerator(client *mongo.Client, logger *log.Logger) (*idGenerator, error) {
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
	ig := &idGenerator{
		collections: make(map[string]*mongo.Collection, 3),
		Logger:      logger.With(zap.String("组件", "id")),
	}
	if err := ig.Init(client, cfg.DB); err != nil {
		return nil, err
	}
	return ig, nil
}

func (i *idGenerator) Keys() map[string]*gdbc.Spec {
	specs := make(map[string]*gdbc.Spec, 1)
	orderSpec, err := gdbc.NewSpec(dbIDVersion, func() interface{} {
		return &IDGenerator{}
	}, func(data interface{}) error {
		return nil
	})
	if err != nil {
		i.Logger.Fatal("构建updater失败:", zap.Error(err))
	}
	specs[dbIDKey] = orderSpec
	return specs
}
func (i *idGenerator) Init(client *mongo.Client, db string) error {

	keys := i.Keys()
	for key, spec := range keys {
		collection := client.Database(db).Collection(key)
		i.collections[key] = collection
		if spec != nil {
			spec.SetCollection(collection)
		}
		if err := gdbc.EnsureIndex(collection, []gdbc.Index{
			{
				Version: 1,
				Name:    "key",
				Data: mongo.IndexModel{
					Keys:    bson.M{"key": 1},
					Options: options.Index().SetUnique(true),
				},
			},
		}, i.Logger); err != nil {
			return errors.Wrap(err, "检查并创建索引失败")
		}
	}
	updater, err := gdbc.NewUpdater(keys, i.Logger)
	if err != nil {
		return errors.Wrap(err, "构建collection升级器失败")
	}
	if err := updater.Update(); err != nil {
		return errors.Wrap(err, "检查并升级collection数据失败")
	}
	return nil
}

func (i *idGenerator) Get(ctx context.Context, key string) (idGenerator, error) {
	addCache := true

}

//
//func (ig *idGenerator) Get(c context.Context, key string) (IDGenerator, error) {
//	ig.lock.Lock()
//	defer ig.lock.Unlock()
//
//	if id, exist := ig.ids[key]; exist {
//		return id, nil
//	}
//	id, err := ig.new(key)
//	if err != nil {
//		return nil, err
//	}
//	ig.ids[key] = id
//	return id, nil
//}
