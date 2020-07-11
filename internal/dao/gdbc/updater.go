package gdbc

import (
	"context"
	"fmt"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"

	"github.com/Naist4869/log"
	"github.com/pkg/errors"
	"go.uber.org/zap"
)

const (
	versionKey = "meta.version"
)

type Updater struct {
	specs  map[string]*Spec
	logger *log.Logger
}

/*NewUpdater 方法说明
参数:
*	specs	map[string]*Spec
返回值:
*	*Updater	*Updater
*	error   	error
*/
func NewUpdater(specs map[string]*Spec, logger *log.Logger) (*Updater, error) {
	return &Updater{
		specs:  specs,
		logger: logger,
	}, nil
}
func (u Updater) Update() error {
	for key, spec := range u.specs {
		if spec == nil {
			continue
		}
		u.logger.Info("初始化升级数据", zap.String("库名", key))
		if hasHigher, err := u.hasHigherVersion(spec.collection, spec.version); err != nil {
			return errors.Wrap(err, "判断有无更高版本数据")
		} else {
			if hasHigher {
				return fmt.Errorf("数据库[%s]有高于版本的数据[%d]", key, spec.version)
			}
			u.logger.Info("没有超过设计版本的数据", zap.String("库名", key))
			if err = u.updateLowerVersion(spec); err != nil {
				return errors.Wrap(err, "升级低版本数据")
			}
		}
	}
	return nil
}

/*hasHigherVersion 判断有无更高版本的数据
参数:
*	collection	*mongo.Collection		数据库
*	version   	int						期望版本
返回值:
*	bool 	bool
*	error	error
*/
func (u Updater) hasHigherVersion(collection *mongo.Collection, version int) (bool, error) {
	count, err := collection.CountDocuments(context.Background(), bson.M{versionKey: bson.M{"$gt": version}})
	return count > 0, err
}

/*updateLowerVersion 升级低版本数据
参数:
*	spec	Spec		数据定义
返回值:
*	error	error
*/
func (u Updater) updateLowerVersion(spec *Spec) error {
	if cursor, err := spec.collection.Find(context.Background(), bson.M{"$or": []bson.M{
		{versionKey: bson.M{"$lt": spec.version}},
		{versionKey: bson.M{"$exists": false}},
	}}); err != nil {
		return errors.Wrap(err, "查询错误")
	} else {
		bg := context.Background()

		defer cursor.Close(bg)
		for cursor.Next(bg) {
			record := spec.generator()
			if err = cursor.Decode(record); err != nil {
				return errors.Wrap(err, "解码数据错误")
			}
			if err = spec.updater(record); err != nil {
				return errors.Wrap(err, "升级版本错误")
			}

		}
	}
	return nil
}
