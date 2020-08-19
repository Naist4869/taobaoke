package dao

import (
	"context"
	"fmt"
	"reflect"
	"strings"

	"github.com/pkg/errors"

	"go.mongodb.org/mongo-driver/bson"

	"github.com/go-kratos/kratos/pkg/conf/paladin"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const (
	AND      = "$and"
	IN       = "$in"
	NIN      = "$nin"
	SET      = "$set"
	ADDTOSET = "$addToSet"
	EQ       = "$eq"
	NE       = "$ne"
	GT       = "$gt"
	GTE      = "$gte"
	LT       = "$lt"
	LTE      = "$lte"
	REGEX    = "$regex"
	OR       = "$or"
	EACH     = "$each"
	PULL     = "$pull"
	INC      = "$inc"
)

type mongoConfig struct {
	Host string
	Port string
	User string
	Pass string
	DB   string
}

func NewMongo() (client *mongo.Client, cf func(), err error) {
	var (
		cfg mongoConfig
		ct  paladin.TOML
	)
	if err = paladin.Get("db.toml").Unmarshal(&ct); err != nil {
		return
	}
	if err = ct.Get("Mongo").UnmarshalTOML(&cfg); err != nil {
		return
	}
	auth := options.Credential{
		AuthMechanism: "SCRAM-SHA-256",
		Username:      cfg.User,
		Password:      cfg.Pass,
		AuthSource:    cfg.DB,
	}
	client, err = mongo.Connect(context.Background(), options.Client().ApplyURI(fmt.Sprintf("mongodb://%s:%s", cfg.Host, cfg.Port)).SetAuth(auth))
	if err != nil {
		return
	}
	cf = func() {
		if err := client.Disconnect(context.Background()); err != nil {
			fmt.Printf("关闭Mongo客户端连接池失败%+v", err)
		}
	}
	return
}

func (d *dao) PingMongo(ctx context.Context) (err error) {
	return
}
func MakeSelect(include, exclude []string) (selection bson.M, err error) {
	if err = validateSelect(include, exclude); err != nil {
		return
	}
	selection = makeSelect(include, exclude)
	return
}

func validateSelect(include, exclude []string) (err error) {
	if len(include) != 0 && len(exclude) != 0 {
		err = errors.New("两个参数必须至少有一个为空")
		return
	}
	return
}

func makeSelect(include, exclude []string) (fields bson.M) {
	fields = make(bson.M, 2)
	for _, field := range include {
		fields[field] = 1
	}
	for _, field := range exclude {
		fields[field] = 0
	}
	return
}
func convertSort(sorts []string) (result bson.D) {
	result = make([]bson.E, 0, len(sorts))
	value := 1
	for _, sort := range sorts {
		if sort[:1] == "-" {
			sort = sort[1:]
			value = -1
		}
		result = append(result, bson.E{
			Key:   sort,
			Value: value,
		})
	}
	return
}
func IsInsertDuplicateError(err error) bool {
	return err != nil && strings.Contains(err.Error(), "E11000 duplicate key error collection")
}
func baseQuery(collection *mongo.Collection, ctx context.Context, query bson.M, sort []string, start, limit int64, include []string, exclude []string, data interface{}, collations ...*options.Collation) (result []interface{}, count int64, err error) {
	var selection bson.M
	if selection, err = MakeSelect(include, exclude); len(selection) == 0 {
		selection = nil
	}
	option := &options.FindOptions{
		Sort:       convertSort(sort),
		Limit:      &limit,
		Skip:       &start,
		Projection: selection,
	}
	switch len(collations) {
	case 0:
	case 1:
		option.Collation = collations[0]
	default:
		err = errors.New("collations参数最多只有1个")
		return
	}
	var cursor *mongo.Cursor
	if cursor, err = collection.Find(ctx, query, option); err != nil {
		err = errors.Wrap(err, "find驱动")
		return
	}
	t := reflect.TypeOf(data)
	defer cursor.Close(ctx)
	for cursor.Next(ctx) {
		element := reflect.New(t)
		if err = cursor.Decode(element.Interface()); err != nil {
			err = errors.Wrap(err, "bson解码")
			return
		}
		result = append(result, element.Elem().Interface())
		count++
	}
	return
}