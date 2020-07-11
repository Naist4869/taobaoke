package gdbc

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/Naist4869/log"
	"github.com/pkg/errors"
	"go.mongodb.org/mongo-driver/mongo"
	"go.uber.org/zap"
)

const (
	// IndexVersionDelimiter 索引版本号之间的连接符
	IndexVersionDelimiter = "-"
)

// MongoIndex Mongodb数据库所存在的索引数据格式
type MongoIndex struct {
	NS      string
	Name    string
	Version int
}

// Index 索引
type Index struct {
	Name    string           // 索引名
	Version int              // 版本
	Data    mongo.IndexModel // 索引信息
}

func EnsureIndex(collection *mongo.Collection, indexes []Index, logger *log.Logger) error {
	bg := context.Background()
	iterator, err := collection.Indexes().List(bg)
	if err != nil {
		return errors.Wrap(err, "查询索引")
	}
	existIndexes := make(map[string]*MongoIndex)
	defer func() {
		if err = iterator.Close(bg); err != nil {
			logger.Warn("关闭迭代器失败", zap.Error(err))
		}
	}()

	for iterator.Next(bg) {
		index := &MongoIndex{}
		if err = iterator.Decode(index); err != nil {
			return errors.Wrap(err, "解码索引")
		}

		if index.Name != "_id_" {
			fields := strings.Split(index.Name, IndexVersionDelimiter)
			if len(fields) != 2 {
				return fmt.Errorf("索引名称错误，应该为[索引名%s版本],实际为[%s]", IndexVersionDelimiter, index.Name)
			}
			logger.Debug("查询到索引", zap.String("集合名", collection.Name()), zap.Any("索引", index), zap.String("索引元数据", iterator.Current.String()))
			index.Version, _ = strconv.Atoi(fields[1])
			existIndexes[fields[0]] = index

		}
	}
	for _, index := range indexes {
		index.Data.Options = index.Data.Options.SetName(fmt.Sprintf("%s%s%d", index.Name, IndexVersionDelimiter, index.Version)).SetVersion(1)
		if existIndex, exist := existIndexes[index.Name]; exist {

			logger.Debug("索引存在,查看具体版本", zap.String("索引", index.Name), zap.Int("索引版本", existIndex.Version))
			if existIndex.Version != index.Version {
				logger.Debug("版本不同")
				if _, err = collection.Indexes().DropOne(bg, existIndex.Name); err != nil {
					return errors.Wrapf(err, "更新索引[%s.v%d],删除旧版本[%d]", index.Name, index.Version, existIndex.Version)
				}
				if _, err = collection.Indexes().CreateOne(bg, index.Data); err != nil {
					return errors.Wrapf(err, "创建索引[%s.v%d]失败", index.Name, index.Version)
				}
			}
		} else {
			logger.Debug("索引不存在，创建", zap.String("索引", index.Name), zap.Int("索引版本", index.Version))
			if _, err = collection.Indexes().CreateOne(bg, index.Data); err != nil {
				return errors.Wrapf(err, "创建索引[%s.v%d]失败", index.Name, index.Version)
			}
		}
	}
	return nil
}
