package dao

import (
	"context"
	"fmt"
	"reflect"
	"strings"

	"github.com/davecgh/go-spew/spew"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

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

/*ZapBsonM 将bson.M转为更加合适的格式，输出到uber.zap
参数:
*	key  	string
*	value	bson.M
返回值:
*	zap.Field	zap.Field
*/
func ZapBsonM(key string, value bson.M) zap.Field {
	return zap.Field{
		Key:    key,
		Type:   zapcore.StringType,
		String: prettyBsonM(value),
	}
}

/*prettyBsonM 将bson.M转为字符串展示
参数:
*	value	bson.M
返回值:
*	string	string
*/
func prettyBsonM(value bson.M) string {
	builder := &strings.Builder{}
	_, err := fmt.Fprint(builder, "{")
	i := 0
	for k, v := range value {
		if i != 0 {
			_, err = fmt.Fprintf(builder, ",")
		}
		if subValue, ok := v.(bson.M); ok {
			_, err = fmt.Fprintf(builder, "%s:%s", k, prettyBsonM(subValue))

		} else {
			if _, ok = v.([]bson.M); ok {
				subValues := v.([]bson.M)
				subValueStrings := make([]string, 0, len(subValues))
				for _, subValue = range subValues {
					subValueStrings = append(subValueStrings, prettyBsonM(subValue))
				}
				_, err = fmt.Fprintf(builder, "%s:%s", k, strings.Join(subValueStrings, ","))
			} else {
				_, err = fmt.Fprintf(builder, "%s:%v", k, v)
			}

		}

		i++
	}
	_, err = fmt.Fprint(builder, "}")
	if err != nil {

	}
	return builder.String()
}

/*CombineBsonM 合并多个Bson.M,如果有重复的字段，那么后出现会覆盖前者，返回的值都是经过复制的
参数:
*	documents	...bson.M
返回值:
*	bson.M	bson.M
*/
func CombineBsonM(documents ...bson.M) bson.M {
	count := 0
	for _, document := range documents {
		count += len(document)
	}

	result := make(bson.M, count)
	for _, document := range documents {
		for k, v := range document {
			result[k] = v
		}
	}
	return result

}

// QuerySpec 查询条件，外部数据->数据使用规则
type QuerySpec map[string]DbSpec

// DbSpec 数据使用规则
type DbSpec struct {
	Field   string   // 针对的字段
	Fields  []string // 针对的多个字段
	Dynamic bool     // 是否动态,结果是条件bson.M，而不是值
	Op      string   // 操作符
	Convert Convert  // 转化规则
}

// Convert 转换规则,将字符串转为对应的类型
type Convert func(data string) (interface{}, error)

/*BuildQuery 构建查询条件
参数:
*	data  	map[string]interface{}	传入的查询数据
*	specs 	QuerySpec			查询规则
*	strict	bool				是否严格，非严格时，将data中未在specs出现的字段转为普通$eq条件
返回值:
*	bson.M	bson.M
*	error 	error
*/
func BuildQuery(data map[string]interface{}, specs QuerySpec, strict bool) (bson.M, error) {
	result := make(map[string]interface{}) // 结果
	mappedSpec, dynamic := remapQuerySpecByField(specs)
	used := make(map[string]bool) // 记录使用过的数据
	var err error
	for field, spec := range mappedSpec { // 遍历规则
		for _, s := range spec { // 单条规则
			if value, exist := data[s.key]; exist { // 如果规则能够适用

				if s.convert != nil {
					if value, err = s.convert(value.(string)); err != nil {
						return nil, fmt.Errorf("转换数据发生问题:\n\t数据:%#v\n\t类型:%T\n\t字段:%s", data[s.key], data[s.key], s.key)
					}
				} else {
				}

				if _, exist = result[field]; !exist {
					result[field] = make(map[string]interface{})
				}
				if len(spec) > 1 {
					result[field].(map[string]interface{})[s.op] = value
					if s.op == "$regex" {
						result[field].(map[string]interface{})["$options"] = "i"
					}
				} else if len(spec) == 1 {
					switch s.op {
					case "$regex":
						result[field] = bson.M{s.op: value, "$options": "i"}
					case "$in", "$nin":
						if value != nil {
							reflectValue := reflect.ValueOf(value)
							if reflectValue.Kind() != reflect.Slice {
								return nil, fmt.Errorf("$in值必须是slice,对应字段[%s]", s.key)
							}
							if !reflectValue.IsNil() {
								result[field] = bson.M{s.op: value}
							}
						}
					default:
						result[field] = bson.M{s.op: value}
					}

				} else {
					delete(result, field)
				}
				used[s.key] = true // 记录已经使用过的数据，用过的数据不能直接删除，可能会多次使用
			}
		}
	}
	for key, convert := range dynamic {
		if value, exist := data[key]; exist { //外部传入值
			if value, err = convert(value.(string)); err != nil {
				return nil, fmt.Errorf("转换数据发生问题:\n\t数据:%#v\n\t类型:%T\n\t字段:%s", data[key], reflect.TypeOf(data[key]).String(), key)
			} else {
				//todo: 选择一种更好的方法， 类型断言失败
				if query, ok := value.(primitive.M); !ok {
					return nil, fmt.Errorf("动态字段[%s]Convert第一个返回值必须是bson.M,实际上是[%s]", key, reflect.TypeOf(value).String())
				} else {
					for key, condition := range query {
						switch key {
						case "$or", "$and":
							//todo: 如果有多个$or或者$and 需要汇总
							result[key] = condition
						default:
							if result[key] == nil {
								result[key] = make(map[string]interface{})
							}
							if fullCondition, ok := condition.(primitive.M); ok {
								for k, v := range fullCondition {
									result[key].(map[string]interface{})[k] = v
								}
							} else {
								spew.Dump(result[key], result, key)
								switch t := result[key].(type) {
								case primitive.M:
									t["$eq"] = condition
								case map[string]interface{}:
									result[key].(map[string]interface{})["$eq"] = condition
								}

							}
						}

					}

				}

			}
			delete(data, key)
		}
	}
	if !strict { // 如果不是严厉规则，将其余的数据加上，判断是否相等
		for k, v := range data {
			if !used[k] {
				result[k] = v
			}

		}
	}
	if len(result) == 0 {
		return nil, nil
	}
	query := clean(bson.M(result))
	return query, nil
}

type querySpec struct {
	key     string
	op      string
	convert Convert
}

func clean(query bson.M) bson.M {
	deleted := make([]string, 0, len(query))
	for k, v := range query {
		if value, ok := v.(map[string]interface{}); ok {
			if len(value) == 0 {
				deleted = append(deleted, k)
			}
		}
	}
	for _, key := range deleted {
		delete(query, key)
	}
	return query
}

func remapQuerySpecByField(spec QuerySpec) (ordinary map[string][]querySpec, dynamic map[string]Convert) {
	ordinary = make(map[string][]querySpec)
	for k, v := range spec {
		if v.Dynamic { //动态字段
			if len(dynamic) == 0 {
				dynamic = make(map[string]Convert, 10)
			}
			dynamic[k] = v.Convert
			continue
		}
		if v.Field != "" {
			if _, exist := ordinary[v.Field]; !exist {
				ordinary[v.Field] = make([]querySpec, 0, 2)
			}
		} else {
			for _, field := range v.Fields {
				if _, exist := ordinary[field]; !exist {
					ordinary[field] = make([]querySpec, 0, 2)
				}
			}
		}
		if v.Field != "" {
			ordinary[v.Field] = append(ordinary[v.Field], querySpec{key: k, op: v.Op, convert: v.Convert})
		} else {
			for _, field := range v.Fields {
				ordinary[field] = append(ordinary[field], querySpec{key: k, op: v.Op, convert: v.Convert})
			}
		}
	}
	return
}
